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

package twitter

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/twitter/internal/api"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/twitter/internal/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/imcoder"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

type ImageType int

const (
	ImageTypeJPEG ImageType = iota
	ImageTypePNG
)

// WithTweet is an interface for messages that contain a Tweet text content to
// be posted along with the mosaic image. Because multiple messages can be posted
// in a single tweet, texts from multiple messages will be separated by a newline.
type WithTweet interface {
	transport.Message
	Tweet() string
}

// Twitter is a transport that uses Twitter as a medium. It can be used to
// broadcast messages and receive messages from other Twitter accounts. The
// message data is encoded into an image (called mosaic) and posted as a tweet.
// The image is then downloaded by other clients and decoded to retrieve the
// message data.
type Twitter struct {
	qmu    sync.Mutex // queue mutex
	ctx    context.Context
	waitCh chan error

	// Values from the Config structure:
	accounts            []string
	topics              map[string]transport.Message
	fetchTweetsInterval time.Duration
	postTweetsInterval  time.Duration
	queueSize           int
	maximumDataSize     int
	maximumTweetLength  int
	mosaicType          ImageType
	mosaicBitsPerChan   uint
	mosaicBlockSize     uint
	log                 log.Logger

	api         *api.API
	id          []byte                                                 // Twitter username.
	accountIDs  []string                                               // Twitter account IDs for accounts from the accounts slice.
	lastTweetID []string                                               // Last tweet ID for each account.
	queue       chan message                                           // Queue for messages to be posted.
	msgCh       map[string]chan transport.ReceivedMessage              // Channels for received messages.
	msgChFO     map[string]*chanutil.FanOut[transport.ReceivedMessage] // Fan-out channels for received messages.
}

type Config struct {
	// Twitter API credentials:
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string

	// Client is an HTTP client to use for Twitter API requests. If nil, the
	// default HTTP client will be used.
	Client http.Client
	// Topics is a list of subscribed topics. A value of the map a type of
	// message given as a nil pointer, e.g.: (*Message)(nil).
	Topics map[string]transport.Message
	// Accounts is a list of Twitter accounts to fetch tweets from.
	Accounts []string
	// PostTweetsInterval is a time interval between posting tweets to Twitter.
	// Transport may post more than one tweet per interval if MaximumDataSize
	// or MaximumTweetLength is reached.
	PostTweetsInterval time.Duration
	// FetchTweetsInterval is a time interval between fetching tweets from
	// Twitter accounts.
	FetchTweetsInterval time.Duration
	// QueueSize is a size of the queue for messages. After reaching the limit
	// the transport will block until the queue is not empty.
	QueueSize uint
	// MaximumDataSize is a maximum size of the data to be sent in a single
	// mosaic image. If the data is larger than MaximumDataSize, the transport
	// will split it into multiple tweets.
	MaximumDataSize uint
	// MaximumTweetLength is a maximum length of the tweet. If the tweet is
	// larger than MaximumTweetLength, the transport will split it into
	// multiple tweets.
	MaximumTweetLength uint
	// MosaicType is a type of the mosaic image. Default is ImageTypeJPEG.
	MosaicType ImageType
	// MosaicBitsPerChannel is a number of bits per channel in the mosaic image.
	MosaicBitsPerChannel uint
	// MosaicBlockSize is a size of the block in the mosaic image.
	MosaicBlockSize uint
	// Logger is a custom logger instance. If not provided then null
	// logger is used.
	Logger log.Logger
}

func New(cfg Config) (*Twitter, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return &Twitter{
		waitCh: make(chan error),

		api: &api.API{
			Signer: &api.OAuth{
				ConsumerKey:    cfg.ConsumerKey,
				ConsumerSecret: cfg.ConsumerSecret,
				AccessToken:    cfg.AccessToken,
				AccessSecret:   cfg.AccessSecret,
			},
			Client: cfg.Client,
		},

		accounts:            sliceutil.Copy(cfg.Accounts),
		topics:              maputil.Copy(cfg.Topics),
		fetchTweetsInterval: cfg.FetchTweetsInterval,
		postTweetsInterval:  cfg.PostTweetsInterval,
		queueSize:           int(cfg.QueueSize),
		maximumDataSize:     int(cfg.MaximumDataSize),
		maximumTweetLength:  int(cfg.MaximumTweetLength),
		mosaicType:          cfg.MosaicType,
		mosaicBitsPerChan:   cfg.MosaicBitsPerChannel,
		mosaicBlockSize:     cfg.MosaicBlockSize,
		log:                 cfg.Logger,

		accountIDs:  make([]string, len(cfg.Accounts)),
		lastTweetID: make([]string, len(cfg.Accounts)),
		queue:       make(chan message, cfg.QueueSize),
		msgCh:       make(map[string]chan transport.ReceivedMessage),
		msgChFO:     make(map[string]*chanutil.FanOut[transport.ReceivedMessage]),
	}, nil
}

// ID implements transport.Transport interface.
func (t *Twitter) ID() []byte {
	return t.id
}

