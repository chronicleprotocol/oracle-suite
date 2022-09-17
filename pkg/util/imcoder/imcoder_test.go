package imcoder

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImcoder(t *testing.T) {
	data := make([]byte, 1024*10)
	rand.Read(data)

	tests := []struct {
		dataLength  int
		blockSize   uint
		bitsPerChan uint
		wantErr     bool
	}{
		// Different bits per channel values:
		{1024 * 10, 4, 1, false},
		{1024 * 10, 4, 2, false},
		{1024 * 10, 4, 3, false},
		{1024 * 10, 4, 4, false},
		{1024 * 10, 4, 5, false},
		{1024 * 10, 4, 6, false},
		{1024 * 10, 4, 7, false},
		{1024 * 10, 4, 8, false},
		// Different data length:
		{0, 1, 1, false},
		{1, 1, 1, false},
		// Invalid options:
		{1024 * 10, 1, 0, true},
		{1024 * 10, 1, 9, true},
		{1024 * 10, 0, 1, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d-%d-%d", tt.blockSize, tt.bitsPerChan, tt.dataLength), func(t *testing.T) {
			opts := Options{BlockSize: tt.blockSize, BitsPerChan: tt.bitsPerChan}
			img, err := Encode(data[:tt.dataLength], opts)
			if err != nil && tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			dec, err := Decode(img)
			require.NoError(t, err)
			assert.Equal(t, data[:tt.dataLength], dec)
		})
	}
}

func FuzzDecode(f *testing.F) {
	f.Fuzz(func(t *testing.T, rgb []byte) {
		rgb = append([]byte{0, 0, 0, 255, 255, 255}, rgb...)
		img := drawImage(rgb, 4)
		dec, err := Decode(img)
		if len(dec) > 0 {
			assert.Error(t, err)
		}
	})
}
