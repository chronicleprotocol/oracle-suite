# Wormhole E2E

This contains scripts that pull dependencies and stands up a mainnet (L1) and Optimism (L2) ethereum environment that can be used to exercise the tools in `oracle-suite` with regard to [Wormhole](https://forum.makerdao.com/t/introducing-maker-wormhole/11550).

# Getting started

Everything is driven from a Makefile. You will need to initialize the environment first, which is by far the longest step since it will pull dependences, stand up Docker instances, and deploy and initialize contracts and accounts.

```
make init
```

**NOTE** the make targets for these tests are integrated at the top-level of this repo, since we need to run code from above (like Leeloo). So while you can use the local `Makefile` you should really run tests from the git root. See the top-level Makefile for `wormhole-*` targets.

For instance, if you have a running environment, you can simply do

```
make wormhole-e2e
```

from *gitroot* to run the full suite of tests.

**NOTE(2022 April)** The above `make init` relies on repos that are in a very early stage of development and in a pretty high state of flux. As a result, building from source is likely to be frail until the dependent repos settle down. In order to get some stability for E2E tests, snapshots of the result of `make init` have been taken and are in a free account in Dockerhub for the Oracles CU:

[https://hub.docker.com/r/makerocu/e2e-wormhole/tags](https://hub.docker.com/r/makerocu/e2e-wormhole/tags)

The `docker-compose.yml` file that E2E tests rely on will pull these images rather than try and build from source via `make init`.

## Requirements

This E2E tries to keep the dependencies to a minimum:

* Docker
* git
* Bash
* jq
* dapptools
* make

Note that the majority of the dependencies come from standing up the local L1/L2 Docker environment from source, i.e. `make init`.

## Tests

Tests are in `e2e/wormhole/tests`. To add a new test simply create a bash script prefixed by `t_`. E.g the attestation test is `e2e/wormhole/tests/t_attest.sh`. The script `e2e/wormhole/e2e.sh` will look in the `tests/` directory for any `t_*.sh` scripts and run them.

The top-level make target `wormhole-e2e-one-shot` will do everything, from pulling the current docker images, starting them, and running `e2e.sh`. This is the target used by Github CI for wormhole E2E.

# Notes and tips

### Dump wormhole events from your host machine.

If you have Node and the proper dependencies installed, you can check L2 wormhole events using the following incantation from the root of `oracle-suite`:


```
node e2e/wormhole/aux/ethutil/dump-events/main.js \
	http://localhost:8545 \
	./e2e/wormhole/optimism-dai-bridge-artifacts/out/L2DaiWormholeGateway.abi \
	$L2DaiWormholeGateway_ADDRESS | jq .

```

### Secrets

The file `e2e/wormhole/secrets.yaml` is a [sops](https://github.com/mozilla/sops) encrypted file and contains secrets for accessing the [makerocu](https://hub.docker.com/u/makerocu) account at Dockerhub, which is where the docker snapshots for the testing environment are pushed. Once you have the master PGP key, you can view them with:

```
sops e2e/wormhole/secrets.yaml
```

### References

1. [https://github.com/makerdao/optimism-dai-bridge](https://github.com/makerdao/optimism-dai-bridge)
2. [https://github.com/makerdao/wormhole-integration-tests/](https://github.com/makerdao/wormhole-integration-tests/) (Protocol Engineering)
3. [Running a local development environment (Optimism)](https://community.optimism.io/docs/developers/build/dev-node/#setting-up-the-environment)
4. [https://github.com/makerdao/dss-wormhole](https://github.com/makerdao/dss-wormhole)
