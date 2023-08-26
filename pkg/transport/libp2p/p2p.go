//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package libp2p

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	cryptoETH "github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/internal"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

const LoggerTag = "LIBP2P"
const TransportName = "libp2p"

// Mode describes operating mode of the node.
type Mode int

const (
	// ClientMode operates the node as client. ClientMode can publish and read messages
	// and provides peer discovery service for other nodes.
	ClientMode Mode = iota
	// BootstrapMode operates the node as a bootstrap node. BootstrapMode node provides
	// only peer discovery service for other nodes.
	BootstrapMode
)

// Values for the connection limiter:
const minConnections = 100
const maxConnections = 150

// Parameters used to calculate peer scoring and rate limiter values:
const maxBytesPerSecond float64 = 10 * 1024 * 1024 // 10MB/s
const priceUpdateInterval = time.Minute
const minAssetPairs = 10                 // below that, score becomes negative
const maxAssetPairs = 100                // it limits the maximum possible score only, not the number of supported pairs
const maxInvalidMsgsPerHour float64 = 60 // per topic

// Timeout has to be a little longer because signing messages using
// the Ethereum wallet requires more time.
const connectionTimeout = 120 * time.Second

// defaultListenAddrs is the list of default multiaddresses on which node will
// be listening on.
var defaultListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}

// P2P is the wrapper for the Node that implements the transport.Transport
// interface.
type P2P struct {
	id        peer.ID
	node      *internal.Node
	mode      Mode
	topics    map[string]transport.Message
	msgCh     map[string]chan transport.ReceivedMessage
	msgFanOut map[string]*chanutil.FanOut[transport.ReceivedMessage]
}

// Config is the configuration for the P2P transport.
type Config struct {
	// Mode describes in what mode the node should operate.
	Mode Mode

	// Topics is a list of subscribed topics. A value of the map a type of
	// message given as a nil pointer, e.g.: (*Message)(nil).
	Topics map[string]transport.Message

	// PeerPrivKey is a key used for peer identity. If empty, then random key
	// is used. Ignored in bootstrap mode.
	PeerPrivKey crypto.PrivKey

	// MessagePrivKey is a key used to sign messages. If empty, then message
	// are signed with the same key which is used for peer identity. Ignored
	// in bootstrap mode.
	MessagePrivKey crypto.PrivKey

	// ListenAddrs is a list of multiaddresses on which this node will be
	// listening on. If empty, the localhost, and a random port will be used.
	ListenAddrs []string

	// BootstrapAddrs is a list multiaddresses of initial peers to connect to.
	// This option is ignored when discovery is disabled.
	BootstrapAddrs []string

	// DirectPeersAddrs is a list multiaddresses of peers to which messages
	// will be sent directly. This option has to be configured symmetrically
	// at both ends.
	DirectPeersAddrs []string

	// BlockedAddrs is a list of multiaddresses to which connection will be
	// blocked. If an address on that list contains an IP and a peer ID, both
	// will be blocked separately.
	BlockedAddrs []string

	// AuthorAllowlist is a list of allowed message authors. Only messages from
	// these addresses will be accepted.
	AuthorAllowlist []types.Address

	// Discovery indicates whenever peer discovery should be enabled.
	// If discovery is disabled, then DirectPeersAddrs must be used
	// to connect to the network. Always enabled in bootstrap mode.
	Discovery bool

	// Signer used to verify price messages. Ignored in bootstrap mode.
	Signer wallet.Key

	// Logger is a custom logger instance. If not provided then null
	// logger is used.
	Logger log.Logger

	// Application info:
	AppName    string
	AppVersion string
}

