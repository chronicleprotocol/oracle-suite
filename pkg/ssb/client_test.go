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
	"go.cryptoscope.co/ssb/plugins/legacyinvites"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.cryptoscope.co/muxrpc/v2"
	"go.cryptoscope.co/ssb/message"
)

func TestClient_WhoAmI(t *testing.T) {
	rpc := &muxrpc.FakeEndpoint{
		AsyncStub: func(ctx context.Context, ret interface{}, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) error {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "whoami", met.String())
			assert.Len(t, args, 0)
			*ret.(*[]byte) = []byte("IamMe")
			return nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	b, err := c.WhoAmI()

	require.NoError(t, err)
	assert.Equal(t, []byte("IamMe"), b)

	assert.Equal(t, 1, rpc.AsyncCallCount())
	assert.Equal(t, 0, rpc.SourceCallCount())
}

func TestClient_Transmit(t *testing.T) {
	rpc := &muxrpc.FakeEndpoint{
		AsyncStub: func(ctx context.Context, ret interface{}, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) error {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "publish", met.String())
			assert.Len(t, args, 1)
			assert.IsType(t, "", args[0])
			*ret.(*[]byte) = []byte("response")
			return nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	b, err := c.Transmit("message")

	require.NoError(t, err)
	assert.Equal(t, []byte("response"), b)

	assert.Equal(t, 1, rpc.AsyncCallCount())
	assert.Equal(t, 0, rpc.SourceCallCount())
}

func TestClient_ReceiveLast(t *testing.T) {
	s := [][]byte{
		[]byte(`{"value":{"content":{"type":"xxx"}}}`),
		[]byte(`{"value":{"content":{"type":"yyy"}}}`),
		[]byte(`{"value":{"content":{"type":"zzz"}}}`),
	}
	rpc := &muxrpc.FakeEndpoint{
		SourceStub: func(ctx context.Context, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) (*muxrpc.ByteSource, error) {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "createUserStream", met.String())
			assert.Len(t, args, 1)
			assert.IsType(t, message.CreateHistArgs{}, args[0])
			return muxrpc.NewTestSource(s...), nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	var err error

	_, err = c.ReceiveLast("", "", 1)
	assert.EqualError(t, err, "ssb: feedRef empty")

	_, err = c.ReceiveLast("abcd", "", 1)
	assert.EqualError(t, err, "ssb: Invalid Ref")

	_, err = c.ReceiveLast("abcd.xyz", "", 1)
	assert.EqualError(t, err, `feedRef: couldn't parse "abcd.xyz": illegal base64 data at input byte 0: ssb: Invalid Hash`)

	_, err = c.ReceiveLast("#KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "", 1)
	assert.EqualError(t, err, "ssb: Invalid Ref Type")

	_, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.xxx", "", 1)
	assert.EqualError(t, err, "unhandled feed algorithm: @KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.xxx: ssb: Invalid Ref Algo")

	_, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCcMIkCc/w==.ed25519", "", 1)
	assert.EqualError(t, err, "ssb: Invalid reference len for ed25519: 37")

	assert.Equal(t, 0, rpc.SourceCallCount())

	var b []byte

	b, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "", 1)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte(`{"value":{"content":{"type":"xxx"}}}`), b)
	}

	b, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "xxx", 1)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte(`{"value":{"content":{"type":"xxx"}}}`), b)
	}

	b, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "yyy", 1)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte(`{"value":{"content":{"type":"yyy"}}}`), b)
	}

	b, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "zzz", 1)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte(`{"value":{"content":{"type":"zzz"}}}`), b)
	}

	assert.Equal(t, 4, rpc.SourceCallCount())
	assert.Equal(t, 0, rpc.AsyncCallCount())
}

func TestClient_ReceiveLast_NoData(t *testing.T) {
	rpc := &muxrpc.FakeEndpoint{
		SourceStub: func(ctx context.Context, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) (*muxrpc.ByteSource, error) {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "createUserStream", met.String())
			assert.Len(t, args, 1)
			assert.IsType(t, message.CreateHistArgs{}, args[0])
			return muxrpc.NewTestSource(), nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	var err error

	_, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "", 1)
	assert.EqualError(t, err, "no data in the stream for ref: @KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519")

	assert.Equal(t, 1, rpc.SourceCallCount())
	assert.Equal(t, 0, rpc.AsyncCallCount())
}

func TestClient_ReceiveLast_MissingData(t *testing.T) {
	s := [][]byte{
		[]byte(`{"value":{"content":{"type":"xxx"}}}`),
		[]byte(`{"value":{"content":{"type":"yyy"}}}`),
		[]byte(`{"value":{"content":{"type":"zzz"}}}`),
	}

	rpc := &muxrpc.FakeEndpoint{
		SourceStub: func(ctx context.Context, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) (*muxrpc.ByteSource, error) {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "createUserStream", met.String())
			assert.Len(t, args, 1)
			assert.IsType(t, message.CreateHistArgs{}, args[0])
			return muxrpc.NewTestSource(s...), nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	var err error

	_, err = c.ReceiveLast("@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519", "aaa", 1)

	assert.EqualError(t, err, "no data of type aaa in the stream for ref: @KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519")

	assert.Equal(t, 1, rpc.SourceCallCount())
	assert.Equal(t, 0, rpc.AsyncCallCount())
}

func TestClient_LogStream(t *testing.T) {
	rpc := &muxrpc.FakeEndpoint{
		SourceStub: func(ctx context.Context, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) (*muxrpc.ByteSource, error) {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "createLogStream", met.String())
			assert.Len(t, args, 1)
			assert.IsType(t, message.CreateLogArgs{}, args[0])
			return nil, nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	ch, err := c.LogStream()
	assert.IsType(t, make(chan []byte), ch)
	assert.NoError(t, err)

	assert.Equal(t, 1, rpc.SourceCallCount())
	assert.Equal(t, 0, rpc.AsyncCallCount())
}

func TestClient_InviteCreate(t *testing.T) {
	rpc := &muxrpc.FakeEndpoint{
		AsyncStub: func(ctx context.Context, ret interface{}, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) error {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "invite.create", met.String())
			assert.Len(t, args, 1)
			assert.IsType(t, legacyinvites.CreateArguments{}, args[0])
			*ret.(*[]byte) = []byte("SomeToken")
			return nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	b, err := c.InviteCreate(10)

	require.NoError(t, err)
	assert.Equal(t, []byte("SomeToken"), b)

	assert.Equal(t, 1, rpc.AsyncCallCount())
	assert.Equal(t, 0, rpc.SourceCallCount())
}

func TestClient_InviteAccept(t *testing.T) {
	rpc := &muxrpc.FakeEndpoint{
		AsyncStub: func(ctx context.Context, ret interface{}, enc muxrpc.RequestEncoding, met muxrpc.Method, args ...interface{}) error {
			assert.NotNil(t, ctx)
			assert.Equal(t, muxrpc.TypeBinary, enc)
			assert.Equal(t, "invite.accept", met.String())
			assert.Len(t, args, 1)
			*ret.(*[]byte) = []byte("OK")
			return nil
		},
	}
	c := &Client{
		ctx: context.Background(),
		rpc: rpc,
	}

	b, err := c.InviteAccept("TokenToAccept")

	require.NoError(t, err)
	assert.Equal(t, []byte("OK"), b)

	assert.Equal(t, 1, rpc.AsyncCallCount())
	assert.Equal(t, 0, rpc.SourceCallCount())
}
