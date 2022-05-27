package starknet

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

type Felt struct {
	*big.Int
}

func HexToFelt(s string) *Felt {
	f := new(Felt)
	f.Int = new(big.Int)
	f.Int, _ = f.Int.SetString(strings.TrimPrefix(s, "0x"), 16)
	return f
}

func (f Felt) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"0x%s"`, f.Text(16))), nil
}

func (f *Felt) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}
	f.Int = new(big.Int)
	s := string(p[1 : len(p)-1])
	if has0xPrefix(s) {
		s := s[2:]
		if len(s)%2 == 1 {
			s = "0" + s
		}
		b, err := hex.DecodeString(s)
		if err != nil {
			return err
		}
		f.Int.SetBytes(b)
		return nil
	}
	return nil
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}
