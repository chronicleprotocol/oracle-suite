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

package ssb

import (
	"encoding/json"
)

type FeedAssetPrice struct {
	Type           string          `json:"type"`
	Version        string          `json:"version"`
	Price          float64         `json:"price"`
	PriceHex       string          `json:"priceHex"`
	Time           int             `json:"time"`
	TimeHex        string          `json:"timeHex"`
	Hash           string          `json:"hash"`
	Signature      string          `json:"signature"`
	Sources        json.RawMessage `json:"sources"`
	StarkSignature json.RawMessage `json:"starkSignature"`
}
