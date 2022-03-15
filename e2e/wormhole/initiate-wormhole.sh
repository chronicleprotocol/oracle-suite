#!/usr/bin/env bash

set -e

master_domain='0x4f5241434c45532d4d41535445522d3100000000000000000000000000000000'  # ORACLES-MASTER-1

pushd $(dirname "$0")
wd=$(pwd)

# This directory is created after running init.sh. Bail out if it does not exist.
tmp_dir=$wd/.tmp
[[ -d $tmp_dir ]] || {
    echo "Missing output from environment setup ($tmp_dir/*.json). Please run 'make init' first."
    exit 1
}

pb() {
    popd > /dev/null
}

echovar() {
    eval "echo $1=\${$1}"
}

get_accounts() {
    cat "$wd/.geth/keystore/"* | jq -r .address | awk '{print "0x" $1}'
}

accounts=($(get_accounts))
export ETH_PASSWORD="$wd/empty"
export ETH_KEYSTORE="$wd/.geth/keystore"
export ETH_RPC_URL='http://localhost:8545'
export ETH_FROM="${accounts[0]}"
export ETH_GAS=6000000
l2_wormhole_gateway_address=$(cat $tmp_dir/optimism.json | jq -r .l2_wormhole_gateway_address)
(set -x; seth send $l2_wormhole_gateway_address 'initiateWormhole(bytes32,address,uint128)' \
    $master_domain $ETH_FROM 500)