// New returns a new instance of a transport, implemented with
// the libp2p library.
//
//nolint:gocyclo,funlen
func New(cfg Config) (*P2P, error) {
	var err error

	if len(cfg.ListenAddrs) == 0 {
		cfg.ListenAddrs = defaultListenAddrs
	}
	if cfg.PeerPrivKey == nil {
		cfg.PeerPrivKey, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader) //nolint:gomnd
		if err != nil {
			return nil, fmt.Errorf("P2P transport error, unable to generate a random private key: %w", err)
		}
	}
	if cfg.MessagePrivKey == nil {
		cfg.MessagePrivKey = cfg.PeerPrivKey
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}

	listenAddrs, err := strsToMaddrs(cfg.ListenAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to parse listenAddrs: %w", err)
	}
	bootstrapAddrs, err := strsToMaddrs(cfg.BootstrapAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to parse bootstrapAddrs: %w", err)
	}
	directPeersAddrs, err := strsToMaddrs(cfg.DirectPeersAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to parse directPeersAddrs: %w", err)
	}
	blockedAddrs, err := strsToMaddrs(cfg.BlockedAddrs)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error: unable to parse blockedAddrs: %w", err)
	}

	logger := cfg.Logger.WithField("tag", LoggerTag)
	opts := []internal.Options{
		internal.DialTimeout(connectionTimeout),
		internal.Logger(logger),
		internal.ConnectionLogger(),
		internal.PeerLogger(),
		internal.UserAgent(fmt.Sprintf("%s/%s", cfg.AppName, cfg.AppVersion)),
		internal.ListenAddrs(listenAddrs),
		internal.DirectPeers(directPeersAddrs),
		internal.Denylist(blockedAddrs),
		internal.ConnectionLimit(
			minConnections,
			maxConnections,
			5*time.Minute,
		),
		internal.Monitor(),
	}
	if cfg.PeerPrivKey != nil {
		opts = append(opts, internal.PeerPrivKey(cfg.PeerPrivKey))
	}
	if cfg.Signer != nil && !sliceutil.Contains[types.Address](cfg.AuthorAllowlist, cfg.Signer.Address()) {
		cfg.AuthorAllowlist = append(cfg.AuthorAllowlist, cfg.Signer.Address())
	}
	for _, addr := range cfg.AuthorAllowlist {
		logger.
			WithField("address", addr.String()).
			Info("Feed")
	}
	for _, addr := range cfg.BootstrapAddrs {
		logger.
			WithField("address", addr).
			Info("Bootstrap")
	}

	switch cfg.Mode {
	case ClientMode:
		priceTopicScoreParams, err := calculatePriceTopicScoreParams(cfg)
		if err != nil {
			return nil, fmt.Errorf("P2P transport error: invalid price topic scoring parameters: %w", err)
		}
		opts = append(opts,
			internal.MessageLogger(),
			internal.RateLimiter(rateLimiterConfig(cfg)),
			internal.PeerScoring(peerScoreParams, thresholds, func(topic string) *pubsub.TopicScoreParams {
				if topic == messages.PriceV0MessageName || topic == messages.PriceV1MessageName { //nolint:staticcheck
					return priceTopicScoreParams
				}
				return nil
			}),
			messageValidator(cfg.Topics, logger), // must be registered before any other validator
			feedValidator(cfg.AuthorAllowlist, logger),
			priceValidator(logger, cryptoETH.ECRecoverer),
		)
		if cfg.MessagePrivKey != nil {
			opts = append(opts, internal.MessagePrivKey(cfg.MessagePrivKey))
		}
		if cfg.Discovery {
			opts = append(opts, internal.Discovery(bootstrapAddrs))
		}
	case BootstrapMode:
		opts = append(opts,
			internal.DisablePubSub(),
			internal.Discovery(bootstrapAddrs),
		)
	default:
		return nil, fmt.Errorf("P2P transport error: invalid mode: %d", cfg.Mode)
	}

	n, err := internal.NewNode(opts...)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to initialize node: %w", err)
	}

	id, err := peer.IDFromPrivateKey(cfg.MessagePrivKey)
	if err != nil {
		return nil, fmt.Errorf("P2P transport error, unable to get public ID from private key: %w", err)
	}

	return &P2P{
		id:        id,
		node:      n,
		mode:      cfg.Mode,
		topics:    cfg.Topics,
		msgCh:     map[string]chan transport.ReceivedMessage{},
		msgFanOut: map[string]*chanutil.FanOut[transport.ReceivedMessage]{},
	}, nil
}

