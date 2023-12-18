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

package config

import (
	_ "embed"
)

//go:embed config-contracts.hcl
var Contracts []byte

//go:embed config-defaults.hcl
var Defaults []byte

//go:embed config-ethereum.hcl
var Ethereum []byte

//go:embed config-ghost.hcl
var Ghost []byte

//go:embed config-gofer.hcl
var Gofer []byte

//go:embed config-spectre.hcl
var Spectre []byte

//go:embed config-spire.hcl
var Spire []byte

//go:embed config-transport.hcl
var Transport []byte

//go:embed config-kafka.hcl
var Kafka []byte
