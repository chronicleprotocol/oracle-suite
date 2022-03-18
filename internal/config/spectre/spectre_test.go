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

package spectre

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	datastoreMemory "github.com/chronicleprotocol/oracle-suite/pkg/datastore/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/spectre"
)

func TestSpectre_Configure(t *testing.T) {
	prevSpectreFactory := spectreFactory
	prevDatastoreFactory := datastoreFactory
	defer func() {
		spectreFactory = prevSpectreFactory
		datastoreFactory = prevDatastoreFactory
	}()

	interval := int64(10)
	signer := &ethereumMocks.Signer{}
	ethClient := &ethereumMocks.Client{}
	feeds := []ethereum.Address{ethereum.HexToAddress("0x07a35a1d4b751a818d93aa38e615c0df23064881")}
	ds := &datastoreMemory.Datastore{}
	logger := null.New()

	config := Spectre{
		Interval: interval,
		Medianizers: map[string]Medianizer{
			"AAABBB": {
				Contract:         "0xe0F30cb149fAADC7247E953746Be9BbBB6B5751f",
				OracleSpread:     0.1,
				OracleExpiration: 15500,
				MsgExpiration:    1800,
			},
		},
	}

	spectreFactory = func(cfg spectre.Config) (*spectre.Spectre, error) {
		assert.Equal(t, signer, cfg.Signer)
		assert.Equal(t, ds, cfg.Datastore)
		assert.Equal(t, secToDuration(interval), cfg.Interval)
		assert.Equal(t, logger, cfg.Logger)
		assert.Equal(t, "AAABBB", cfg.Pairs[0].AssetPair)
		assert.Equal(t, secToDuration(config.Medianizers["AAABBB"].OracleExpiration), cfg.Pairs[0].OracleExpiration)
		assert.Equal(t, secToDuration(config.Medianizers["AAABBB"].MsgExpiration), cfg.Pairs[0].PriceExpiration)
		assert.Equal(t, config.Medianizers["AAABBB"].OracleSpread, cfg.Pairs[0].OracleSpread)
		assert.Equal(t, ethereum.HexToAddress(config.Medianizers["AAABBB"].Contract), cfg.Pairs[0].Median.Address())
		return &spectre.Spectre{}, nil
	}

	s, err := config.ConfigureSpectre(Dependencies{
		Signer:         signer,
		Datastore:      ds,
		EthereumClient: ethClient,
		Feeds:          feeds,
		Logger:         logger,
	})
	require.NoError(t, err)
	require.NotNil(t, s)
}

func secToDuration(s int64) time.Duration {
	return time.Duration(s) * time.Second
}
