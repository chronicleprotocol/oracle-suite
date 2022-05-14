package rpcsplitter

import (
	"time"

	gethRPC "github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type Options func(s *server) error

func WithEndpoints(endpoints []string) Options {
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

func WithRequirements(minResponses int, maxBlockBehind int) Options {
	return func(s *server) error {
		s.defaultResolver = &defaultResolver{minResponses: minResponses}
		s.gasValueResolver = &gasValueResolver{minResponses: minResponses}
		s.blockNumberResolver = &blockNumberResolver{minResponses: minResponses, maxBlocksBehind: maxBlockBehind}
		return nil
	}
}

func WithTotalTimeout(t time.Duration) Options {
	return func(s *server) error {
		s.totalTimeout = t
		return nil
	}
}

func WithGracefulTimeout(t time.Duration) Options {
	return func(s *server) error {
		s.gracefulTimeout = t
		return nil
	}
}

func WithLogger(logger log.Logger) Options {
	return func(s *server) error {
		s.log = logger
		return nil
	}
}

func withCallers(callers map[string]caller) Options {
	return func(s *server) error {
		s.callers = callers
		return nil
	}
}