// Broadcast implements transport.Transport interface.
func (t *Twitter) Broadcast(topic string, msg transport.Message) error {
	if _, ok := t.topics[topic]; !ok {
		return fmt.Errorf("topic %q is not subscribed", topic)
	}
	data, err := msg.MarshallBinary()
	if err != nil {
		return fmt.Errorf("unable to broadcast message: %w", err)
	}
	var tweet string
	if t, ok := msg.(WithTweet); ok {
		tweet = t.Tweet()
	}
	t.qmu.Lock()
	defer t.qmu.Unlock()
	t.queue <- message{
		tweet: tweet,
		topic: topic,
		data:  data,
	}
	return nil
}

// Messages implements transport.Transport interface.
func (t *Twitter) Messages(topic string) <-chan transport.ReceivedMessage {
	if ch, ok := t.msgChFO[topic]; ok {
		return ch.Chan()
	}
	return nil
}

// Start implements transport.Transport interface.
func (t *Twitter) Start(ctx context.Context) error {
	if t.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	t.ctx = ctx
	// Get the Twitter username. The purpose of this code is mostly to check
	// if the API credentials are valid.
	me, err := t.api.Me(ctx)
	if err != nil {
		return fmt.Errorf("twitter transport: unable to start transport: %w", err)
	}
	t.id = []byte(me.Data.Username)
	// Get accounts IDs for accounts for followed accounts. Twitter API allows
	// to fetch tweets only by using account ID.
	for i, account := range t.accounts {
		user, err := t.api.UserByUsername(
			ctx,
			&api.UserByUsernameQuery{Username: strings.TrimLeft(account, "@")},
		)
		if err != nil {
			t.log.
				WithField("account", account).
				Warn("unable to get account ID")
			continue
		}
		t.accountIDs[i] = user.Data.ID
	}
	for topic := range t.topics {
		t.msgCh[topic] = make(chan transport.ReceivedMessage)
		t.msgChFO[topic] = chanutil.NewFanOut(t.msgCh[topic])
	}
	go t.contextCancelHandler()
	go t.postTweetsRoutine()
	go t.fetchTweetsRoutine()
	return nil
}

// Wait implements transport.Transport interface.
func (t *Twitter) Wait() <-chan error {
	return t.waitCh
}

// postTweet posts a tweet to Twitter along with the mosaic image.
func (t *Twitter) postTweet(tweet []string, m *mosaic) error {
	// Encode data to image.
	img, err := m.encode(imcoder.Options{
		BlockSize:   t.mosaicBlockSize,
		BitsPerChan: t.mosaicBitsPerChan,
	}, t.mosaicType)
	// Upload image to Twitter.
	media, err := t.api.Upload(t.ctx, bytes.NewReader(img))
	if err != nil {
		return fmt.Errorf("twitter transport: unable to broadcast messages: %w", err)
	}
	// Create tweet with encoded data in attached image.
	if _, err = t.api.CreateTweet(t.ctx, &api.CreateTweetRequest{
		Text:  strings.Join(tweet, "\n"),
		Media: &api.CreateTweetMediaRequest{MediaIDs: []string{media.MediaIDString}},
	}); err != nil {
		return fmt.Errorf("twitter transport: unable to broadcast messages: %w", err)
	}
	return nil
}

// handleMosaic handles mosaic data from the URL.
func (t *Twitter) handleMosaic(url string) (*mosaic, error) {
	// Download image.
	res, err := t.api.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// Decode mosaic image.
	m := &mosaic{}
	if err := m.decode(body); err != nil {
		return nil, fmt.Errorf("twitter transport: unable to decode image: %w", err)
	}
	return m, nil
}

func (t *Twitter) postTweetsRoutine() {
	ticker := time.NewTicker(t.postTweetsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			// Drain queue and group messages by topic. Queue mutex is locked until
			// all messages are drained from the queue.
			tm := map[string][]message{}
			func() {
				t.qmu.Lock()
				defer t.qmu.Unlock()
				for len(t.queue) > 0 {
					msg := <-t.queue
					tm[msg.topic] = append(tm[msg.topic], msg)
				}
			}()
			for topic, msgs := range tm {
				var (
					tweets   []string // Tweet content. Every element of the slice is new line.
					dataLen  int      // Total length of all messages.
					tweetLen int      // Length of current tweet.
				)
				m := &mosaic{topic: topic}
				// Iterate over messages and group them into tweets. The new
				// tweet is posted every time the tweet length or the data
				// length exceeds the maximum allowed values defined in the
				// maximumDataSize and maximumTweetSize fields.
				for _, msg := range msgs {
					nextDataLen := dataLen + len(msg.data)
					nextTweetLen := tweetLen + len(msg.tweet)
					// Post tweet if the next message will exceed the maximum
					// allowed lengths for the tweet or the data.
					if len(m.data) > 0 && (nextDataLen > t.maximumDataSize || nextTweetLen > t.maximumTweetLength) {
						if err := t.postTweet(tweets, m); err != nil {
							t.log.
								WithError(err).
								Warn("unable to post tweet")
						}
						tweets = nil
						dataLen = 0
						tweetLen = 0
						m.data = nil
					}
					if len(msg.tweet) > 0 {
						tweets = append(tweets, msg.tweet)
					}
					dataLen = nextDataLen
					tweetLen = nextTweetLen
					m.data = append(m.data, msg.data)
				}
				// Post leftover messages if any left.
				if len(m.data) > 0 {
					if err := t.postTweet(tweets, m); err != nil {
						t.log.
							WithError(err).
							Warn("unable to post tweet")
					}
				}
			}
		}
	}
}

