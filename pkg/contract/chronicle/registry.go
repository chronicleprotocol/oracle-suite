//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package chronicle

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/defiweb/go-eth/types"
)

type Deployment struct {
	Address types.Address
	ChainID uint64
	Feeds   []types.Address
	Wat     string
	Bar     int
}

// Registry is a wrapper around the FeedRegistry and WatRegistry contracts.
//
// It provides a single interface for querying on-chain registries.
type Registry struct {
	feedRegistry *FeedRegistry
	watRegistry  *WatRegistry
}

func NewRegistry(feedRegistry *FeedRegistry, watRegistry *WatRegistry) (*Registry, error) {
	if feedRegistry == nil {
		return nil, fmt.Errorf("feed registry is nil")
	}
	if watRegistry == nil {
		return nil, fmt.Errorf("wat registry is nil")
	}
	if feedRegistry.Client() != watRegistry.Client() {
		return nil, fmt.Errorf("feed registry and wat registry must use the same client")
	}
	return &Registry{
		feedRegistry: feedRegistry,
		watRegistry:  watRegistry,
	}, nil
}

// Deployments returns a list of all deployed contracts from the WatRegistry.
func (r *Registry) Deployments(ctx context.Context) ([]Deployment, error) {
	blockNumber, err := r.feedRegistry.Client().BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("on-chain registry: %w", err)
	}
	feeds, err := r.feedRegistry.Feeds().Call(ctx, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return nil, fmt.Errorf("on-chain registry: %w", err)
	}
	wats, err := r.watRegistry.Wats().Call(ctx, types.BlockNumberFromBigInt(blockNumber))
	if err != nil {
		return nil, fmt.Errorf("on-chain registry: %w", err)
	}
	var deployments []Deployment
	for _, wat := range wats {
		config, err := r.watRegistry.Config(wat).Call(ctx, types.BlockNumberFromBigInt(blockNumber))
		if err != nil {
			return nil, fmt.Errorf("on-chain registry: %w", err)
		}
		chainIDs, err := r.watRegistry.Chains(wat).Call(ctx, types.BlockNumberFromBigInt(blockNumber))
		if err != nil {
			return nil, fmt.Errorf("on-chain registry: %w", err)
		}
		for _, chainID := range chainIDs {
			address, err := r.watRegistry.Deployment(wat, chainID).Call(ctx, types.BlockNumberFromBigInt(blockNumber))
			if err != nil {
				return nil, fmt.Errorf("on-chain registry: %w", err)
			}
			var liftedFeeds []types.Address
			for _, feed := range feeds {
				if config.Bloom.Has(feed) {
					liftedFeeds = append(liftedFeeds, feed)
				}
			}
			deployments = append(deployments, Deployment{
				Address: address,
				ChainID: chainID,
				Feeds:   liftedFeeds,
				Wat:     wat,
				Bar:     config.Bar,
			})
		}
	}
	sort.Slice(deployments, func(i, j int) bool {
		return bytes.Compare(deployments[i].Address.Bytes(), deployments[j].Address.Bytes()) < 0
	})
	return deployments, nil
}
