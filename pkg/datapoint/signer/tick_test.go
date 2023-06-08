package signer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Hash for the AAABBB asset pair, with the price set to 42 and the age to 1605371361:
var priceHash = "0xc678b27c20ef30b95452d8d61f8f3916899717692d8a01c595971035b25a00ff"

/*
func TestTickHandler(t *testing.T) {
	tests := []struct {
		name          string
		dataPoint     data.Point
		signer        *mocks.Key
		expectedEvent *messages.Event
		expectedError bool
	}{
		{
			name: "valid data point",
			dataPoint: data.Point{
				Value: origin.Tick{
					Pair:      origin.Pair{Base: "AAA", Quote: "BBB"},
					Price:     bn.Float(42),
					Volume24h: bn.Float(100),
				},
				Time: time.Unix(1605371361, 0),
			},
			expectedEvent: &messages.Event{
				Type:        tickMessageName,
				ID:          hexutil.MustHexToBytes(priceHash),
				Index:       hexutil.MustHexToBytes(priceHash),
				EventDate:   time.Unix(1605371361, 0),
				MessageDate: time.Now(),
				Data: map[string][]byte{
					"val": bn.Float(42).Mul(bn.Float(priceMultiplier)).BigInt().Bytes(),
					"age": bn.Int(1605371361).BigInt().Bytes(),
					"wat": []byte("AAABBB"),
				},
			},
		},
		{
			name: "invalid tick",
			dataPoint: data.Point{
				Value: origin.Tick{
					Pair:  origin.Pair{Base: "AAA", Quote: "BBB"},
					Price: bn.Float(-1),
				},
				Time: time.Unix(1605371361, 0),
			},
			expectedError: true,
		},
		{
			name: "invalid data point",
			dataPoint: data.Point{
				Value: origin.Tick{
					Pair:  origin.Pair{Base: "AAA", Quote: "BBB"},
					Price: bn.Float(42),
				},
				Time:  time.Unix(1605371361, 0),
				Error: errors.New("something went wrong"),
			},
			expectedError: true,
		},
		{
			name: "invalid value",
			dataPoint: data.Point{
				Value: nil,
				Time:  time.Unix(1605371361, 0),
				Error: errors.New("something went wrong"),
			},
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer := &mocks.Key{}
			handler := NewTickHandler(signer)

			signer.On("SignMessage", mock.Anything).Return(types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0x01}, 65)), nil)
			signer.On("Address", mock.Anything).Return(types.MustAddressFromBytes(bytes.Repeat([]byte{0x01}, 20)), nil)

			event, err := handler.Handle("AAABBB", tt.dataPoint)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, handler.Supports(tt.dataPoint))
			assert.Equal(t, tt.expectedEvent.Type, event.Type)
			assert.Equal(t, tt.expectedEvent.ID, event.ID)
			assert.Equal(t, tt.expectedEvent.Index, event.Index)
			assert.Equal(t, tt.expectedEvent.EventDate, event.EventDate)
			assert.InDelta(t, tt.expectedEvent.MessageDate.Unix(), event.MessageDate.Unix(), 1)
			require.NotNil(t, event.Data["val"])
			require.NotNil(t, event.Data["age"])
			require.NotNil(t, event.Data["wat"])
			require.NotNil(t, event.Data["trace"])
			assert.Equal(t, tt.expectedEvent.Data["val"], event.Data["val"])
			assert.Equal(t, tt.expectedEvent.Data["age"], event.Data["age"])
			assert.Equal(t, tt.expectedEvent.Data["wat"], event.Data["wat"])
			require.NotNil(t, event.Signatures["ethereum"])
			assert.Equal(t, bytes.Repeat([]byte{0x01}, 20), event.Signatures["ethereum"].Signer)
			assert.Equal(t, bytes.Repeat([]byte{0x01}, 65), event.Signatures["ethereum"].Signature)
		})
	}
}

*/

func TestHashTick(t *testing.T) {
	assert.Equal(t, priceHash, hashTick("AAABBB", bn.Float(42), time.Unix(1605371361, 0)).String())
}
