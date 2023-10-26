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

_CONTRACT_MAP="$({
	findAllConfigs "$1/deployments" '^(WatRegistry|Chainlog)$'
	findAllConfigs "$1/deployments" '^TorAddressRegister_Feeds' 'name'
} | jq -c '{(.environment+"-"+.chain+"-"+.contract):.address}' | sort | jq -s 'add')"

_CONTRACTS="$({
	findAllConfigs "$1/deployments" '^Scribe(Optimistic)?$' \
	| jq -c --argjson p "$__relays" \
	'{
		env: .environment,
		chain,
		wat: .IScribe.wat,
		address,
		chain_id:.chainId,
		IScribe: (.IScribe != null),
		IScribeOptimistic: (.IScribeOptimistic != null),
		challenge_period:.IScribeOptimistic.opChallengePeriod,
		poke:$p[(.environment+"-"+.chain+"-"+.IScribe.wat+"-scribe-poke")],
		poke_optimistic:$p[(.environment+"-"+.chain+"-"+.IScribe.wat+"-scribe-poke-optimistic")],
	} | del(..|nulls)'
	jq <<<"$__medians" --argjson p "$__relays" -c '{
		env,
		chain,
		wat,
		address,
		IMedian:true,
		poke:$p[(.env+"-"+.chain+"-"+.wat+"-median-poke")],
	} | del(..|nulls)'
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
# the commented code might still be useful
#	jq <<<"$__medians" -c '{
#		key: (.env+"-"+.chain+"-"+.wat+"-median-poke"),
#		value: (.poke // {expiration:null,spread:null,interval:null}),
#	}'
#	jq <<<"$__relays" -c 'to_entries | .[] | select(.value.poke != null) | {
#		key: (.key+"-scribe-poke"),
#		value: .value.poke,
#	}'
#	jq <<<"$__relays" -c 'to_entries | .[] | select(.value.optimistic_poke != null) | {
#		key: (.key+"-scribe-poke-optimistic"),
#		value: .value.optimistic_poke,
#	}'
	jq <<<"$_CONTRACTS" -c '.[] | select(.IMedian) | {
		key: (.env+"-"+.chain+"-"+.wat+"-median-poke"),
		value: (if .poke == null or (.poke | length) == 0 then {expiration:null,spread:null,interval:null} else .poke end),
	}'
	jq <<<"$_CONTRACTS" -c '.[] | select(.IScribe) | {
		key: (.env+"-"+.chain+"-"+.wat+"-scribe-poke"),
		value: (if .poke == null or (.poke | length) == 0 then {expiration:null,spread:null,interval:null} else .poke end),
	}'
	jq <<<"$_CONTRACTS" -c '.[] | select(.IScribeOptimistic) | {
		key: (.env+"-"+.chain+"-"+.wat+"-scribe-poke-optimistic"),
		value: (if .poke_optimistic == null or (.poke_optimistic | length) == 0 then {expiration:null,spread:null,interval:null} else .poke_optimistic end),
	}'
} | sort | jq -s 'from_entries' > config/relays.json

#TODO go through all contracts and make sure they are in the relay.json config with 0 values, so they can be easily fixed
#todo write an adr
