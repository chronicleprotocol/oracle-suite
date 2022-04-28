#!/usr/bin/env bash

set -e

pushd $(dirname "$0")
wd=$(pwd)

# Config
optimism_dir=$wd/optimism
dss_wormhole_dir=$wd/dss-wormhole
dss_dir=$wd/dss
pause_dir=$wd/pause_proxy
esm_dir=$wd/esm
optimism_bridge_dir=$wd/optimism-dai-bridge-artifacts
tmp_dir=$wd/.tmp
l1endpoint='http://localhost:9545'
l2endpoint='http://localhost:8545'
ilk='0x57482d4f5241434c45532d544553542d31000000000000000000000000000000'            # WH-ORACLES-TEST-1
master_domain='0x4f5241434c45532d4d41535445522d3100000000000000000000000000000000'  # ORACLES-MASTER-1
slave_domain='0x4f5241434c45532d534c4156452d4f5054494d49534d2d310000000000000000'   # ORACLES-SLAVE-OPTIMISM-1
l2_xdomain_messenger='0x4200000000000000000000000000000000000007'
esm_min=10000000000000000000000                                                     # 10_000 * (10^18)
line=10000000000000000000000000

# TODO This was necessary to get images to run on github actions, at least when
# building on MacOS/M1. Prefer to run arm64, but this arch is working atm...
export DOCKER_DEFAULT_PLATFORM=linux/amd64

pd() {
    pushd "$1" > /dev/null
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

json_to_env() {
    for e in $(echo "$1" | jq -r 'to_entries|map("\"\(.key)=\(.value|tostring)\"")|.[]'); do
        echo "$e"
        eval "export $e"
    done
}

ethutil_init() {
    docker build -f $wd/auxiliary/ethutil/compute-contract-address/Dockerfile \
        $wd/auxiliary/ethutil/compute-contract-address -t oracle-suite-compute-address
    docker build -f $wd/auxiliary/ethutil/dump-events/Dockerfile \
        $wd/auxiliary/ethutil/dump-events -t oracle-suite-dump-events
}

ethutil() {
    case $1 in
      "compute-contract-address") docker run oracle-suite-compute-address $@;;
      "dump-events-l2") shift && docker run --network=ops_default \
          --volume "$optimism_bridge_dir/out":/artifacts \
          oracle-suite-dump-events \
          dump-events \
          http://ops-l2geth-1:8545 $@;;
      *) echo "wumpwump";;
    esac

}

eth_env_l1() {
    export ETH_PASSWORD="$wd/empty"
    export ETH_KEYSTORE="$wd/.geth/keystore"
    export ETH_RPC_URL=$l1endpoint
    # NOTE predeployed addresses on Optimism l1/l2 testnets:
    export ETH_FROM=0xa0ee7a142d267c1f36714e4a8f75612f20a79720
    export ETH_GAS=6000000
    printenv | ack '^ETH_' | awk '{print "export " $0}'
}

eth_env_l2() {
    eth_env_l1 > /dev/null
    export ETH_RPC_URL=$l2endpoint
    printenv | ack '^ETH_' | awk '{print "export " $0}'
}

l1_l2_init() {
    echo 'Build Optimism L2/L1 env...'
    echo '  https://community.optimism.io/docs/developers/build/dev-node'
    echo '  NOTE be sure Docker has enough memory!'
    git clone https://github.com/ethereum-optimism/optimism.git $optimism_dir || true
    pd $optimism_dir/ops
    git pull origin develop
    # TODO(jamesr) Since the Optimism team switched to alpine a dep in the
    # deployer image seems to cause a crash, at least in MacOS. Locking into
    # this commit for now.
    git reset --hard 032731b50f860c91e21d36beae9b6fd21bccaf5f
    make build
    pb
}

