# Spire-Bootstrap CLI Readme

Spire-Bootstrap starts the libp2p bootstrap node for the Spire network.

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go get -u github.com/chronicleprotocol/oracle-suite/cmd/spire-bootstrap`.

Alternatively, you can build Spire-Bootstrap using `Makefile` directly from the repository. This approach is recommended
if you
wish to work on Spire-Bootstrap source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Spire-Bootstrap, you have to create configuration file first. By default, the default config file
location is `config.json` in the current working directory. You can change the config file location using the `--config` 
flag. Spire-Bootstrap supports JSON and YAML configuration files.

### Example configuration

```json
{
  "transport": {
    "transport": "libp2p",
    "libp2p": {
      "privKeySeed": "02082cf471002b5c5dfefdd6cbd30666ff02c4df90169f766877caec26ed4f88",
      "listenAddrs": [
        "/ip4/0.0.0.0/tcp/8000"
      ],
      "bootstrapAddrs": [],
      "directPeersAddrs": [],
      "blockedAddrs": [],
      "disableDiscovery": false
    }
  }
}
```

### Configuration reference

- `transport` - Configuration parameters for transports mechanisms used to relay messages.
    - `transport` (string) - Transport to use. Currently only the `libp2p` transport if supported. This field could be
      omitted.
    - `libp2p` - Configuration parameters for the libp2p transport.
        - `privKeySeed` (`string`) - The random hex-encoded 32 bytes. It is used to generate a unique identity on the
          libp2p network. The value may be empty to generate a random seed.
        - `listenAddrs` (`[]string`) - List of listening addresses for libp2p node encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `bootstrapAddrs` (`[]string`) - List of addresses of bootstrap nodes for the libp2p node encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `directPeersAddrs` (`[]string`) - List of direct peer addresses to which messages will be sent directly.
          Addresses are encoded using the format. [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
          This option must be configured symmetrically on both ends.
        - `blockedAddrs` (`[]string`) - List of blocked peers or IP addresses encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `disableDiscovery` (`bool`) - Disables node discovery. If enabled, the IP address of a node will not be
          broadcast to other peers. This option must be used together with `directPeersAddrs`.
- `feeds` (`[]string`) - List of hex-encoded addresses of other Oracles. Event messages from Oracles outside that list
  will be ignored.
- `logger` - Optional logger configuration.
    - `grafana` - Configuration of Grafana logger. Grafana logger can extract values from log messages and send them to
      Grafana Cloud.
        - `enable` (`string`) - Enable Grafana metrics.
        - `interval` (`int`) - Specifies how often, in seconds, logs should be sent to the Grafana Cloud server. Logs
          with the same name in that interval will be replaced with never ones.
        - `endpoint` (`string`) - Graphite server endpoint.
        - `apiKey` (`string`) - Graphite API key.
        - `[]metrics` - List of metric definitions
            - `matchMessage` (`string`) - Regular expression that must match a log message.
            - `matchFields` (`[string]string`) - Map of fields whose values must match a regular expression.
            - `name` (`string`) - Name of metric. It can contain references to log fields in the format `$${path}`,
              where
              path is the dot-separated path to the field.
            - `tags` (`[string][]string`) - List of metric tags. They can contain references to log fields in the
              format `${path}`, where path is the dot-separated path to the field.
            - `value` (`string`) - Dot-separated path of the field with the metric value. If empty, the value 1 will be
              used as the metric value.
            - `scaleFactor` (`float`) - Scales the value by the specified number. If it is zero, scaling is not
              applied (
              default: 0).
            - `onDuplicate` (`string`) - Specifies how duplicated values in the same interval should be handled. Allowed
              options are:
                - `sum` - Add values.
                - `sub` - Subtract values.
                - `max` - Use higher value.
                - `min` - Use lower value.
                - `replace` (default) - Replace the value with a newer one.

### Environment variables

It is possible to use environment variables anywhere in the configuration file. The syntax is similar as in the
shell: `${ENV_VAR}`. If the environment  variable is not set, the error will be returned during the application
startup. To escape the dollar sign, use `\$` or `$$`. The latter syntax is not supported inside variables. It is
possible to define default values for environment variables. To do so, use the following syntax: `${ENV_VAR-default}`.

## Commands

```
Usage:
  spire-bootstrap [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  run         Starts bootstrap node

Flags:
  -c, --config string                                  ghost config file (default "./config.json")
  -h, --help                                           help for spire-bootstrap
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for spire-bootstrap

Use "spire-bootstrap [command] --help" for more information about a command.

```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
