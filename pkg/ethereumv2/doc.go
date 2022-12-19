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

// Package ethereumv2 provides interfaces to interact with the Ethereum
// compatible blockchains. It is intended to replace the ethereum package.
// After the ethereum package is removed, this package will be renamed to
// ethereum.
//
// For now, there is only one implementation of the rpcclient, but in the
// future, there may be a need for specific implementations for different L2s.
// The same may be true for rpcsplitter.
package ethereumv2
