package webapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

func TestMultiAddressBook_Consumers(t *testing.T) {
	tests := []struct {
		firstAddressBook  []string
		secondAddressBook []string
		expectedAddresses []string
	}{
		{
			firstAddressBook:  nil,
			secondAddressBook: nil,
			expectedAddresses: nil,
		},
		{
			firstAddressBook:  []string{},
			secondAddressBook: []string{},
			expectedAddresses: nil,
		},
		{
			firstAddressBook: []string{
				"domain1.example",
				"domain2.example",
			},
			secondAddressBook: nil,
			expectedAddresses: []string{
				"domain1.example",
				"domain2.example",
			},
		},
		{
			firstAddressBook: nil,
			secondAddressBook: []string{
				"domain1.example",
				"domain2.example",
			},
			expectedAddresses: []string{
				"domain1.example",
				"domain2.example",
			},
		},
		{
			firstAddressBook: []string{
				"domain1.example",
				"domain2.example",
			},
			secondAddressBook: []string{
				"domain1.example",
				"domain2.example",
			},
			expectedAddresses: []string{
				"domain1.example",
				"domain2.example",
			},
		},
		{
			firstAddressBook: []string{
				"domain1.example",
				"domain2.example",
				"domain3.example",
			},
			secondAddressBook: []string{
				"domain1.example",
				"domain2.example",
				"domain4.example",
			},
			expectedAddresses: []string{
				"domain1.example",
				"domain2.example",
				"domain3.example",
				"domain4.example",
			},
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			ctx := context.Background()
			book := NewMultiAddressBook(
				NewStaticAddressBook(tt.firstAddressBook),
				NewStaticAddressBook(tt.secondAddressBook),
			)
			consumers, err := book.Consumers(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAddresses, consumers)
		})
	}
}

func TestStaticAddressBook_Consumers(t *testing.T) {
	tests := []struct{ addresses []string }{
		{addresses: nil},
		{addresses: []string{}},
		{addresses: []string{
			"domain1.example",
			"domain2.example",
		}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			ctx := context.Background()
			book := NewStaticAddressBook(tt.addresses)
			consumers, err := book.Consumers(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.addresses, consumers)
		})
	}
}

func TestEthereumAddressBook_Consumers(t *testing.T) {
	tests := []struct{ addresses []string }{
		{addresses: []string{}},
		{addresses: []string{
			"domain1.example",
			"domain2.example",
		}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var (
				ctx   = context.Background()
				rpc   = &mocks.RPC{}
				to    = types.MustAddressFromHex("0x1234567890123456789012345678901234567890")
				input = hexutil.MustHexToBytes("0x0f560cd7")
			)
			rpc.On("Call", ctx, types.Call{
				To:    &to,
				Input: input,
			}, types.LatestBlockNumber).Return(encodeAddresses(tt.addresses), nil).Once()
			book := NewEthereumAddressBook(rpc, to, time.Minute)
			consumers, err := book.Consumers(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.addresses, consumers)
		})
	}
}

func TestEthereumAddressBook_Cache(t *testing.T) {
	var (
		ctx   = context.Background()
		rpc   = &mocks.RPC{}
		to    = types.MustAddressFromHex("0x1234567890123456789012345678901234567890")
		input = hexutil.MustHexToBytes("0x0f560cd7")
	)

	// Call method should be called once, because the result is cached.
	rpc.On("Call", ctx, types.Call{
		To:    &to,
		Input: input,
	}, types.LatestBlockNumber).Return(encodeAddresses([]string{"domain1.example"}), nil).Once()
	book := NewEthereumAddressBook(rpc, to, time.Second)
	_, err := book.Consumers(ctx)
	require.NoError(t, err)
	_, err = book.Consumers(ctx)
	require.NoError(t, err)

	// After one second, the cache is invalided.
	time.Sleep(time.Second)
	rpc.On("Call", ctx, types.Call{
		To:    &to,
		Input: input,
	}, types.LatestBlockNumber).Return(encodeAddresses([]string{"domain1.example"}), nil).Once()
	_, err = book.Consumers(ctx)
	require.NoError(t, err)
}

func encodeAddresses(addresses []string) []byte {
	return errutil.Must(abi.EncodeValues(consumersMethod.Outputs(), addresses))
}