start_l1_l2() {
    pd $optimism_dir/ops
    make up && scripts/wait-for-sequencer.sh && echo "System is ready to accept transactions"

    retries=10
    for i in `seq $retries`; do
        l1_xdomain_messenger=$(curl -s http://localhost:8080/addresses.json | jq -r .OVM_L1CrossDomainMessenger)
        if [ -z "$l1_xdomain_messenger" ]; then
            echo "OVM_L1CrossDomainMessenger not found yet, waiting on deployer..."
            echo "Sleeping 30 sec ($i/$retries)."
            sleep 30
        else
            echovar l1_xdomain_messenger
            pb
            return
        fi
    done
    echo "Timed out... :/"
    echo "Check deployer logs, never able to get OVM_L1CrossDomainMessenger value."
    exit 1
}

pause_proxy_init_l1() {
    git clone https://github.com/dapphub/ds-pause.git $pause_dir || true
    pd $pause_dir
    git pull origin master
    dapp update
    DAPP_BUILD_OPTIMIZE=0 DAPP_BUILD_OPTIMIZE_RUNS=0 dapp --use solc:0.5.9 build

    res=$(dapp create DSPauseProxy)
    pause_address=$(echo "$res" | tail -n 1)
    echovar pause_address
    pb
}

esm_init_l1() {
    git clone https://github.com/makerdao/esm.git $esm_dir || true
    pd $esm_dir
    git pull origin master
    dapp update
    make build

    res=$(dapp create ESM $dai_address $end_address $pause_address $esm_min)
    esm_address=$(echo "$res" | tail -n 1)
    echovar esm_address
    pb
}

dss_init_l1() {
    echo 'Build and deploy dss...'
    echo '  https://github.com/makerdao/dss'
    eth_env_l1

    git clone https://github.com/makerdao/dss.git $dss_dir || true
    pd $dss_dir
    git pull origin master
    dapp update
    make build

    chainid=$(seth chain-id)

    res=$(dapp create Vat)
    vat_address=$(echo "$res" | tail -n 1)
    echovar vat_address

    res=$(dapp create Dai $chainid)
    dai_address=$(echo "$res" | tail -n 1)
    echovar dai_address

    res=$(dapp create Flapper $vat_address $dai_address)
    flap_address=$(echo "$res" | tail -n 1)
    echovar flap_address

    res=$(dapp create Flopper $vat_address $dai_address)
    flop_address=$(echo "$res" | tail -n 1)
    echovar flop_address

    res=$(dapp create Vow $vat_address $flap_address $flop_address)
    vow_address=$(echo "$res" | tail -n 1)
    echovar vow_address

    res=$(dapp create DaiJoin $vat_address $dai_address)
    join_address=$(echo "$res" | tail -n 1)
    echovar join_address

    res=$(dapp create End)
    end_address=$(echo "$res" | tail -n 1)
    echovar end_address

    pause_proxy_init_l1
    esm_init_l1

    j=$(cat<<ENDJSON
{
   "vat_address": "$vat_address",
   "dai_address": "$dai_address",
   "flap_address": "$flap_address",
   "flop_address": "$flop_address",
   "vow_address": "$vow_address",
   "join_address": "$join_address",
   "end_address": "$end_address",
   "pause_address": "$pause_address",
   "esm_address": "$esm_address"
}
ENDJSON
)
    echo $(echo "$j" | jq . | tr -d '\n')
    pb
}

dss_wormhole_init_l1() {
    echo 'Build and deploy dss-wormhole...'
    echo '  https://github.com/makerdao/dss-wormhole'
    git clone https://github.com/makerdao/dss-wormhole.git $dss_wormhole_dir || true
    pd $dss_wormhole_dir
    git pull origin master

    # TODO shouldn't have to manually run this... what's up nix?
    nix-env -f https://github.com/dapphub/dapptools/archive/master.tar.gz -iA solc-static-versions.solc_0_8_13

    dapp update
    make all

    eth_env_l1
    res=$(dapp create WormholeJoin $vat_address $join_address $ilk $master_domain)
    whjoin_address=$(echo "$res" | tail -n 1)
    echovar whjoin_address

    res=$(dapp create WormholeConstantFee 1 1)
    whfee_address=$(echo "$res" | tail -n 1)
    echovar whfee_address

    res=$(dapp create WormholeOracleAuth $whjoin_address)
    whauth_address=$(echo "$res" | tail -n 1)
    echovar whauth_address

    res=$(dapp create WormholeRouter $dai_address)
    whrouter_address=$(echo "$res" | tail -n 1)
    echovar whrouter_address

    res=$(dapp create BasicRelay $whauth_address $join_address)
    whrelay_address=$(echo "$res" | tail -n 1)
    echovar whrelay_address

    j=$(cat<<ENDJSON
{
   "whjoin_address": "$whjoin_address",
   "whfee_address": "$whfee_address",
   "whauth_address": "$whauth_address",
   "whrouter_address": "$whrouter_address",
   "whrelay_address": "$whrelay_address"
}
ENDJSON
)
    echo $(echo "$j" | jq . | tr -d '\n')
    pb
}

optimism_spells() {
    tmp_wd=$optimism_bridge_dir/.tmp
    mkdir -p $tmp_wd
    pd $tmp_wd/

    git clone https://github.com/makerdao/wormhole-integration-tests.git || true
    pd wormhole-integration-tests
    git pull origin master
    git reset --hard origin/master
    pb

    rm -fr src
    mv wormhole-integration-tests/contracts/deploy src
    rm -fr wormhole-integration-tests

    # TODO shouldn't have to manually run this... what's up nix?
    nix-env -f https://github.com/dapphub/dapptools/archive/master.tar.gz -iA solc-static-versions.solc_0_8_13

    DAPP_BUILD_OPTIMIZE=0 DAPP_BUILD_OPTIMIZE_RUNS=0 dapp --use solc:0.8.13 build

    eth_env_l2
    res=$(set -x; dapp create L2AddWormholeDomainSpell $l2_dai_address $l2_wormhole_gateway_address $master_domain)
    l2_domain_spell=$(echo "$res" | tail -n 1)
    echovar l2_domain_spell
    (set -x; seth send $l2_dai_address 'rely(address)' $l2_domain_spell)
    (set -x; seth send $l2_wormhole_gateway_address 'rely(address)' $l2_domain_spell)
    (set -x; seth send $l2_domain_spell 'execute()')

    eth_env_l1
    res=$(set -x; dapp create L1AddWormholeOptimismSpell \
            $slave_domain \
            $whjoin_address \
            $whfee_address \
            $line \
            $whrouter_address \
            $l1_wormhole_gateway_address \
            $l1_escrow_address \
            $dai_address \
            $l1_governance_address \
            $l2_domain_spell
    )
    l1_domain_spell=$(echo "$res" | tail -n 1)
    echovar l1_domain_spell
    (set -x; seth send $whrouter_address 'rely(address)' $l1_domain_spell)
    (set -x; seth send $whjoin_address 'rely(address)' $l1_domain_spell)
    (set -x; seth send $l1_escrow_address 'rely(address)' $l1_domain_spell)
    (set -x; seth send $l1_governance_address 'rely(address)' $l1_domain_spell)
    (set -x; seth send $l1_domain_spell 'execute()')

    rm -fr $tmp_wd
    pb
}

optimism_bridge_init() {
    echo 'Build and deploy Optimism bridge...'

    pd $optimism_bridge_dir/

    # TODO(jamesr) Built locally from https://github.com/makerdao/optimism-dai-bridge.git,
    # which doesn't build without heavy coercion on MacOS. When code is stable,
    # pull from git, build, etc.
    #      Dai.*
    #      L1DAITokenBridge.*
    #      L1DaiWormholeGateway.*
    #      L1Escrow.*
    #      L1GovernanceRelay.*
    #      L2DAITokenBridge.*
    #      L2DaiWormholeGateway.*
    #      L2GovernanceRelay.*
    ls -altr $(pwd)/out

    # NOTE(jamesr) We have to precompute addresses for some contracts in order to deploy.
    eth_env_l2
    l2addr=$(ethutil compute-contract-address $ETH_FROM $(seth nonce $ETH_FROM))

    eth_env_l1
    res=$(set -x; seth send --create $(cat ./out/L1GovernanceRelay.bin) 'L1GovernanceRelay(address,address)' "0x${l2addr}" $l1_xdomain_messenger)
    l1_governance_address=$(echo "$res" | tail -n 1)
    echovar l1_governance_address

    res=$(set -x; seth send --create $(cat out/L1Escrow.bin) 'L1Escrow()')
    l1_escrow_address=$(echo "$res" | tail -n 1)
    echovar l1_escrow_address

    eth_env_l2
    res=$(set -x; seth send --create $(cat ./out/L2GovernanceRelay.bin) 'L2GovernanceRelay(address,address)' $l1_governance_address $l2_xdomain_messenger)
    l2_governance_address=$(echo "$res" | tail -n 1)
    echovar l2_governance_address

    res=$(set -x; seth send --create $(cat out/Dai.bin) 'Dai()')
    l2_dai_address=$(echo "$res" | tail -n 1)
    echovar l2_dai_address

    l2_precomputed_token_bridge_address="0x$(ethutil compute-contract-address $ETH_FROM $(seth nonce $ETH_FROM))"
    echovar l2_precomputed_token_bridge_address

    eth_env_l1
    res=$(
        set -x; seth send --create $(cat ./out/L1DAITokenBridge.bin) \
            'L1DAITokenBridge(address,address,address,address,address)' \
            $dai_address \
            "$l2_precomputed_token_bridge_address" \
            $l2_dai_address \
            $l1_xdomain_messenger \
            $l1_escrow_address
    )
    l1_token_bridge_address=$(echo "$res" | tail -n 1)
    echovar l1_token_bridge_address

    eth_env_l2
    res=$(
        set -x; seth send --create $(cat ./out/L2DAITokenBridge.bin) \
            'L2DAITokenBridge(address,address,address,address)' \
            $l2_xdomain_messenger \
            $l2_dai_address \
            $dai_address \
            $l1_token_bridge_address
    )
    l2_token_bridge_address=$(echo "$res" | tail -n 1)
    [[ "$l2_precomputed_token_bridge_address" = "$l2_token_bridge_address" ]]

    eth_env_l1
    l1_precomputed_wormhole_bridge_address="0x$(ethutil compute-contract-address $ETH_FROM $(seth nonce $ETH_FROM))"
    echovar l1_precomputed_wormhole_bridge_address

    eth_env_l2
    # https://kovan-optimistic.etherscan.io/address/0x45440ae4988965a4cd94651e715fc9a04e62fb41#code
    res=$(set -x; seth send --create $(cat ./out/L2DaiWormholeGateway.bin) \
        'L2DaiWormholeGateway(address,address,address,bytes32)' \
        $l2_xdomain_messenger $l2_dai_address "$l1_precomputed_wormhole_bridge_address" $slave_domain
    )
    l2_wormhole_gateway_address=$(echo "$res" | tail -n 1)
    echovar l2_wormhole_gateway_address

    eth_env_l1
    res=$(
        set -x; seth send --create $(cat ./out/L1DaiWormholeGateway.bin) \
            'L1DaiWormholeGateway(address,address,address,address,address)' \
            $dai_address $l2_wormhole_gateway_address $l1_xdomain_messenger $l1_escrow_address $whrouter_address
    )
    l1_wormhole_gateway_address=$(echo "$res" | tail -n 1)
    echovar l1_wormhole_gateway_address
    [[ "$l1_precomputed_wormhole_bridge_address" = "$l1_wormhole_gateway_address" ]]

    optimism_spells

    j=$(cat<<ENDJSON
{
    "l1_dai_address": "$dai_address",
    "l1_escrow_address": "$l1_escrow_address",
    "l1_governance_address": "$l1_governance_address",
    "l1_token_bridge_address": "$l1_token_bridge_address",
    "l1_wormhole_gateway_address": "$l1_wormhole_gateway_address",
    "l2_dai_address": "$l2_dai_address",
    "l2_governance_address": "$l2_governance_address",
    "l2_token_bridge_address": "$l2_token_bridge_address",
    "l2_wormhole_gateway_address": "$l2_wormhole_gateway_address"
}
ENDJSON
)
    echo $(echo "$j" | jq . | tr -d '\n')
    pb
}

post_deploy_setup() {
    # Mint some DAI on the chains
    eth_env_l1
    for account in $(get_accounts); do
        echo "$account (l1)"
        (set -x; seth send $dai_address 'mint(address,uint256)' $account $esm_min)
    done

    eth_env_l2
    for account in $(get_accounts); do
        echo "$account (l2)"
        (set -x; seth send $l2_dai_address 'mint(address,uint256)' $account $esm_min)
    done
}

ethutil_init
l1_l2_init
start_l1_l2

res=$(dss_init_l1)
dss_json=$(echo "$res" | tail -n 1)
json_to_env "$dss_json"
echo "$dss_json" | jq . | tee $tmp_dir/dss.json

res=$(dss_wormhole_init_l1)
dsswh_json=$(echo "$res" | tail -n 1)
json_to_env "$dsswh_json"
echo "$dsswh_json" | jq . | tee $tmp_dir/dss-wormhole.json

res=$(optimism_bridge_init)
opt_json=$(echo "$res" | tail -n 1)
json_to_env "$opt_json"
echo "$opt_json" | jq . | tee $tmp_dir/optimism.json

post_deploy_setup

# Try a transfer between L2 -> L1. If all goes well you should get a dump of
# events, including WormholeInitialized.
(set -x; \
    seth send $l2_wormhole_gateway_address 'initiateWormhole(bytes32,address,uint128)' \
    $master_domain $ETH_FROM 500
)
ethutil dump-events-l2 /artifacts/L2DaiWormholeGateway.abi $l2_wormhole_gateway_address | jq .

mkdir -p $tmp_dir || true
