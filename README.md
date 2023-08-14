# Oracle Suite

[![Run Tests](https://github.com/chronicleprotocol/oracle-suite/actions/workflows/test.yml/badge.svg)](https://github.com/chronicleprotocol/oracle-suite/actions/workflows/test.yml)
[![Build & Push Docker Image](https://github.com/chronicleprotocol/oracle-suite/actions/workflows/docker.yml/badge.svg)](https://github.com/chronicleprotocol/oracle-suite/actions/workflows/docker.yml)

A set of tools that can be used to run Oracles.

## Gofer

A tool to fetch and calculate reliable asset prices.

see: [Gofer CLI Readme](cmd/gofer/README.md)

## Spire

A peer-to-peer node & client for broadcast signed asset prices.

see: [Spire CLI Readme](cmd/spire/README.md)

## Spire-Bootstrap

A bootstrap node for the Spire network.

see: [Spire Bootstrap CLI Readme](cmd/spire-bootstrap/README.md)

## Leeloo

A tool to observe and attest blockchain events.

see: [Leeloo CLI Readme](cmd/leeloo/README.md)

## Lair

A tool to store and provide HTTP API for blockchain events provided by Leeloo.

see: [Lair CLI Readme](cmd/lair/README.md)

## RPC-Splitter

The Ethereum RPC proxy that splits the request across multiple endpoints to verify that none of them are compromised.

see: [RPC-Splitter CLI Readme](cmd/rpc-splitter/README.md)

## Ghost

A tool used by feeds for broadcast signed prices.

see: [Ghost CLI Readme](cmd/ghost/README.md)

## Spectre

A tool used by relays to update Oracle contracts.

see: [Spectre CLI Readme](cmd/spectre/README.md)


## Setup pre-commit hooks

### Install pre-commit

Using homebrew:
```bash
$ brew install pre-commit
```

Using pip:
```bash
$ pip install pre-commit
```

Check pre-commit version
```bash
$ pre-commit --version
pre-commit 3.3.3
```

### Configure .pre-commit-config.yaml
- You can create a file named `.pre-commit-config.yaml`
- You can generate a very basic configuration using `pre-commit sample-config`
- Or you can directly use [`.pre-commit-config.yaml`](https://github.com/chronicleprotocol/oracle-suite/.pre-commit-hooks.yaml) file in the repository

Example:
```editorconfig
# See https://pre-commit.com for more information
# See https://pre-commit.com/hooks.html for more hooks
repos:
-   repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v3.2.0
    hooks:
    -   id: trailing-whitespace
    -   id: end-of-file-fixer
    -   id: check-yaml
    -   id: check-added-large-files
```

### Install git hook scripts
Run `pre-commit install` to set up the git hooks scripts
```bash
$ pre-commit install
pre-commit installed at .git/hooks/pre-commit
```

Now you're ready to `git commit`!

### Run all the files (optional)
```bash
pre-commit run --all-files
```
