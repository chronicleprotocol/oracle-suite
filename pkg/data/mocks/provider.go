package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
)

type Provider struct {
	mock.Mock
}

func (p *Provider) ModelNames(ctx context.Context) []string {
	args := p.Called(ctx)
	return args.Get(0).([]string)
}

func (p *Provider) DataPoint(ctx context.Context, model string) (data.Point, error) {
	args := p.Called(ctx, model)
	return args.Get(0).(data.Point), args.Error(1)

}

func (p *Provider) DataPoints(ctx context.Context, models ...string) (map[string]data.Point, error) {
	args := p.Called(ctx, models)
	return args.Get(0).(map[string]data.Point), args.Error(1)
}

func (p *Provider) Model(ctx context.Context, model string) (data.Model, error) {
	args := p.Called(ctx, model)
	return args.Get(0).(data.Model), args.Error(1)
}

func (p *Provider) Models(ctx context.Context, models ...string) (map[string]data.Model, error) {
	args := p.Called(ctx, models)
	return args.Get(0).(map[string]data.Model), args.Error(1)
}
