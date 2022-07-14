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

package geth

import (
	"context"
	"errors"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
)

var ErrStorageQueryFailed = errors.New("oracle contract storage query failed")

// TODO: make it configurable
const gasLimit = 200000
const maxReadRetries = 3
const delayBetweenReadRetries = 5 * time.Second

// Median implements the oracle.Median interface using go-ethereum packages.
type Median struct {
	ethereum ethereum.Client
	address  ethereum.Address
}

// NewMedian creates the new Median instance.
func NewMedian(ethereum ethereum.Client, address ethereum.Address) *Median {
	return &Median{
		ethereum: ethereum,
		address:  address,
	}
}

// Address implements the oracle.Median interface.
func (m *Median) Address() common.Address {
	return m.address
}

// Age implements the oracle.Median interface.
func (m *Median) Age(ctx context.Context) (time.Time, error) {
	r, err := m.read(ctx, "age")
	if err != nil {
		return time.Unix(0, 0), err
	}

	return time.Unix(int64(r[0].(uint32)), 0), nil
}

// Bar implements the oracle.Median interface.
func (m *Median) Bar(ctx context.Context) (int64, error) {
	r, err := m.read(ctx, "bar")
	if err != nil {
		return 0, err
	}

	return r[0].(*big.Int).Int64(), nil
}

// Wat implements the oracle.Median interface.
func (m *Median) Wat(ctx context.Context) (string, error) {
	r, err := m.read(ctx, "wat")
	if err != nil {
		return "", err
	}

	b := r[0].([32]byte)
	return string(b[:]), nil
}

// Val implements the oracle.Median interface.
func (m *Median) Val(ctx context.Context) (*big.Int, error) {
	const (
		offset = 16
		length = 16
	)

	b, err := m.ethereum.Storage(ctx, m.address, common.BigToHash(big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	if len(b) < (offset + length) {
		return nil, ErrStorageQueryFailed
	}

	return new(big.Int).SetBytes(b[length : offset+length]), err
}

// Feeds implements the oracle.Median interface.
func (m *Median) Feeds(ctx context.Context) ([]ethereum.Address, error) {
	var (
		err   error
		null  ethereum.Address
		orcl  []ethereum.Address
		calls []ethereum.Call
	)

	// Prepare the call list:
	for i := 0; i < 256; i++ {
		cd, err := medianABI.Pack("slot", uint8(i))
		if err != nil {
			return nil, err
		}
		calls = append(calls, ethereum.Call{
			Address: m.address,
			Data:    cd,
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
		addr, err := medianABI.Unpack("slot", data)
		if err != nil {
			return nil, err
		}
		if len(addr) == 1 && addr[0] != null {
			orcl = append(orcl, addr[0].(common.Address))
		}
	}

	return orcl, nil
}

// Poke implements the oracle.Median interface.
func (m *Median) Poke(ctx context.Context, prices []*oracle.Price, simulateBeforeRun bool) (*ethereum.Hash, error) {
	// It's important to send prices in correct order, otherwise contract will fail:
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Val.Cmp(prices[j].Val) < 0
	})

	var (
		val []*big.Int
		age []*big.Int
		v   []uint8
		r   [][32]byte
		s   [][32]byte
	)

	for _, arg := range prices {
		val = append(val, arg.Val)
		age = append(age, big.NewInt(arg.Age.Unix()))
		v = append(v, arg.V)
		r = append(r, arg.R)
		s = append(s, arg.S)
	}

	if simulateBeforeRun {
		if _, err := m.read(ctx, "poke", val, age, v, r, s); err != nil {
			return nil, err
		}
	}

	return m.write(ctx, "poke", val, age, v, r, s)
}

// Lift implements the oracle.Median interface.
func (m *Median) Lift(ctx context.Context, addresses []common.Address, simulateBeforeRun bool) (*ethereum.Hash, error) {
	if simulateBeforeRun {
		if _, err := m.read(ctx, "lift", addresses); err != nil {
			return nil, err
		}
	}

	return m.write(ctx, "lift", addresses)
}

// Drop implements the oracle.Median interface.
func (m *Median) Drop(ctx context.Context, addresses []common.Address, simulateBeforeRun bool) (*ethereum.Hash, error) {
	if simulateBeforeRun {
		if _, err := m.read(ctx, "drop", addresses); err != nil {
			return nil, err
		}
	}

	return m.write(ctx, "drop", addresses)
}

// SetBar implements the oracle.Median interface.
func (m *Median) SetBar(ctx context.Context, bar *big.Int, simulateBeforeRun bool) (*ethereum.Hash, error) {
	if simulateBeforeRun {
		if _, err := m.read(ctx, "setBar", bar); err != nil {
			return nil, err
		}
	}

	return m.write(ctx, "setBar", bar)
}

func (m *Median) read(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	cd, err := medianABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	var data []byte
	err = retry(maxReadRetries, delayBetweenReadRetries, func() error {
		data, err = m.ethereum.Call(ctx, ethereum.Call{Address: m.address, Data: cd})
		return err
	})
	if err != nil {
		return nil, err
	}

	return medianABI.Unpack(method, data)
}

func (m *Median) write(ctx context.Context, method string, args ...interface{}) (*ethereum.Hash, error) {
	cd, err := medianABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	return m.ethereum.SendTransaction(ctx, &ethereum.Transaction{
		Address:  m.address,
		GasLimit: new(big.Int).SetUint64(gasLimit),
		Data:     cd,
	})
}

func retry(maxRetries int, delay time.Duration, f func() error) error {
	for i := 0; ; i++ {
		err := f()
		if err != nil {
			return err
		}
		if i >= (maxRetries - 1) {
			break
		}
		time.Sleep(delay)
	}
	return nil
}
