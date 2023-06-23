//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package geth

import (
	"context"
	"errors"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
)

var ErrStorageQueryFailed = errors.New("oracle contract storage query failed")

// TODO: make it configurable
const gasLimit = 200000
const maxReadRetries = 3
const delayBetweenReadRetries = 5 * time.Second

// Median implements the oracle.Median interface using go-ethereum packages.
type Median struct {
	ethereum ethereum.Client //nolint:staticcheck // deprecated ethereum.Client
	address  types.Address
}

// NewMedian creates the new Median instance.
//
//nolint:staticcheck // deprecated ethereum.Client
func NewMedian(ethereum ethereum.Client, address types.Address) *Median {
	return &Median{
		ethereum: ethereum,
		address:  address,
	}
}

// Address implements the oracle.Median interface.
func (m *Median) Address() types.Address {
	return m.address
}

// Age implements the oracle.Median interface.
func (m *Median) Age(ctx context.Context) (time.Time, error) {
	var age int64
	if err := m.read(ctx, "age", nil, []any{&age}); err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(age, 0), nil
}

// Bar implements the oracle.Median interface.
func (m *Median) Bar(ctx context.Context) (int64, error) {
	var bar int64
	if err := m.read(ctx, "bar", nil, []any{&bar}); err != nil {
		return 0, err
	}
	return bar, nil
}

// Wat implements the oracle.Median interface.
func (m *Median) Wat(ctx context.Context) (string, error) {
	wat := [32]byte{}
	if err := m.read(ctx, "wat", nil, []any{&wat}); err != nil {
		return "", err
	}
	return string(wat[:]), nil
}

// Val implements the oracle.Median interface.
func (m *Median) Val(ctx context.Context) (*big.Int, error) {
	const (
		offset = 16
		length = 16
	)
	b, err := m.ethereum.Storage(ctx, m.address, types.MustHashFromBigInt(big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	if len(b) < (offset + length) {
		return nil, ErrStorageQueryFailed
	}
	return new(big.Int).SetBytes(b[length : offset+length]), err
}

// Feeds implements the oracle.Median interface.
func (m *Median) Feeds(ctx context.Context) ([]types.Address, error) {
	var (
		err   error
		orcl  []types.Address
		calls []types.Call
	)

	// Prepare the call list:
	for i := 0; i < 256; i++ {
		cd, err := medianABI.Methods["slot"].EncodeArgs(uint8(i))
		if err != nil {
			return nil, err
		}
		calls = append(calls, types.Call{
			To:    &m.address,
			Input: cd,
		})
	}

	// Call:
	var results [][]byte
	err = retry(maxReadRetries, delayBetweenReadRetries, func() error {
		results, err = m.ethereum.MultiCall(ctx, calls)
		return err
	})

	// Parse results:
	for _, data := range results {
		var addr types.Address
		if err := medianABI.Methods["slot"].DecodeValues(data, &addr); err != nil {
			return nil, err
		}
		orcl = append(orcl, addr)
	}

	return orcl, nil
}

// Poke implements the oracle.Median interface.
func (m *Median) Poke(ctx context.Context, prices []*median.Price, simulateBeforeRun bool) (*types.Hash, error) {
	// It's important to send prices in correct order, otherwise contract will fail:
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Val.Cmp(prices[j].Val) < 0
	})

	// Prepare arguments:
	var (
		val []*big.Int
		age []*big.Int
		v   []uint8
		r   [][32]byte
		s   [][32]byte
	)
	for _, arg := range prices {
		vByte := uint8(arg.Sig.V.Uint64())
		rHash := types.MustHashFromBytes(arg.Sig.R.Bytes(), types.PadLeft)
		sHash := types.MustHashFromBytes(arg.Sig.S.Bytes(), types.PadLeft)
		val = append(val, arg.Val)
		age = append(age, big.NewInt(arg.Age.Unix()))
		v = append(v, vByte)
		r = append(r, rHash)
		s = append(s, sHash)
	}
	args := []any{val, age, v, r, s}

	// Simulate:
	if simulateBeforeRun {
		if err := m.read(ctx, "poke", args, nil); err != nil {
			return nil, err
		}
	}

	// Send transaction:
	return m.write(ctx, "poke", args)
}

// Lift implements the oracle.Median interface.
func (m *Median) Lift(ctx context.Context, addresses []types.Address, simulateBeforeRun bool) (*types.Hash, error) {
	args := []any{addresses}
	if simulateBeforeRun {
		if err := m.read(ctx, "lift", args, nil); err != nil {
			return nil, err
		}
	}
	return m.write(ctx, "lift", args)
}

// Drop implements the oracle.Median interface.
func (m *Median) Drop(ctx context.Context, addresses []types.Address, simulateBeforeRun bool) (*types.Hash, error) {
	args := []any{addresses}
	if simulateBeforeRun {
		if err := m.read(ctx, "drop", args, nil); err != nil {
			return nil, err
		}
	}
	return m.write(ctx, "drop", args)
}

// SetBar implements the oracle.Median interface.
func (m *Median) SetBar(ctx context.Context, bar *big.Int, simulateBeforeRun bool) (*types.Hash, error) {
	args := []any{bar}
	if simulateBeforeRun {
		if err := m.read(ctx, "setBar", args, nil); err != nil {
			return nil, err
		}
	}
	return m.write(ctx, "setBar", args)
}

func (m *Median) read(ctx context.Context, method string, args []any, res []any) error {
	cd, err := medianABI.Methods[method].EncodeArgs(args...)
	if err != nil {
		return err
	}

	var data []byte
	err = retry(maxReadRetries, delayBetweenReadRetries, func() error {
		data, err = m.ethereum.Call(ctx, types.Call{To: &m.address, Input: cd})
		return err
	})
	if err != nil {
		return err
	}

	return medianABI.Methods[method].DecodeValues(data, res...)
}

func (m *Median) write(ctx context.Context, method string, args []any) (*types.Hash, error) {
	cd, err := medianABI.Methods[method].EncodeArgs(args...)
	if err != nil {
		return nil, err
	}

	gl := uint64(gasLimit)
	return m.ethereum.SendTransaction(ctx, &types.Transaction{
		Call: types.Call{
			To:       &m.address,
			Input:    cd,
			GasLimit: &gl,
		},
	})
}

func retry(maxRetries int, delay time.Duration, f func() error) error {
	for i := 0; ; i++ {
		err := f()
		if err != nil {
			return err
		}
		if err == nil || i >= (maxRetries-1) {
			break
		}
		time.Sleep(delay)
	}
	return nil
}
