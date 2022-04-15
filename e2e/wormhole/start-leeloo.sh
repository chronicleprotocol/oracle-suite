#!/usr/bin/env bash
# NOTE This must run from top-level of git root

set -e

gitroot=$(git rev-parse --show-toplevel)
cd $gitroot
wh_dir=$gitroot/e2e/wormhole
tmp_dir=$wh_dir/.tmp
mkdir -p $tmp_dir

cfgfile=$tmp_dir/leeloo.config.json
keystore="$wh_dir/.geth/keystore"
cat $wh_dir/template.config.json | sed 's/__BOOTSTRAP_ADDRS__//' | sed "s!__KEYSTORE_PATH__!$keystore!" > $cfgfile 

go run cmd/leeloo/*.go -v debug -c $cfgfile --log.format json run 2>&1 | tee $tmp_dir/leeloo.log
