package tor

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/tor/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type producer struct {
	mu     sync.Mutex
	ctx    context.Context
	waitCh chan error

	urlSigner     *urlSigner
	messageSigner *messageSigner
	client        *http.Client
	consumers     Consumers
	flushInterval *timeutil.Ticker
	log           log.Logger

	queue map[string][][]byte
}

type producerConfig struct {
	URLSigner     *urlSigner
	MessageSigner *messageSigner
	Client        *http.Client
	Consumers     Consumers
	FlushInterval *timeutil.Ticker
	Logger        log.Logger
}

func newProducer(cfg producerConfig) *producer {
	return &producer{
		waitCh:        make(chan error),
		urlSigner:     cfg.URLSigner,
		messageSigner: cfg.MessageSigner,
		client:        cfg.Client,
		consumers:     cfg.Consumers,
		flushInterval: cfg.FlushInterval,
		log:           cfg.Logger,
		queue:         make(map[string][][]byte),
	}
}

func (p *producer) Start(ctx context.Context) error {
	if p.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	p.ctx = ctx
	p.flushInterval.Start(ctx)
	go p.flushRoutine(ctx)
	go p.contextCancelHandler()
	return nil
}

func (p *producer) Wait() <-chan error {
	return p.waitCh
}

func (p *producer) Send(topic string, msg []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.queue[topic]; !ok {
		p.queue[topic] = make([][]byte, 0)
	}
	p.queue[topic] = append(p.queue[topic], msg)
	return nil
}

func (p *producer) flush(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	mp := &pb.MessagePack{
		Messages: make(map[string]*pb.MessagePack_Messages, 0),
	}
	for topic, msgs := range p.queue {
		m := &pb.MessagePack_Messages{}
		for _, msg := range msgs {
			m.Data = append(m.Data, msg)
		}
		mp.Messages[topic] = m
	}
	if err := p.messageSigner.SignMessage(mp); err != nil {
		return err
	}
	bin, err := proto.Marshal(mp)
	if err != nil {
		return err
	}
	bin, err = gzipCompress(bin)
	if err != nil {
		return err
	}
	consumers, err := p.consumers.Consumers(ctx)
	if err != nil {
		return err
	}
	for _, addr := range consumers {
		go p.flushOne(ctx, addr, bin)
	}
	return nil
}

func (p *producer) flushOne(ctx context.Context, addr string, data []byte) {
	p.log.
		WithField("addr", addr).
		Info("Sending messages to consumer")

	url, err := p.urlSigner.SignURL(fmt.Sprintf("%s%s", strings.TrimRight(addr, "/"), consumePath))
	if err != nil {
		p.log.
			WithError(err).
			WithField("addr", addr).
			Error("Failed to sign URL")
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	req.WithContext(ctx)
	if err != nil {
		p.log.
			WithField("addr", addr).
			WithError(err).
			Error("Failed to create request")
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")
	_, err = p.client.Do(req)
	if err != nil {
		p.log.
			WithField("addr", addr).
			WithError(err).
			Error("Failed to send messages to consumer")
	}
}

func (p *producer) flushRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.flushInterval.TickCh():
			if err := p.flush(ctx); err != nil {
				p.log.
					WithError(err).
					Error("Failed to send messages")
			}
		}
	}
}

// contextCancelHandler handles context cancellation.
func (p *producer) contextCancelHandler() {
	defer func() { close(p.waitCh) }()
	<-p.ctx.Done()
}

func gzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	g := gzip.NewWriter(&b)
	_, err := g.Write(data)
	if err != nil {
		return nil, err
	}
	err = g.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
