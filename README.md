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
