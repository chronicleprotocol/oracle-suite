package store

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type mockRecoverer struct{}

func (r *mockRecoverer) Supports(_ context.Context, data datapoint.Point) bool {
	return true
}

func (r *mockRecoverer) Recover(_ context.Context, _ string, p datapoint.Point, _ types.Signature) (*types.Address, error) {
	return types.MustAddressFromHexPtr(p.Meta["addr"].(string)), nil
}

var (
	aaabbb1 = &messages.DataPoint{
		Model: "AAABBB",
		Point: datapoint.Point{
			Value: value.StaticValue{Value: bn.DecFloatPoint(1)},
			Time:  time.Unix(1234567890, 0),
			Meta: map[string]any{
				"addr": "0x1111111111111111111111111111111111111111",
			},
		},
		ECDSASignature: types.MustSignatureFromBytes(bytes.Repeat([]byte{0x01}, 65)),
	}
	aaabbb2 = &messages.DataPoint{
		Model: "AAABBB",
		Point: datapoint.Point{
			Value: value.StaticValue{Value: bn.DecFloatPoint(2)},
			Time:  time.Unix(1234567890, 0),
			Meta: map[string]any{
				"addr": "0x2222222222222222222222222222222222222222",
			},
		},
		ECDSASignature: types.MustSignatureFromBytes(bytes.Repeat([]byte{0x01}, 65)),
	}
	xxxyyy1 = &messages.DataPoint{
		Model: "XXXYYY",
		Point: datapoint.Point{
			Value: value.StaticValue{Value: bn.DecFloatPoint(3)},
			Time:  time.Unix(1234567890, 0),
			Meta: map[string]any{
				"addr": "0x1111111111111111111111111111111111111111",
			},
		},
		ECDSASignature: types.MustSignatureFromBytes(bytes.Repeat([]byte{0x01}, 65)),
	}
	xxxyyy2 = &messages.DataPoint{
		Model: "XXXYYY",
		Point: datapoint.Point{
			Value: value.StaticValue{Value: bn.DecFloatPoint(4)},
			Time:  time.Unix(1234567891, 0),
			Meta: map[string]any{
				"addr": "0x2222222222222222222222222222222222222222",
			},
		},
		ECDSASignature: types.MustSignatureFromBytes(bytes.Repeat([]byte{0x01}, 65)),
	}
)

func TestStore(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	transport := local.New(
		[]byte("test"),
		0,
		map[string]transport.Message{messages.DataPointV1MessageName: (*messages.DataPoint)(nil)},
	)
	require.NoError(t, transport.Start(ctx))

	// Wait to be sure that the transport is ready.
	time.Sleep(100 * time.Millisecond)

	store, err := New(Config{
		Storage:    NewMemoryStorage(),
		Transport:  transport,
		Models:     []string{"AAABBB", "XXXYYY"},
		Recoverers: []datapoint.Recoverer{&mockRecoverer{}},
	})
	require.NoError(t, err)
	require.NoError(t, store.Start(ctx))

	// Wait to be sure that the store is ready.
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, transport.Broadcast(messages.DataPointV1MessageName, aaabbb1))
	assert.NoError(t, transport.Broadcast(messages.DataPointV1MessageName, aaabbb2))
	assert.NoError(t, transport.Broadcast(messages.DataPointV1MessageName, xxxyyy1))
	assert.NoError(t, transport.Broadcast(messages.DataPointV1MessageName, xxxyyy2))

	// Wait to be sure that the store has processed the messages.
	assert.Eventually(t, func() bool {
		a, _ := store.Latest(context.Background(), "AAABBB")
		b, _ := store.Latest(context.Background(), "XXXYYY")
		return len(a) == 2 && len(b) == 2
	}, 1*time.Second, 100*time.Millisecond)

	// Verify if the messages are stored correctly.
	a, _ := store.Latest(context.Background(), "AAABBB")
	b, _ := store.Latest(context.Background(), "XXXYYY")
	assert.Equal(t, "1", a[types.MustAddressFromHex("0x1111111111111111111111111111111111111111")].DataPoint.Value.Print())
	assert.Equal(t, "2", a[types.MustAddressFromHex("0x2222222222222222222222222222222222222222")].DataPoint.Value.Print())
	assert.Equal(t, "3", b[types.MustAddressFromHex("0x1111111111111111111111111111111111111111")].DataPoint.Value.Print())
	assert.Equal(t, "4", b[types.MustAddressFromHex("0x2222222222222222222222222222222222222222")].DataPoint.Value.Print())
}
