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

package rpcsplitter

import (
	"time"

	gethRPC "github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type Option func(s *server) error

// WithEndpoints options instructs RPC-Splitter to use provided list of
// Ethereum RPC nodes.
func WithEndpoints(endpoints []string) Option {
	return func(s *server) error {
		for _, e := range endpoints {
			c, err := gethRPC.Dial(e)
			if err != nil {
				return err
			}
			s.callers[e] = c
		}
		return nil
	}
}

// WithRequirements specifies the requirements that must be met in order for
// responses to be considered valid.
//
// minResponses - minimum number of same responses, ideally this number should
// be greater than half the number of endpoints provided. For methods that
// return a gas price or block number, this is the number of non-error
// responses.
//
// maxBlockBehind - the maximum number of blocks after the last known block.
// Because some RPC endpoints may be behind others, the RPC-Splitter uses the
// lowest block number of all responses, but the difference from the last known
// cannot be less than specified in this parameter.
func WithRequirements(minResponses int, maxBlockBehind int) Option {
	return func(s *server) error {
		s.defaultResolver = &defaultResolver{minResponses: minResponses}
		s.gasValueResolver = &gasValueResolver{minResponses: minResponses}
		s.blockNumberResolver = &blockNumberResolver{minResponses: minResponses, maxBlocksBehind: maxBlockBehind}
		return nil
	}
}

// WithTotalTimeout sets the total timeout for all endpoints. When the timeout
// is exceeded, RPC-Splitter cancels all requests to the endpoints.
func WithTotalTimeout(t time.Duration) Option {
	return func(s *server) error {
		s.totalTimeout = t
		return nil
	}
}

// WithGracefulTimeout sets a timeout for slower endpoints. If the RPC-Splitter
// gets enough responses to return a valid response, it will wait until the
// timeout for slower endpoints is exceeded. This will allow slower requests
// to be gracefully finished, and for endpoints that calculate the median value,
// it will return a more accurate response.
func WithGracefulTimeout(t time.Duration) Option {
	return func(s *server) error {
		s.gracefulTimeout = t
		return nil
	}
}

// WithLogger sets logger.
func WithLogger(logger log.Logger) Option {
	return func(s *server) error {
		s.log = logger
		return nil
	}
}

func withCallers(callers map[string]caller) Option {
	return func(s *server) error {
		s.callers = callers
		return nil
	}
}
