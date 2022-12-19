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

package rpcsplitter

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransport(t *testing.T) {
	rpcMock := &mockClient{t: t}
	roundTripper, err := NewTransport(
		"rpcsplitter-vhost",
		nil,
		withCallers(map[string]caller{"caller": rpcMock}),
		WithRequirements(1, 1),
	)
	if err != nil {
		panic(err)
	}
	httpClient := http.Client{Transport: roundTripper}
	msg := jsonMarshal(t, rpcReq{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "net_version",
		Params:  nil,
	})

	rpcMock.mockCall(1, "net_version")

	res, err := httpClient.Post("http://rpcsplitter-vhost", "application/json", bytes.NewReader(msg))
	defer func() { _ = res.Body.Close() }()

	body, _ := io.ReadAll(res.Body)

	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","id":1,"result":1}`, string(body))
}
