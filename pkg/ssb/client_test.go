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

package ssb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.cryptoscope.co/muxrpc/v2"
)

func TestClient_WhoAmI(t *testing.T) {
	edp := func(
		funcName string,
		ctx context.Context,
		enc muxrpc.RequestEncoding,
		met muxrpc.Method,
		args ...interface{},
	) ([]byte, error) {
		assert.Equal(t, "Async", funcName)
		assert.NotNil(t, ctx)
		assert.Equal(t, muxrpc.TypeBinary, enc)
		assert.Equal(t, "whoami", met.String())
		assert.Len(t, args, 0)
		return []byte("test"), nil
	}
	c := &Client{
		ctx: context.Background(),
		rpc: testingEndpoint(edp),
	}
	b, err := c.WhoAmI()
	require.NoError(t, err)
	assert.Equal(t, []byte("test"), b)
}

type testingEndpoint func(string, context.Context, muxrpc.RequestEncoding, muxrpc.Method, ...interface{}) ([]byte, error)

func (t testingEndpoint) Async(ctx context.Context, ret interface{}, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) error {
	b, err := t("Async", ctx, enc, met, args...)
	if err != nil {
		return err
	}
	*ret.(*[]byte) = b
	return nil
}
func (t testingEndpoint) Source(ctx context.Context, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) (*muxrpc.ByteSource, error) {
	b, err := t("Source", ctx, enc, met, args...)
	return muxrpc.NewTestSource(b), err
}
