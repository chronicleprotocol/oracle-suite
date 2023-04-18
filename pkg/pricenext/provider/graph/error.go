package graph

import (
	"fmt"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type Errors []error

func (e Errors) Error() string {
	b := strings.Builder{}
	b.WriteString("following errors occurred: ")
	for i, err := range e {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

type ErrInvalidTick struct {
	Tick provider.Tick
}

func (e ErrInvalidTick) Error() string {
	return fmt.Sprintf("invalid tick %s: %v", e.Tick, e.Tick.Validate())
}
