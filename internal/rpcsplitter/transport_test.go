package rpcsplitter

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestTransport(t *testing.T) {
	rpcMock := &mockClient{t: t}
	rpcHandler, _ := newHandlerWithClients([]rpcClient{{rpcCaller: rpcMock}}, null.New())
	roundTripper, _ := newTransport(rpcHandler, "rpcsplitter-vhost", nil)
	httpClient := http.Client{Transport: roundTripper}
	msg := jsonMarshal(t, rpcReq{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "net_version",
		Params:  nil,
	})

	rpcMock.mockCall(1, "net_version")

	res, err := httpClient.Post("http://rpcsplitter-vhost", "application/json", bytes.NewReader(msg))
	body, _ := io.ReadAll(res.Body)

	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","id":1,"result":1}`, string(body))
}
