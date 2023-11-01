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

__medians="$(jq -c 'select(.enabled==true) | del(.enabled)' "$1/deployments/medians.jsonl")"
__relays="$(jq -c '.' "config/relays.json")"
__relays_initial="$(jq -c '.' "config/relays-initial.json")"

_CONTRACT_MAP="$({
	findAllConfigs "$1/deployments" '^(WatRegistry|Chainlog)$'
	findAllConfigs "$1/deployments" '^TorAddressRegister_Feeds' 'name'
} | jq -c '{(.environment+"-"+.chain+"-"+.contract):.address}' | sort | jq -s 'add')"

_CONTRACTS="$({
	findAllConfigs "$1/deployments" '^Scribe(Optimistic)?$' \
	| jq -s 'group_by(.environment + " " + .chain + " " + .name | sub("_\\d+$";""))[] | sort_by(.name)[-1]' \
	| jq -c --argjson r "$__relays" --argjson ri "$__relays_initial" \
	'{
		env: .environment,
		chain,
		wat: .IScribe.wat,
		address,
		chain_id:.chainId,
		is_scribe: (.IScribe != null),
		is_scribe_optimistic: (.IScribeOptimistic != null),
		challenge_period:.IScribeOptimistic.opChallengePeriod,
		poke:($ri[(.environment+"-"+.chain+"-"+.IScribe.wat+"-scribe-poke")] // $r[(.environment+"-"+.chain+"-"+.IScribe.wat+"-scribe-poke")] // {}),
		poke_optimistic:($ri[(.environment+"-"+.chain+"-"+.IScribe.wat+"-scribe-poke-optimistic")] // $r[(.environment+"-"+.chain+"-"+.IScribe.wat+"-scribe-poke-optimistic")] // {}),
	} | del(..|nulls) | del(..|select(type=="object" and length==0))'
	jq <<<"$__medians" --argjson r "$__relays" -c '{
		env,
		chain,
		wat,
		address,
		is_median:true,
		poke:$r[(.env+"-"+.chain+"-"+.wat+"-median-poke")],
	} | del(..|nulls) | del(..|select(type=="object" and length==0))'
} | sort | jq -s '.')"

_MODELS="$(go run ./cmd/gofer models | grep '/' | jq -R '.' | sort | jq -s '.')"

{
	echo "variables {"
	echo "contract_map = $_CONTRACT_MAP"
	echo "contracts = $_CONTRACTS"
	echo "models = $_MODELS"
	echo "}"
} > config/config-contracts.hcl

{
	jq <<<"$_CONTRACTS" -c '.[] | select(.is_median) | {
		key: (.env+"-"+.chain+"-"+.wat+"-median-poke"),
		value: (if .poke == null or (.poke | length) == 0 then {spread:1,expiration:86400} else .poke end),
	}'
	jq <<<"$_CONTRACTS" -c '.[] | select(.is_scribe) | {
		key: (.env+"-"+.chain+"-"+.wat+"-scribe-poke"),
		value: (if .poke == null or (.poke | length) == 0 then {spread:1,expiration:32400} else .poke end),
	}'
	jq <<<"$_CONTRACTS" -c '.[] | select(.is_scribe_optimistic) | {
		key: (.env+"-"+.chain+"-"+.wat+"-scribe-poke-optimistic"),
		value: (if .poke_optimistic == null or (.poke_optimistic | length) == 0 then {spread:0.5,expiration:28800} else .poke_optimistic end),
	}'
} | sort | jq -s 'from_entries' > config/relays.json

#TODO go through all contracts and make sure they are in the relay.json config with 0 values, so they can be easily fixed
#todo write an adr

