package webapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/httpserver"
	logMocks "github.com/chronicleprotocol/oracle-suite/pkg/log/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/webapi/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type message struct {
	data []byte
}

func (m *message) MarshallBinary() ([]byte, error) {
	return m.data, nil
}

func (m *message) UnmarshallBinary(i []byte) error {
	m.data = i
	return nil
}

type addressBook struct {
	addresses []string
}

func (a *addressBook) Consumers(ctx context.Context) ([]string, error) {
	return a.addresses, nil
}

func Test_WebAPI(t *testing.T) {
	address1 := ethereum.HexToAddress("0x1234567890123456789012345678901234567890")
	address2 := ethereum.HexToAddress("0x2345678901234567890123456789012345678901")

	tests := []struct {
		test func(T *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI)
	}{
		{
			// Single valid message.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now()
				ch := c.Messages("test")

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return(&address1, nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), msgSig).Return(&address1, nil).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for message and verify:
				msg := <-ch
				assert.Equal(t, []byte("data"), msg.Message.(*message).data)
				assert.Equal(t, address1.Bytes(), msg.Author)
				assert.Nil(t, msg.Error)
			},
		},
		{
			// Invalid URL signature.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now()

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return((*ethereum.Address)(nil), errors.New("err")).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Debug" && m.Arguments[0].([]any)[0] == "Invalid request signature" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Feeder not allowed to broadcast.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now()

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return(&address2, nil).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Debug" && m.Arguments[0].([]any)[0] == "Feeder not allowed to send messages" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Message too old.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now().Add(-time.Minute * 2)

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return(&address1, nil).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Warn" && m.Arguments[0].([]any)[0] == "Timestamp too far in the past" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Message timestamp in the future.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now().Add(time.Minute)

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return(&address1, nil).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Warn" && m.Arguments[0].([]any)[0] == "Timestamp too far in the future" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Message received too soon.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm1 := time.Now().Add(-time.Second * 45)
				tm2 := time.Now()
				ch := c.Messages("test")

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig1 := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm1.Unix()))
				urlSig2 := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm2.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig1).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig2).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig1).Return(&address1, nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig2).Return(&address1, nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), msgSig).Return(&address1, nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), msgSig).Return(&address1, nil).Once()

				// Send first message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm1)

				// Wait for first message to be sure, that second message will
				// be sent in a separate batch:
				<-ch

				// Send second message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm2)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Warn" && m.Arguments[0].([]any)[0] == "Too many messages received in a short time" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Invalid message signature.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now()

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return(&address1, nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), msgSig).Return((*ethereum.Address)(nil), errors.New("err")).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Warn" && m.Arguments[0].([]any)[0] == "Invalid message pack signature" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Request author does not match message author.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				tm := time.Now()

				// Prepare mocks:
				msgSig := []byte("testdata")
				urlSig := []byte(fmt.Sprintf("%d30313233343536373839616263646566", tm.Unix()))
				s.On("Signature", msgSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Signature", urlSig).Return(ethereum.SignatureFromBytes([]byte("signature")), nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), urlSig).Return(&address1, nil).Once()
				s.On("Recover", ethereum.SignatureFromBytes([]byte("signature")), msgSig).Return(&address2, nil).Once()

				// Send message:
				require.NoError(t, p.Broadcast("test", &message{data: []byte("data")}))
				p.flushTicker.TickAt(tm)

				// Wait for error log:
				assert.Eventually(t, func() bool {
					for _, m := range l.Mock().Calls {
						if m.Method == "Warn" && m.Arguments[0].([]any)[0] == "Message pack author does not match request author" {
							return true
						}
					}
					return false
				}, time.Second, time.Millisecond*100)
			},
		},
		{
			// Channel must be closed after context is done.
			test: func(t *testing.T, l *logMocks.Logger, s *mocks.Signer, p, c *WebAPI) {
				ch := c.Messages("test")
				_, ok := <-ch
				assert.False(t, ok)
			},
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
			defer ctxCancel()

			// Dependencies:
			consSrv := httpserver.New(&http.Server{Addr: "127.0.0.1:0"})
			consTick := timeutil.NewTicker(60 * time.Second)
			prodTick := timeutil.NewTicker(60 * time.Second)
			rand := bytes.NewReader([]byte(strings.Repeat("0123456789abcdef", 32)))
			ab := &addressBook{addresses: []string{}}
			signer := &mocks.Signer{}
			logger := logMocks.New()
			logger.Mock().On("WithError", mock.Anything).Return(logger)
			logger.Mock().On("WithField", mock.Anything, mock.Anything).Return(logger)
			logger.Mock().On("WithFields", mock.Anything).Return(logger)
			logger.Mock().On("Warn", mock.Anything).Return()
			logger.Mock().On("Info", mock.Anything)
			logger.Mock().On("Debug", mock.Anything)

			// WebAPI instance for a consumer.
			prod, err := New(Config{
				ListenAddr:      "127.0.0.1:0",
				Topics:          map[string]transport.Message{"test": (*message)(nil)},
				AuthorAllowlist: []common.Address{address1},
				AddressBook:     ab,
				Signer:          signer,
				Timeout:         0,
				FlushTicker:     prodTick,
				Rand:            rand,
				Logger:          logger,
			})
			require.NoError(t, err)

			// WebAPI instance for a producer.
			cons, err := New(Config{
				Topics:          map[string]transport.Message{"test": (*message)(nil)},
				AuthorAllowlist: []common.Address{address1},
				AddressBook:     ab,
				Signer:          signer,
				Timeout:         0,
				FlushTicker:     consTick,
				Server:          consSrv,
				Rand:            rand,
				Logger:          logger,
			})
			require.NoError(t, err)

			// Start transport.
			require.NoError(t, prod.Start(ctx))
			require.NoError(t, cons.Start(ctx))
			ab.addresses = []string{"http://" + consSrv.Addr().String()}

			// Run test.
			tt.test(t, logger, signer, prod, cons)

			// Consumer receives a message from the producer before the HTTP
			// request is finished. Because of this, it may happen that the
			// test will be finished before the HTTP request is finished.
			// In that case, context will be canceled and the HTTP request
			// will be interrupted. To avoid this error in tests, we add
			// a small delay.
			//
			// On production, this is not a problem because we actually want
			// to cancel the HTTP request when the context is canceled.
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func Test_signMessage(t *testing.T) {
	// Prepare mocks:
	mp := &pb.MessagePack{
		Messages: map[string]*pb.MessagePack_Messages{
			"a": {
				Data: [][]byte{
					[]byte("1"),
					[]byte("2"),
					[]byte("3"),
				},
			},
			"b": {
				Data: [][]byte{
					[]byte("1"),
					[]byte("2"),
					[]byte("3"),
				},
			},
		},
	}
	expSig := "a123b123"
	signature := ethereum.SignatureFromBytes([]byte("signature"))
	address := ethereum.HexToAddress("0x1234567890123456789012345678901234567890")
	signer := &mocks.Signer{}
	signer.On("Signature", []byte(expSig)).Return(signature, nil)
	signer.On("Recover", signature, []byte(expSig)).Return(&address, nil)

	// Sign message:
	err := signMessage(mp, signer)
	require.NoError(t, err)

	// Check signature:
	retAddress, err := verifyMessage(mp, signer)
	require.NoError(t, err)
	assert.Equal(t, &address, retAddress)
}

func Test_signURL(t *testing.T) {
	// Prepare mocks:
	tm := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	rand := "0123456789ab"
	url := "https://test.onion/consume"
	expSig := "157783680030313233343536373839616200000000"
	expURL := "https://test.onion/consume?t=1577836800&r=30313233343536373839616200000000&s=7369676e61747572650000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	signature := ethereum.SignatureFromBytes([]byte("signature"))
	address := ethereum.HexToAddress("0x1234567890123456789012345678901234567890")
	signer := &mocks.Signer{}
	signer.On("Signature", []byte(expSig)).Return(signature, nil)
	signer.On("Recover", signature, []byte(expSig)).Return(&address, nil)

	// Sign URL:
	got, err := signURL(url, tm, signer, bytes.NewReader([]byte(rand)))
	require.NoError(t, err)
	assert.Equal(t, expURL, got)

	// Check signature:
	retAddress, retTime, err := verifyURL(got, signer)
	require.NoError(t, err)
	assert.Equal(t, &address, retAddress)
	assert.Equal(t, tm.Unix(), retTime.Unix())
}
