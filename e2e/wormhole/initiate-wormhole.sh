#!/usr/bin/env bash

set -e

master_domain='0x4f5241434c45532d4d41535445522d3100000000000000000000000000000000'  # ORACLES-MASTER-1

gitroot=$(git rev-parse --show-toplevel)
cd $gitroot
wd=$gitroot/e2e/wormhole

[[ -z "$1" ]] && {
    echo "Please pass the address to the L2 Wormhole Gateway."
    exit 1
}
l2_wormhole_gateway_address="$1"

get_accounts() {
    cat "$wd/.geth/keystore/"* | jq -r .address | awk '{print "0x" $1}'
}

accounts=($(get_accounts))
export ETH_PASSWORD="$wd/empty"
export ETH_KEYSTORE="$wd/.geth/keystore"
export ETH_RPC_URL='http://localhost:8545'
export ETH_FROM="${accounts[0]}"
export ETH_GAS=6000000
(set -x; seth send $l2_wormhole_gateway_address 'initiateWormhole(bytes32,address,uint128)' \
    $master_domain $ETH_FROM 500)
