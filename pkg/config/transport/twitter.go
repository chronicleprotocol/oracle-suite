package transport

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/middleware"
)

type TweetPriceMessage struct {
	*messages.Price
}

func (t *TweetPriceMessage) Tweet() string {
	f, _ := new(big.Float).SetInt(t.Price.Price.Val).Float64()
	f = f / oracle.PriceMultiplier
	if f >= 100 {
		return fmt.Sprintf("%s: %.2f", t.Price.Price.Wat, f)
	}
	return fmt.Sprintf("%s: %f", t.Price.Price.Wat, f)
}

func twitterPriceMiddleware(_ context.Context, next middleware.BroadcastFunc) middleware.BroadcastFunc {
	return func(topic string, msg transport.Message) error {
		if msg, ok := msg.(*messages.Price); ok {
			return next(topic, &TweetPriceMessage{Price: msg})
		}
		return next(topic, msg)
	}
}

type msgLimiterMiddleware struct {
	maximumAge    int
	minimumSpread float64
	prices        map[string]*messages.Price
}

func newMsgLimitMiddleware(maxAge int, minSpread float64) *msgLimiterMiddleware {
	return &msgLimiterMiddleware{
		maximumAge:    maxAge,
		minimumSpread: minSpread,
		prices:        make(map[string]*messages.Price),
	}
}

func (t *msgLimiterMiddleware) Broadcast(_ context.Context, next middleware.BroadcastFunc) middleware.BroadcastFunc {
	return func(topic string, msg transport.Message) error {
		if topic == messages.PriceV0MessageName {
			return nil
		}
		if price, ok := msg.(*messages.Price); ok {
			if price.Price.Age.IsZero() {
				return nil
			}
			prev, ok := t.prices[price.Price.Wat]
			if !ok {
				t.prices[price.Price.Wat] = price
				return next(topic, msg)
			}
			isFresh := time.Since(prev.Price.Age) <= time.Minute*time.Duration(t.maximumAge)
			isSimilar := spread(price.Price.Val, prev.Price.Val) <= t.minimumSpread
			if isFresh && isSimilar {
				return nil
			}
			t.prices[price.Price.Wat] = price
			return next(topic, msg)
		}
		return next(topic, msg)
	}
}

func spread(a, b *big.Int) float64 {
	if a.Sign() == 0 || b.Sign() == 0 {
		return math.Inf(1)
	}

	oldPriceF := new(big.Float).SetInt(a)
	newPriceF := new(big.Float).SetInt(b)

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))
	xf, _ := x.Float64()

	return math.Abs(xf)
}
