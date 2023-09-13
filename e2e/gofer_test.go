package e2e

import (
	"encoding/json"
	"time"
)

type goferPoint struct {
	Meta      map[string]any `json:"meta"`
	SubPoints []goferPoint   `json:"sub_points,omitempty"`
	Time      time.Time      `json:"time"`
	Value     struct {
		Pair      string  `json:"pair"`
		Price     float64 `json:"price,string"`
		Volume24h string  `json:"volume24h,omitempty"`
	} `json:"value"`
	Error string `json:"error,omitempty"`
}

// e.g. price is
// {"BTC/USD":{"meta":{"type":"reference"},"sub_points":[{"meta":{"min_values":3,"type":"median"},"sub_points":[{"meta":{"expiry_threshold":300000000000,"freshness_threshold":60000000000,"origin":"binance","query":"BTC/USD","type":"origin"},"time":"2023-09-13T12:17:49Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}},{"meta":{"expiry_threshold":300000000000,"freshness_threshold":60000000000,"origin":"bitstamp","query":"BTC/USD","type":"origin"},"time":"2023-09-13T12:17:49.867684Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}},{"error":"data point is expired","meta":{"expiry_threshold":300000000000,"freshness_threshold":60000000000,"origin":"coinbase","query":"BTC/USD","type":"origin"},"time":"2023-09-13T12:17:49.868161Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}},{"meta":{"expiry_threshold":300000000000,"freshness_threshold":60000000000,"origin":"gemini","query":"BTC/USD","type":"origin"},"time":"2023-09-13T12:17:49Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}},{"meta":{"expiry_threshold":300000000000,"freshness_threshold":60000000000,"origin":"kraken","query":"BTC/USD","type":"origin"},"time":"2023-09-13T12:17:50Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}}],"time":"2023-09-13T12:17:49Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}}],"time":"2023-09-13T12:17:49Z","value":{"pair":"BTC/USD","price":"1","volume24h":"0"}}}
func parseGoferPrice(price []byte) (map[string]goferPoint, error) {
	var p map[string]goferPoint

	err := json.Unmarshal(price, &p)
	if err != nil {
		return p, err
	}
	return p, nil
}
