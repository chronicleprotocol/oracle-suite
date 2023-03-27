package e2e

import (
	"encoding/json"
	"time"
)

type goferPrice struct {
	Type       string            `json:"type"`
	Base       string            `json:"base"`
	Quote      string            `json:"quote"`
	Price      float64           `json:"price"`
	Bid        float64           `json:"bid"`
	Ask        float64           `json:"ask"`
	Volume24h  float64           `json:"vol24h"`
	Timestamp  time.Time         `json:"ts"`
	Parameters map[string]string `json:"params,omitempty"`
	Prices     []goferPrice      `json:"prices,omitempty"`
	Error      string            `json:"error,omitempty"`
}

func parseGoferPrice(price []byte) (goferPrice, error) {
	var p goferPrice
	err := json.Unmarshal(price, &p)
	if err != nil {
		return p, err
	}
	return p, nil
}
