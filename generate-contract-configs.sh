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

	local i
	for i in $(find "$_path" -name '*.json'); do
		jq -c 'select(.contract == "'"$_contract"'")' "$i"
	done
}

function findAll() {
	{
		findAllConfigs "$1" Scribe
		findAllConfigs "$1" ScribeOptimistic
  } | jq -c --argjson m '{"production":"prod","staging":"stage"}' '{
  	env: $m[.environment],
  	chain,
  	chain_id,
  	contract,
  	address,
  	i_scribe: (if .IScribe != null then {
  		wat: .IScribe.wat,
  		bar: .IScribe.bar,
  		decimals: .IScribe.decimals,
  		indexes: ([.IScribe.feeds, .IScribe.feedIndexes] | transpose | map( {(.[0]): .[1]} ) | add),
  	} else null end),
		i_scribe_optimistic: (if .IScribeOptimistic != null then {
  		challenge_period:.IScribeOptimistic.opChallengePeriod,
  	} else null end),
  } | del(..|nulls)'
}

{
	echo "variables {"
	echo -n "  contracts = "
	findAll "$1" | jq -s '.'
	echo "}"
} > config-contracts.hcl

{
	echo "variables {"
	echo -n "  contract_params = "
	findAll "$1" | jq -c '{(.env + "-" + .chain + "-" + .address):{
		optimistic_poke: (if .i_scribe_optimistic != null then {
			spread: 0.5,
			expiration: 3600,
			interval: 10,
		} else null end),
		poke: (if .i_scribe != null then {
			spread: (0.5*2),
			expiration: (3600*2),
			interval: (10*2),
		} else null end),
	}} | del(..|nulls)' | jq -s 'add'
	echo "}"
} > config-contract-params.hcl
