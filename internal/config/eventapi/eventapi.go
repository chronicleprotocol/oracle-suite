//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package eventapi

import (
	"context"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

//nolint
var eventAPIFactory = func(ctx context.Context, cfg api.Config) (*api.EventAPI, error) {
	return api.NewEventAPI(ctx, cfg)
}

type EventAPI struct {
	Address string `json:"address"`
}

type Dependencies struct {
	Context    context.Context
	EventStore *store.EventStore
	Transport  transport.Transport
	Logger     log.Logger
}

type DatastoreDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *EventAPI) Configure(d Dependencies) (*api.EventAPI, error) {
	cfg := api.Config{
		EventStore: d.EventStore,
		Address:    c.Address,
		Logger:     d.Logger,
	}
	return eventAPIFactory(d.Context, cfg)
}
