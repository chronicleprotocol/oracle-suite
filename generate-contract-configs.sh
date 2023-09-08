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
# ./generate-contract-configs.sh <path/to/chronicle-repo>

function findAllConfigs() {
	local _path="$1"
	local _contract="$2"
	local _key="${3:-"contract"}"

	local i
	for i in $(find "$_path" -name '*.json' | sort); do
		jq -c 'select(.'"$_key"' // "" | test("'"$_contract"'","ix"))' "$i"
	done
}

{
	echo "variables {"
	echo -n "contract_map = "
	{
		findAllConfigs "$1/deployments" '^(WatRegistry|Chainlog)$'
		findAllConfigs "$1/deployments" '^TorAddressRegister_Feeds_1$' 'name'
	} | jq -c --argjson m '{"production":"prod","staging":"stage"}' \
	'{($m[.environment]+"-"+.chain+"-"+.contract):.address}' | sort | jq -s 'add'
	echo -n "contracts = "
	{
		findAllConfigs "$1/deployments" '^Scribe(Optimistic)?$' \
		| jq -c \
		--argjson m '{"production":"prod","staging":"stage"}' \
		--argjson p "$(jq -c '.' "$1/relays/params.json")" \
		'{
			env: $m[.environment],
			chain,
			chain_id:.chainId,
			IScribe: (.IScribe != null),
			wat: .IScribe.wat,
			IScribeOptimistic: (.IScribeOptimistic != null),
			address,
			challenge_period:.IScribeOptimistic.opChallengePeriod,
		} + ($p[($m[.environment] + "-" + .chain + "-" + .address)] | del(.wat)) | del(..|nulls)'
		jq -c 'select(.enabled==true) | del(.enabled)' "$1/deployments/medians.jsonl"
	} | sort | jq -s '.'
	echo "}"
} > config-contracts.hcl
