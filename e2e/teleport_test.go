package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"time"
)

type lairResponse []struct {
	Timestamp  int                              `json:"timestamp"`
	Data       map[string]string                `json:"data"`
	Signatures map[string]lairResponseSignature `json:"signatures"`
}

type lairResponseSignature struct {
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
}

func waitForLair(ctx context.Context, url string, responses int) (lairResponse, error) {
	lairResponse := lairResponse{}
	for ctx.Err() == nil {
		time.Sleep(10 * time.Second)
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			continue
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &lairResponse)
		if err != nil {
			return nil, err
		}
		if len(lairResponse) == responses {
			break
		}
	}
	sort.Slice(lairResponse, func(i, j int) bool {
		return lairResponse[i].Signatures["ethereum"].Signer < lairResponse[j].Signatures["ethereum"].Signer
	})
	return lairResponse, nil
}
