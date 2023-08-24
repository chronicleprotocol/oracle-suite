# List of files to include in the order they are specified.
# Supports glob patterns.
# By default, all relative paths will be resolved based on the first config file provided to the application.
# [-c config.hcl] is the the default.
include = [
  "config-contracts.hcl",
  "config-defaults.hcl",
  "config-ethereum.hcl",
  "config-transport.hcl",
  "config-spectre.hcl",
  "config-spire.hcl",
  "config-gofer.hcl",
  "config-ghost.hcl",
]





