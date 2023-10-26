#!/usr/bin/env bash
#  Copyright (C) 2021-2023 Chronicle Labs, Inc.
#
#  This program is free software: you can redistribute it and/or modify
#  it under the terms of the GNU Affero General Public License as
#  published by the Free Software Foundation, either version 3 of the
#  License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU Affero General Public License for more details.
#
#  You should have received a copy of the GNU Affero General Public License
#  along with this program.  If not, see <http://www.gnu.org/licenses/>.

set -euo pipefail

# Usage:
# ./generate-contract-configs.sh <path/to/chronicle-repo> [<path/to/musig-repo>]

function findAllConfigs() {
	local _path="$1"
	local _contract="$2"
	local _key="${3:-"contract"}"

	local i
	for i in $(find "$_path" -name '*.json' | sort); do
		jq -c 'select(.'"$_key"' // "" | test("'"$_contract"'","ix"))' "$i"
	done
}

_MODELS="$(go run ./cmd/gofer models | grep '/' | jq -R '.' | sort | jq -s '.')"

{
	echo "variables {"

	echo -n "contract_map = "
	{
		findAllConfigs "$1/deployments" '^(WatRegistry|Chainlog)$'
		findAllConfigs "$1/deployments" '^TorAddressRegister_Feeds_1$' 'name'
	} | jq -c '{(.environment+"-"+.chain+"-"+.contract):.address}' | sort | jq -s 'add'

	echo -n "contracts = "
	{
		findAllConfigs "$1/deployments" '^Scribe(Optimistic)?$' \
		| jq -c \
		--argjson p "$(jq -c '.' "config/relays.json")" \
		'{
			env: .environment,
			chain,
			chain_id:.chainId,
			IScribe: (.IScribe != null),
			wat: .IScribe.wat,
			IScribeOptimistic: (.IScribeOptimistic != null),
			address,
			challenge_period:.IScribeOptimistic.opChallengePeriod,
		} + ($p[(.environment + "-" + .chain + "-" + .IScribe.wat)]) | del(..|nulls)'
		jq -c 'select(.enabled==true) | del(.enabled)' "$1/deployments/medians.jsonl"
	} | sort | jq -s '.'

	echo "models = $_MODELS"

	echo "}"
} > config/config-contracts.hcl
