//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package starknet

type Block struct {
	BlockHash           *Felt                 `json:"block_hash"`
	ParentBlockHash     *Felt                 `json:"parent_block_hash"`
	BlockNumber         uint64                `json:"block_number"`
	StateRoot           string                `json:"state_root"`
	Status              string                `json:"status"`
	GasPrice            string                `json:"gas_price"`
	Transactions        []*Transaction        `json:"Transactions"`
	Timestamp           int64                 `json:"timestamp"`
	SequencerAddress    string                `json:"sequencer_address"`
	TransactionReceipts []*TransactionReceipt `json:"transaction_receipts"`
}

type Transaction struct {
	ContractAddress     *Felt   `json:"contract_address"`
	ContractAddressSalt *Felt   `json:"contract_address_salt,omitempty"`
	ClassHash           *Felt   `json:"class_hash,omitempty"`
	ConstructorCalldata []*Felt `json:"constructor_calldata,omitempty"`
	TransactionHash     *Felt   `json:"transaction_hash"`
	Type                string  `json:"type"`
	EntryPointSelector  *Felt   `json:"entry_point_selector,omitempty"`
	EntryPointType      string  `json:"entry_point_type,omitempty"`
	Calldata            []*Felt `json:"calldata,omitempty"`
	MaxFee              *Felt   `json:"max_fee,omitempty"`
}

type BuiltinInstanceCounter struct {
	PedersenBuiltin   int `json:"pedersen_builtin"`
	RangeCheckBuiltin int `json:"range_check_builtin"`
	OutputBuiltin     int `json:"output_builtin"`
	EcdsaBuiltin      int `json:"ecdsa_builtin"`
	BitwiseBuiltin    int `json:"bitwise_builtin"`
	EcOpBuiltin       int `json:"ec_op_builtin"`
}

type ExecutionResource struct {
	NSteps                 int                     `json:"n_steps"`
	BuiltinInstanceCounter *BuiltinInstanceCounter `json:"builtin_instance_counter"`
	NMemoryHoles           int                     `json:"n_memory_holes"`
}

type L1ToL2Message struct {
	FromAddress *Felt   `json:"from_address"`
	ToAddress   *Felt   `json:"to_address"`
	Selector    *Felt   `json:"selector"`
	Payload     []*Felt `json:"payload"`
	Nonce       *Felt   `json:"nonce"`
}

type Event struct {
	FromAddress *Felt   `json:"from_address"`
	Keys        []*Felt `json:"keys"`
	Data        []*Felt `json:"data"`
}

type TransactionReceipt struct {
	TransactionIndex      int                `json:"transaction_index"`
	TransactionHash       *Felt              `json:"transaction_hash"`
	L2ToL1Messages        []*L1ToL2Message   `json:"l2_to_l1_messages"`
	Events                []*Event           `json:"events"`
	ExecutionResources    *ExecutionResource `json:"execution_resources"`
	ActualFee             string             `json:"actual_fee"`
	L1ToL2ConsumedMessage *L1ToL2Message     `json:"l1_to_l2_consumed_message,omitempty"`
}
