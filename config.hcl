# List of files to include in the order they are specified.
# Supports glob patterns.
# By default, all relative paths will be resolved based on the first config file provided to the application.
# [-c config.hcl] is the the default.
include = [
  "config-contract-params.hcl", # auto generated, but needs to be manually adjusted
  "config-contracts.hcl", # auto generated from contract configurations
  "config-contracts-median.hcl", # legacy median contracts
  "config-defaults.hcl",
  "config-ethereum.hcl",
  "config-transport.hcl",
  "config-spectre.hcl",
  "config-spire.hcl",
  "config-gofer.hcl",
  "config-ghost.hcl",
]
