package starknet

import "github.com/ethereum/go-ethereum/common"

type Block struct {
	BlockHash    *Felt          `json:"block_hash"`
	ParentHash   *Felt          `json:"parent_hash"`
	BlockNumber  uint64         `json:"block_number"`
	Status       string         `json:"status"`
	Sequencer    *Felt          `json:"sequencer"`
	NewRoot      *Felt          `json:"new_root"`
	OldRoot      *Felt          `json:"old_root"`
	AcceptedTime int64          `json:"accepted_time"`
	Transactions []Transactions `json:"transactions"`
}

type Transactions struct {
	TxnHash            *Felt        `json:"txn_hash"`
	ContractAddress    *Felt        `json:"contract_address"`
	EntryPointSelector *Felt        `json:"entry_point_selector"`
	Calldata           []*Felt      `json:"calldata"`
	Status             string       `json:"status"`
	StatusData         string       `json:"status_data"`
	MessagesSent       []*L1Message `json:"messages_sent"`
	L1OriginMessage    *L2Message   `json:"l1_origin_message"`
	Events             []*Event     `json:"events"`
}

type L1Message struct {
	ToAddress *Felt   `json:"to_address"`
	Payload   []*Felt `json:"payload"`
}

type L2Message struct {
	FromAddress common.Address `json:"from_address"`
	Payload     []*Felt        `json:"payload"`
}

type Event struct {
	FromAddress *Felt   `json:"from_address"`
	Keys        []*Felt `json:"keys"`
	Data        []*Felt `json:"data"`
}
