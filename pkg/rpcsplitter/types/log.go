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

// Log represents a contract log event.
type Log struct {
	Address     Address `json:"address"`
	Topics      []Hash  `json:"topics"`
	Data        Bytes   `json:"data"`
	BlockHash   Hash    `json:"blockHash"`
	BlockNumber Number  `json:"blockNumber"`
	TxHash      Hash    `json:"transactionHash"`
	TxIndex     Number  `json:"transactionIndex"`
	LogIndex    Number  `json:"logIndex"`
	Removed     bool    `json:"removed"`
}

// FilterLogsQuery represents a query to filter logs.
type FilterLogsQuery struct {
	Address   Addresses    `json:"address"`
	FromBlock *BlockNumber `json:"fromBlock,omitempty"`
	ToBlock   *BlockNumber `json:"toBlock,omitempty"`
	Topics    []Hashes     `json:"topics"`
	BlockHash *Hash        `json:"blockhash,omitempty"`
}
