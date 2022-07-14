package errutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	tests := []struct {
		fn    func() (int, error)
		panic bool
		value int
	}{
		{
			fn: func() (int, error) {
				return 1, nil
			},
			panic: false,
			value: 1,
		},
		{
			fn: func() (int, error) {
				return 1, fmt.Errorf("error")
			},
			panic: true,
			value: 1,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			defer func() {
				assert.Equal(t, tt.panic, recover() != nil)
			}()
			assert.Equal(t, tt.value, Must(tt.fn()))
		})
	}
}
