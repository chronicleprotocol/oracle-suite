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

package suite

import (
	_ "embed"
)

// Version that can be used by commands.
// It is set by the linker during build.
var Version = "unknown"

//go:embed config-contracts.hcl
var ConfigContracts []byte

//go:embed config-defaults.hcl
var ConfigDefaults []byte

//go:embed config-ethereum.hcl
var ConfigEthereum []byte

//go:embed config-ghost.hcl
var ConfigGhost []byte

//go:embed config-gofer.hcl
var ConfigGofer []byte

//go:embed config-spectre.hcl
var ConfigSpectre []byte

//go:embed config-spire.hcl
var ConfigSpire []byte

//go:embed config-transport.hcl
var ConfigTransport []byte
