package ethereum

import (
	"context"
	"math/big"
	"time"

	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const WormholeEventType = "wormhole"

const retryAttempts = 3
const retryInterval = 5 * time.Second

// WormholeTopic0 is Keccak256("WormholeInitialized((bytes32,bytes32,address,address,uint128,uint80,uint48))")
var wormholeTopic0 = ethereum.HexToHash("0x0162851814eed5360ac17d1b3d942e6619fa2d803de71e7159ed9bebf724072a")

type EthClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q geth.FilterQuery) ([]types.Log, error)
}

// WormholeListener listens to particular logs on Ethereum compatible blockchain and
// converts them into event messages.
type WormholeListener struct {
	lastBlock uint64               // last block from which logs were pulled
	msgCh     chan *messages.Event // list of channels to which messages will be sent

	// Configuration fields:
	client        EthClient          // Ethereum client
	addresses     []ethereum.Address // Contract addresses.
	interval      time.Duration      // time interval between pulling logs from Ethereum client
	confirmations int                // number of required block confirmations
}

// WormholeListenerConfig contains a configuration options for NewWormholeListener.
type WormholeListenerConfig struct {
	Client        EthClient
	Addresses     []ethereum.Address
	Interval      time.Duration
	Confirmations int
}

// NewWormholeListener returns a new instance of the WormholeListener struct.
func NewWormholeListener(cfg WormholeListenerConfig) *WormholeListener {
	return &WormholeListener{
		msgCh:         make(chan *messages.Event),
		client:        cfg.Client,
		addresses:     cfg.Addresses,
		interval:      cfg.Interval,
		confirmations: cfg.Confirmations,
	}
}

// Events implements the eventobserver.Listener interface.
func (l *WormholeListener) Events() chan *messages.Event {
	return l.msgCh
}

// Start implements the eventobserver.Listener interface.
func (l *WormholeListener) Start(ctx context.Context) error {
	if l.interval == 0 {
		return nil
	}
	go l.listenerLoop(ctx)
	return nil
}

func (l *WormholeListener) listenerLoop(ctx context.Context) {
	t := time.NewTicker(l.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			curr, err := l.client.BlockNumber(ctx)
			if err != nil {
				continue
			}
			from, to := findBlockRange(l.lastBlock, curr, l.confirmations)
			for _, ll := range fetchLogs(ctx, l.client, l.addresses, from, to) {
				msg, err := logToMessage(ll)
				if err != nil {
					// TODO: logging
					continue
				}
				l.msgCh <- msg
			}
			l.lastBlock = curr
		}
	}
}

// findBlockRange returns the range of blocks to examine based on the last
// block number examined, the current block number, and the required number
// of block confirmations.
func findBlockRange(last, curr uint64, conf int) (*big.Int, *big.Int) {
	lastBig := new(big.Int).SetUint64(last)
	currBig := new(big.Int).SetUint64(curr)
	from := new(big.Int).Add(lastBig, big.NewInt(1))
	to := new(big.Int).Sub(currBig, new(big.Int).SetInt64(int64(conf)))
	if lastBig.Cmp(big.NewInt(0)) == 0 || from.Cmp(to) > 0 {
		from = to
	}
	return from, to
}

func fetchLogs(ctx context.Context, client EthClient, addrs []ethereum.Address, from, to *big.Int) []types.Log {
	var err error
	var all []types.Log
	for _, addr := range addrs {
		var logs []types.Log
		if err = retry(func() error {
			logs, err = client.FilterLogs(ctx, geth.FilterQuery{
				FromBlock: from,
				ToBlock:   to,
				Addresses: []common.Address{addr},
				Topics:    [][]common.Hash{{wormholeTopic0}},
			})
			return err
		}); err != nil {
			continue
		}
		all = append(all, logs...)
	}
	return all
}

// logToMessage creates a transport message of "event" type from
// given Ethereum log.
func logToMessage(log types.Log) (*messages.Event, error) {
	guid, err := unpackWormholeGUID(log.Data)
	if err != nil {
		return nil, err
	}
	hash, err := guid.hash()
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{
		"hash": hash.Bytes(),
		"data": log.Data,
	}
	return &messages.Event{
		Date:       time.Now(),
		Type:       WormholeEventType,
		ID:         append(log.TxHash.Bytes(), big.NewInt(int64(log.Index)).Bytes()...),
		Group:      log.TxHash.Bytes(),
		Data:       data,
		Signatures: map[string][]byte{},
	}, nil
}

// retry runs the f function until it returns nil. Maximum number of retries
// and delay between them are defined in the retryAttempts and retryInterval
// constants.
func retry(f func() error) (err error) {
	for i := 0; i < retryAttempts; i++ {
		if i > 0 {
			time.Sleep(retryInterval)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}

type wormholeGUID struct {
	sourceDomain common.Hash
	targetDomain common.Hash
	receiver     common.Address
	operator     common.Address
	amount       *big.Int
	nonce        *big.Int
	timestamp    int64
}

func (g *wormholeGUID) hash() (common.Hash, error) {
	b, err := packWormholeGUID(g)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(b), nil
}

func packWormholeGUID(g *wormholeGUID) ([]byte, error) {
	b, err := abiWormholeGUID.Pack(
		g.sourceDomain,
		g.targetDomain,
		g.receiver,
		g.operator,
		g.amount,
		g.nonce,
		big.NewInt(g.timestamp),
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func unpackWormholeGUID(data []byte) (*wormholeGUID, error) {
	u, err := abiWormholeGUID.Unpack(data)
	if err != nil {
		return nil, err
	}
	return &wormholeGUID{
		sourceDomain: bytes32ToHash(u[0].([32]uint8)),
		targetDomain: bytes32ToHash(u[1].([32]uint8)),
		receiver:     u[2].(common.Address),
		operator:     u[3].(common.Address),
		amount:       u[4].(*big.Int),
		nonce:        u[5].(*big.Int),
		timestamp:    u[6].(*big.Int).Int64(),
	}, nil
}

func bytes32ToHash(b [32]uint8) common.Hash {
	return common.BytesToHash(b[:])
}

var abiWormholeGUID abi.Arguments

func init() {
	bytes32, _ := abi.NewType("bytes32", "", nil)
	address, _ := abi.NewType("address", "", nil)
	uint128, _ := abi.NewType("uint128", "", nil)
	uint80, _ := abi.NewType("uint128", "", nil)
	uint48, _ := abi.NewType("uint48", "", nil)
	abiWormholeGUID = abi.Arguments{
		{Type: bytes32}, // sourceDomain
		{Type: bytes32}, // targetDomain
		{Type: address}, // receiver
		{Type: address}, // operator
		{Type: uint128}, // amount
		{Type: uint80},  // nonce
		{Type: uint48},  // timestamp
	}
}
