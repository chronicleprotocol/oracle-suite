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

package types

// Block represents an Ethereum block.
type Block struct {
	Number           Number  `json:"number"`
	Hash             Hash    `json:"hash"`
	ParentHash       Hash    `json:"parentHash"`
	Nonce            Nonce   `json:"nonce"`
	Sha3Uncles       Hash    `json:"sha3Uncles"`
	LogsBloom        Bloom   `json:"logsBloom"`
	TransactionsRoot Hash    `json:"transactionsRoot"`
	StateRoot        Hash    `json:"stateRoot"`
	ReceiptsRoot     Hash    `json:"receiptsRoot"`
	Miner            Address `json:"miner"`
	MixHash          Hash    `json:"mixHash"`
	Difficulty       Number  `json:"difficulty"`
	TotalDifficulty  Number  `json:"totalDifficulty"`
	ExtraData        Bytes   `json:"extraData"`
	Size             Number  `json:"size"`
	GasLimit         Number  `json:"gasLimit"`
	GasUsed          Number  `json:"gasUsed"`
	Timestamp        Number  `json:"timestamp"`
	Uncles           []Hash  `json:"uncles"`
}

// BlockTxHashes represents Ethereum block with transaction hashes.
type BlockTxHashes struct {
	Block
	Transactions []Hash `json:"transactions"`
}

// BlockTxObjects represents Ethereum block with full transactions.
type BlockTxObjects struct {
	Block
	Transactions []Transaction `json:"transactions"`
}
