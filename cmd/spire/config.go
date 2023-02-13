package main

import (
	"context"
	"path/filepath"

	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

func PrepareAgentServices(ctx context.Context, opts *options) (*supervisor.Supervisor, error) {
	switch filepath.Ext(opts.ConfigFilePath) {
	case ".hcl":
		return PrepareAgentServicesHCL(ctx, opts)
	default:
		return PrepareAgentServicesYAML(ctx, opts)
	}
}

func PrepareClientServices(ctx context.Context, opts *options) (*supervisor.Supervisor, *spire.Client, error) {
	switch filepath.Ext(opts.ConfigFilePath) {
	case ".hcl":
		return PrepareClientServicesHCL(ctx, opts)
	default:
		return PrepareClientServicesYAML(ctx, opts)
	}
}

func PrepareStreamServices(ctx context.Context, opts *options) (*supervisor.Supervisor, transport.Transport, error) {
	switch filepath.Ext(opts.ConfigFilePath) {
	case ".hcl":
		return PrepareStreamServicesHCL(ctx, opts)
	default:
		return PrepareStreamServicesYAML(ctx, opts)
	}
}
