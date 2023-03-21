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

package types

// Transaction represents a transaction.
type Transaction struct {
	Hash             Hash    `json:"hash"`
	BlockHash        Hash    `json:"blockHash"`
	BlockNumber      Number  `json:"blockNumber"`
	TransactionIndex Number  `json:"transactionIndex"`
	From             Address `json:"from"`
	To               Address `json:"to"`
	Gas              Number  `json:"gas"`
	GasPrice         Number  `json:"gasPrice"`
	Input            Bytes   `json:"input"`
	Nonce            Number  `json:"nonce"`
	Value            Number  `json:"value"`
	V                Number  `json:"v"`
	R                Number  `json:"r"`
	S                Number  `json:"s"`
}

// TransactionReceiptType represents transaction receipt.
type TransactionReceiptType struct {
	TransactionHash   Hash     `json:"transactionHash"`
	TransactionIndex  Number   `json:"transactionIndex"`
	BlockHash         Hash     `json:"blockHash"`
	BlockNumber       Number   `json:"blockNumber"`
	From              Address  `json:"from"`
	To                Address  `json:"to"`
	CumulativeGasUsed Number   `json:"cumulativeGasUsed"`
	GasUsed           Number   `json:"gasUsed"`
	ContractAddress   *Address `json:"contractAddress"`
	Logs              []Log    `json:"logs"`
	LogsBloom         Bytes    `json:"logsBloom"`
	Root              *Hash    `json:"root"`
	Status            *Number  `json:"status"`
}
