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
	findAllConfigs "$1" '^(WatRegistry|Chainlog)$'
	findAllConfigs "$1" '^TorAddressRegister_Feeds_1$' 'name'
	} | jq -c --argjson m '{"production":"prod","staging":"stage"}' '{($m[.environment]+"-"+.chain+"-"+.contract):.address}' \
	| sort | jq -s 'add'

	echo -n "contracts = "
	{
		findAllConfigs "$1" '^Scribe(Optimistic)?$' \
		| jq -c --argjson m '{"production":"prod","staging":"stage"}' '{
			env: $m[.environment],
			chain,
			chain_id,
			address,
			IScribe: (.IScribe != null),
			wat: .IScribe.wat,
			IScribeOptimistic: (.IScribeOptimistic != null),
			challenge_period:.IScribeOptimistic.opChallengePeriod,
		} | del(..|nulls)'

		jq -c --argjson m '{"eth":"prod","arb1":"prod","oeth":"prod","gor":"stage","arb-goerli":"stage","ogor":"stage"}' 'to_entries[] | {chain: .key, value: .value|to_entries[]} | {
			env: $m[.chain],
			chain,
			IMedian:true,
			wat:.value.key,
		} + .value.value | {env,chain,IMedian,wat,address:.oracle,poke:{expiration:.oracleExpiration,spread:.oracleSpread}}' "$1/medians.json"
	} | grep -v 'MANA/USD' | sort | jq -s '.'

	echo "}"
} > config-contracts.hcl

#{
#	echo "variables {"
#	echo -n "  contract_params = "
#	findAll "$1" | jq -c '{(.env + "-" + .chain + "-" + .address):{
#		optimistic_poke: (if .i_scribe_optimistic != null then {
#			spread: 0.5,
#			expiration: 28800,
#			interval: 120,
#		} else null end),
#		poke: (if .i_scribe != null then {
#			spread: 1,
#			expiration: 32400,
#			interval: 120,
#		} else null end),
#	}} | del(..|nulls)' | jq -s 'add'
#	echo "}"
#} > config-contract-params.hcl