// Start implements the transport.Transport interface.
func (p *P2P) Start(ctx context.Context) error {
	if err := p.node.Start(ctx); err != nil {
		return fmt.Errorf("P2P transport error, unable to start node: %w", err)
	}
	if p.mode == ClientMode {
		for topic := range p.topics {
			msgCh := make(chan transport.ReceivedMessage)
			p.msgCh[topic] = msgCh
			p.msgFanOut[topic] = chanutil.NewFanOut(msgCh)
			if err := p.subscribe(topic); err != nil {
				return err
			}
		}
	}
	return nil
}

// Wait implements the transport.Transport interface.
func (p *P2P) Wait() <-chan error {
	return p.node.Wait()
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(topic string, message transport.Message) error {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return fmt.Errorf("P2P transport error, unable to get subscription for %s topic: %w", topic, err)
	}
	data, err := message.MarshallBinary()
	if err != nil {
		return fmt.Errorf("P2P transport error, unable to marshall message: %w", err)
	}
	return sub.Publish(data)
}

// Messages implements the transport.Transport interface.
func (p *P2P) Messages(topic string) <-chan transport.ReceivedMessage {
	if fo, ok := p.msgFanOut[topic]; ok {
		return fo.Chan()
	}
	return nil
}

func (p *P2P) subscribe(topic string) error {
	sub, err := p.node.Subscribe(topic)
	if err != nil {
		return fmt.Errorf("P2P transport error, unable to subscribe to topic %s: %w", topic, err)
	}
	go p.messagesLoop(topic, sub)
	return nil
}

func (p *P2P) messagesLoop(topic string, sub *internal.Subscription) {
	for {
		nodeMsg, ok := <-sub.Next()
		if !ok {
			return
		}
		if msg, ok := nodeMsg.ValidatorData.(transport.Message); ok {
			p.msgCh[topic] <- transport.ReceivedMessage{
				Message: msg,
				Author:  ethkey.PeerIDToAddress(nodeMsg.GetFrom()).Bytes(),
				Data:    nodeMsg,
				Meta: transport.Meta{
					Transport:            TransportName,
					Topic:                topic,
					MessageID:            hex.EncodeToString([]byte(nodeMsg.ID)),
					PeerID:               nodeMsg.GetFrom().String(),
					PeerAddr:             ethkey.PeerIDToAddress(nodeMsg.GetFrom()).String(),
					ReceivedFromPeerID:   nodeMsg.ReceivedFrom.String(),
					ReceivedFromPeerAddr: ethkey.PeerIDToAddress(nodeMsg.ReceivedFrom).String(),
				},
			}
		}
	}
}

// strsToMaddrs converts multiaddresses given as strings to a
// list of multiaddr.Multiaddr.
func strsToMaddrs(addrs []string) ([]core.Multiaddr, error) {
	var maddrs []core.Multiaddr
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return nil, err
		}
		maddrs = append(maddrs, maddr)
	}
	return maddrs, nil
}

func rateLimiterConfig(cfg Config) internal.RateLimiterConfig {
	bytesPerSecond := maxBytesPerSecond
	burstSize := maxBytesPerSecond * priceUpdateInterval.Seconds()
	return internal.RateLimiterConfig{
		BytesPerSecond:      maxBytesPerSecond / float64(len(cfg.AuthorAllowlist)),
		BurstSize:           int(burstSize / float64(len(cfg.AuthorAllowlist))),
		RelayBytesPerSecond: bytesPerSecond,
		RelayBurstSize:      int(burstSize),
	}
}
