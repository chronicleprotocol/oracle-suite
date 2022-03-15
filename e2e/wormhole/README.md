# Wormhole E2E

This contains scripts that pull dependencies and stands up a mainnet (L1) and Optimism (L2) ethereum environment that can be used to exercise the tools in `oracle-suite` with regard to [Wormhole](https://forum.makerdao.com/t/introducing-maker-wormhole/11550).

# Getting started

Everything is driven from a Makefile. You will need to initialize the environment first, which is by far the longest step since it will pull dependences, stand up Docker instances, and deploy and initialize contracts and accounts.

```
make init
```

**NOTE** the make targets for these tests are integrated at the top-level of this repo, since we need to run code from above (like Leeloo). So while you can use the local `Makefile` you should really run tests from the git root:

```
make wormhole-e2e-init
```

This initializes the environment and starts Leeloo.

```
make wormhole-e2e
```

This will create a `initiateWormhole` event.

```
make wormhole-e2e-clean
```

Does what you expect.

## Requirements

These tests attempt to keep dependencies to a minimimum:

* Docker
* git
* Bash
* jq
* dapptools
* make

# Notes and tips

### Dump wormhole events from your host machine.

If you have Node and the proper dependencies installed, you can check L2 wormhole events using the following incantation from the root of `oracle-suite`:


```
node e2e/wormhole/aux/ethutil/dump-events/main.js \
	http://localhost:8545 \
	./e2e/wormhole/optimism-dai-bridge-artifacts/out/L2DaiWormholeGateway.abi \
	$L2DaiWormholeGateway_ADDRESS | jq .

```

### References

1. [https://github.com/makerdao/optimism-dai-bridge](https://github.com/makerdao/optimism-dai-bridge)
2. [https://github.com/makerdao/wormhole-integration-tests/](https://github.com/makerdao/wormhole-integration-tests/) (Protocol Engineering)
3. [Running a local development environment (Optimism)](https://community.optimism.io/docs/developers/build/dev-node/#setting-up-the-environment)
4. [https://github.com/makerdao/dss-wormhole](https://github.com/makerdao/dss-wormhole)
