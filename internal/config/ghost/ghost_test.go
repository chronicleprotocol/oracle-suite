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

package ghost

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/ghost"
	goferMocks "github.com/chronicleprotocol/oracle-suite/pkg/gofer/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
)

func TestGhost_Configure(t *testing.T) {
	prevGhostFactory := ghostFactory
	defer func() { ghostFactory = prevGhostFactory }()

	interval := 10
	pairs := []string{"AAABBB", "XXXYYY"}
	gofer := &goferMocks.Gofer{}
	signer := &ethereumMocks.Signer{}
	transport := local.New([]byte("test"), 0, nil)
	logger := null.New()

	config := Ghost{
		Interval: interval,
		Pairs:    pairs,
	}

	ghostFactory = func(cfg ghost.Config) (*ghost.Ghost, error) {
		assert.Equal(t, time.Duration(interval)*time.Second, cfg.Interval)
		assert.Equal(t, pairs, cfg.Pairs)
		assert.Equal(t, signer, cfg.Signer)
		assert.Equal(t, transport, cfg.Transport)
		assert.Equal(t, logger, cfg.Logger)

		return &ghost.Ghost{}, nil
	}

	g, err := config.Configure(Dependencies{
		Gofer:     gofer,
		Signer:    signer,
		Transport: transport,
		Logger:    logger,
	})
	require.NoError(t, err)
	assert.NotNil(t, g)
}