func (t *Twitter) fetchTweetsRoutine() {
	ticker := time.NewTicker(t.fetchTweetsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			for i, accountID := range t.accountIDs {
				// Fetch tweets from the account since the last known tweet ID.
				// During the first iteration the last known tweet ID is 0, in
				// that case the Twitter API will return up to 3200 tweets most
				// recent tweets.
				tweets, err := t.api.UserTweets(t.ctx, &api.UserTweetsQuery{
					UserID:      accountID,
					SinceID:     t.lastTweetID[i],
					Expansions:  "attachments.media_keys",
					MediaFields: "url",
				})
				if err != nil {
					t.log.
						WithError(err).
						Warn("unable to fetch tweets")
					continue
				}
				// Messages are returned from the newest to oldest. We store
				// the ID of the newest tweet. That ID is then used as a
				// starting point for the next fetch.
				if len(tweets.Data) > 0 {
					t.lastTweetID[i] = tweets.Data[0].ID
				}
				// Iterate over mosaics from the tweets.
				for _, media := range tweets.Includes.Media {
					m, err := t.handleMosaic(media.URL)
					if err != nil {
						t.log.
							WithError(err).
							Warn("unable to handle mosaic")
						continue
					}
					typ, ok := t.topics[m.topic]
					if !ok {
						continue
					}
					// Iterate over all messages in the mosaic, unmarshal them into
					// the appropriate type and send them to the msgCh channel.
					tr := reflect.TypeOf(typ).Elem()
					for _, bin := range m.data {
						msg := reflect.New(tr).Interface().(transport.Message)
						if err := msg.UnmarshallBinary(bin); err != nil {
							t.log.
								WithError(err).
								Warn("unable to unmarshal message from mosaic")
							continue
						}
						if _, ok := t.msgCh[m.topic]; !ok {
							// This should never happen.
							t.log.
								WithField("topic", m.topic).
								Panic("topic channel is not initialized")
						}
						t.msgCh[m.topic] <- transport.ReceivedMessage{
							Message: msg,
							Author:  []byte(t.accounts[i]),
						}
					}
				}
			}
		}
	}
}

// contextCancelHandler handles context cancellation.
func (t *Twitter) contextCancelHandler() {
	defer func() { close(t.waitCh) }()
	<-t.ctx.Done()
}

// message represents a message to be posted on Twitter. One Tweet may contain
// multiple messages.
type message struct {
	tweet string
	topic string
	data  []byte
}

// mosaic is a mosaic image that is posted along with a Tweet. It contains a topic
// and a list of messages encoded in binary format. The mosaic structure can be
// encoded and decoded to and from a mosaic image.
type mosaic struct {
	topic string
	data  [][]byte
}

// encode encodes the mosaic structure into a mosaic image.
func (c *mosaic) encode(opts imcoder.Options, typ ImageType) ([]byte, error) {
	// Step 1: Encode data to Protobuf.
	pbc := &pb.Container{
		Topic:    c.topic,
		Messages: c.data,
	}
	bin, err := proto.Marshal(pbc)
	if err != nil {
		return nil, err
	}
	// Step 2: Compress data.
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	defer zw.Close()
	if _, err := zw.Write(bin); err != nil {
		return nil, err
	}
	// Step 3: Encode compressed data to an image.
	im, err := imcoder.Encode(compressed.Bytes(), opts)
	if err != nil {
		return nil, err
	}
	// Step 4: Encode image.
	var img bytes.Buffer
	switch typ {
	case ImageTypePNG:
		if err := png.Encode(&img, im); err != nil {
			return nil, err
		}
	case ImageTypeJPEG:
		if err := jpeg.Encode(&img, im, &jpeg.Options{Quality: 100}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown image type: %v", typ)
	}
	return img.Bytes(), nil
}

// decode decodes a mosaic image into a mosaic structure.
func (c *mosaic) decode(data []byte) error {
	var err error
	// Step 1: Decode image.
	var img image.Image
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	switch format {
	case "png":
		img, err = png.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
	case "jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown image type: %v", format)
	}
	// Step 2: Decode image data.
	compressed, err := imcoder.Decode(img)
	if err != nil {
		return err
	}
	// Step 3: Decompress data.
	var bin bytes.Buffer
	zr, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	defer zr.Close()
	if _, err := io.Copy(&bin, zr); err != nil {
		return err
	}
	// Step 4: Decode Protobuf.
	pbc := &pb.Container{}
	if err := proto.Unmarshal(bin.Bytes(), pbc); err != nil {
		return err
	}
	c.topic = pbc.Topic
	c.data = pbc.Messages
	return nil
}
