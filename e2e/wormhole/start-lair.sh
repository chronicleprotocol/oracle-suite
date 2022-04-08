#!/usr/bin/env bash

set -e

gitroot=$(git rev-parse --show-toplevel)

wh_dir=$gitroot/e2e/wormhole
tmp_dir=$wh_dir/.tmp
[[ -d $tmp_dir ]] || {
    mkdir -p $tmp_dir
}

[[ -f "$tmp_dir/leeloo.log" ]] || {
    echo "Missing '$tmp_dir/leeloo.log'. Please run start-leeloo.sh first!"
    exit 1
}

# Get the Leeloo listening address and config as the bootstrap address for the
# local libp2p network.
bootstrapaddr=$(cat "$tmp_dir/leeloo.log" | jq -r '. | select(has("listenAddrs")) | select(.msg | test("Listening")) | .listenAddrs | .[] | select(test("127.0.0.1"))' | sed 's/127.0.0.1/0.0.0.0/')
bootstrapaddr_t="\"${bootstrapaddr}\""

cfgfile=$tmp_dir/lair.config.json
cat $wh_dir/template.config.json | sed "s!__BOOTSTRAP_ADDRS__!$bootstrapaddr_t!" | sed "s!__KEYSTORE_PATH__!$keystore!" > $cfgfile 


bootstrapaddr_check=$(cat $cfgfile | jq -r .transport.libp2p.bootstrapAddrs[0])
[[ "$bootstrapaddr" = "$bootstrapaddr_check" ]] || {
    echo "Mismatch: $bootstrapaddr != $bootstrapaddr_check"
    echo "Could not parse the JSONified $tmp_dir/leeloo.log correctly?"
    exit 1
}
echo "Discovered local $bootstrapaddr as bootstrap address. (You should already be running a local version of Leeloo; looks like you are.)"

go run cmd/lair/*.go -c $cfgfile -v debug --log.format json run 2>&1 | tee $tmp_dir/lair.log
