#!/usr/bin/env bash

set -e
echo "======== TEST: Wormhole attestation ========"

gitroot=$(git rev-parse --show-toplevel)
cd $gitroot

wd=$gitroot/e2e/wormhole
tmp_dir=$wd/.tmp
mkdir -p $tmp_dir

get_accounts() {
    cat "$wd/.geth/keystore/"* | jq -r .address | awk '{print "0x" $1}'
}

slep() {
    echo "Sleeping for $1..."
    sleep $1
}

accounts=($(get_accounts))
export ETH_PASSWORD="$wd/empty"
export ETH_KEYSTORE="$wd/.geth/keystore"
export ETH_RPC_URL='http://localhost:8545'
export ETH_FROM="${accounts[0]}"
export ETH_GAS=6000000
printenv | grep -E '^ETH_'

echo "Building Leeloo + Lair..."
go build -o $tmp_dir/leeloo cmd/leeloo/*.go
go build -o $tmp_dir/lair cmd/lair/*.go

echo 'Starting leeloo...'
leeloo_cfgfile=$tmp_dir/leeloo.config.json
keystore="$wd/.geth/keystore"
cat $wd/template.config.json | sed 's/__BOOTSTRAP_ADDRS__//' | sed "s!__KEYSTORE_PATH__!$keystore!" > $leeloo_cfgfile
($tmp_dir/leeloo -v debug -c $leeloo_cfgfile --log.format json run 2>&1 | tee $tmp_dir/leeloo.log)&
slep 10

bootstrapaddr=$(cat "$tmp_dir/leeloo.log" | jq -r '. | select(has("listenAddrs")) | select(.msg | test("Listening")) | .listenAddrs | .[] | select(test("127.0.0.1"))' | sed 's/127.0.0.1/0.0.0.0/')
bootstrapaddr_t="\"${bootstrapaddr}\""

lair_cfgfile=$tmp_dir/lair.config.json
cat $wd/template.config.json | sed "s!__BOOTSTRAP_ADDRS__!$bootstrapaddr_t!" | sed "s!__KEYSTORE_PATH__!$keystore!" > $lair_cfgfile

bootstrapaddr_check=$(cat $lair_cfgfile | jq -r .transport.libp2p.bootstrapAddrs[0])
[[ "$bootstrapaddr" = "$bootstrapaddr_check" ]] || {
    echo "Mismatch: $bootstrapaddr != $bootstrapaddr_check"
    echo "Could not parse the JSONified $tmp_dir/leeloo.log correctly?"
    exit 1
}

echo 'Starting lair...'
($tmp_dir/lair -c $lair_cfgfile -v debug --log.format json run 2>&1 | tee $tmp_dir/lair.log)&
slep 30

endit() {
    pkill leeloo
    pkill lair
}

echo "Initiating the wormhole..."
$wd/initiate-wormhole.sh $(cat $wd/auxiliary/optimism.json | jq -r .l2_wormhole_gateway_address);
slep 60

idx="$(cat $tmp_dir/leeloo.log | jq -r '. | select( has("type") == true ) | select(.type | test("wormhole")) | .index' | tail -n 1)"
echo "Querying lair for attestation for $idx..."
if [ -z "$LAIR_HOST" ]; then
    lair_host='http://0.0.0.0:8208'
else
    lair_host="$LAIR_HOST"
fi
res=$(curl -v -H 'Content-Type: application/json' "$lair_host?type=wormhole&index=$idx")
[[ "$?" = "0" ]] || {
    echo "NOT OK - curl failed."
    endit
    exit 1
}
[[ $(echo "$res" | jq '.[] | has("data") == true') = "true" ]] || {
    echo "NOT OK - response is $res"
    endit
    exit 1
}

endit

echo "$res" | jq .
echo "OK :)"
