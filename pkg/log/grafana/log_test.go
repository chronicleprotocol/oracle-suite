package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestLogger(t *testing.T) {
	type want struct {
		name  string
		value float64
		tags  []string
	}
	tests := []struct {
		metrics []Metric
		want    []want
		logs    func(l log.Logger)
	}{
		// MatchMessage test:
		{
			metrics: []Metric{{MatchMessage: regexp.MustCompile("foo"), Name: "test"}},
			want: []want{
				{name: "test", value: 1},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
				l.Info("bar")
			},
		},
		// MatchFields test:
		{
			metrics: []Metric{{MatchFields: map[string]*regexp.Regexp{"key": regexp.MustCompile("foo")}, Name: "test"}},
			want: []want{
				{name: "test", value: 1},
			},
			logs: func(l log.Logger) {
				l.WithField("key", "foo").Info("test")
				l.WithField("key", "bar").Info("test")
			},
		},
		// Value form string:
		{
			metrics: []Metric{{MatchMessage: regexp.MustCompile("foo"), Name: "test", Value: "key"}},
			want: []want{
				{name: "test", value: 1.5},
			},
			logs: func(l log.Logger) {
				l.WithField("key", "1.5").Info("foo")
			},
		},
		// Value from float:
		{
			metrics: []Metric{{MatchMessage: regexp.MustCompile("foo"), Name: "test", Value: "key"}},
			want: []want{
				{name: "test", value: 1.5},
			},
			logs: func(l log.Logger) {
				l.WithField("key", 1.5).Info("foo")
			},
		},
		// Value from time duration as seconds:
		{
			metrics: []Metric{{MatchMessage: regexp.MustCompile("foo"), Name: "test", Value: "key"}},
			want: []want{
				{name: "test", value: 1.5},
			},
			logs: func(l log.Logger) {
				l.WithField("key", time.Millisecond*1500).Info("foo")
			},
		},
		// Replace variables:
		{
			metrics: []Metric{{
				MatchMessage: regexp.MustCompile("foo"),
				Name:         "test.${key1}.${key2}.${key3.a}",
				Tags: map[string][]string{
					"tag1": {"${key1}"},
					"tag2": {"${key2}", "${key3.a}"},
				},
			}},
			want: []want{
				{
					name:  "test.a.42.b",
					tags:  []string{"tag1=a", "tag2=42", "tag2=b"},
					value: 1,
				},
			},
			logs: func(l log.Logger) {
				l.WithFields(log.Fields{
					"key1": "a",
					"key2": 42,
					"key3": map[string]string{"a": "b"},
				}).Info("foo")
			},
		},
		// Multiple matches:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a"},
				{MatchMessage: regexp.MustCompile("foo"), Name: "b"},
			},
			want: []want{
				{name: "a", value: 1},
				{name: "b", value: 1},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
				l.Info("foo")
			},
		},
		// Replace duplicated values:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a"},
			},
			want: []want{
				{name: "a", value: 1},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
				l.Info("foo")
			},
		},
		// Sum duplicated values:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", OnDuplicate: Sum},
			},
			want: []want{
				{name: "a", value: 2},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
				l.Info("foo")
			},
		},
		// Sub duplicated values:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", OnDuplicate: Sub},
			},
			want: []want{
				{name: "a", value: 0},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
				l.Info("foo")
			},
		},
		// Use lower value:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", Value: "val", OnDuplicate: Min},
			},
			want: []want{
				{name: "a", value: 1},
			},
			logs: func(l log.Logger) {
				l.WithField("val", 2).Info("foo")
				l.WithField("val", 1).Info("foo")
				l.WithField("val", 2).Info("foo")
			},
		},
		// Use higher value:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", Value: "val", OnDuplicate: Max},
			},
			want: []want{
				{name: "a", value: 2},
			},
			logs: func(l log.Logger) {
				l.WithField("val", 1).Info("foo")
				l.WithField("val", 2).Info("foo")
				l.WithField("val", 1).Info("foo")
			},
		},
		// Ignore duplicated logs:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", Value: "val", OnDuplicate: Ignore},
			},
			want: []want{
				{name: "a", value: 1},
			},
			logs: func(l log.Logger) {
				l.WithField("val", 1).Info("foo")
				l.WithField("val", 2).Info("foo")
			},
		},
		// Scale value by 10^2:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", Value: "val", TransformFunc: func(v float64) float64 { return v / math.Pow(10, 2) }},
			},
			want: []want{
				{name: "a", value: 0.01},
			},
			logs: func(l log.Logger) {
				l.WithField("val", 1).Info("foo")
			},
		},
		// Ignore metrics that uses invalid path in a name:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "test.${invalid}"},
				{MatchMessage: regexp.MustCompile("foo"), Name: "test.valid"},
			},
			want: []want{
				{name: "test.valid", value: 1},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
			},
		},
		// Ignore tags with that uses invalid path:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile("foo"), Name: "a", Tags: map[string][]string{"tag": {"${invalid}"}}},
			},
			want: []want{
				{name: "a", value: 1},
			},
			logs: func(l log.Logger) {
				l.Info("foo")
			},
		},
		// Test all log types except panics:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile(".*"), Name: "${name}"},
			},
			want: []want{
				{name: "debug", value: 1},
				{name: "debugf", value: 1},
				{name: "error", value: 1},
				{name: "errorf", value: 1},
				{name: "info", value: 1},
				{name: "infof", value: 1},
				{name: "warn", value: 1},
				{name: "warnf", value: 1},
			},
			logs: func(l log.Logger) {
				l.WithField("name", "debug").Debug("debug")
				l.WithField("name", "debugf").Debugf("debugf")
				l.WithField("name", "error").Error("error")
				l.WithField("name", "errorf").Errorf("errorf")
				l.WithField("name", "info").Info("info")
				l.WithField("name", "infof").Infof("infof")
				l.WithField("name", "warn").Warn("warn")
				l.WithField("name", "warnf").Warnf("warnf")
			},
		},
		// Test panic:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile(".*"), Name: "${name}"},
			},
			want: []want{
				{name: "panic", value: 1},
			},
			logs: func(l log.Logger) {
				l.WithField("name", "panic").Panic("panic")
			},
		},
		// Test panicf:
		{
			metrics: []Metric{
				{MatchMessage: regexp.MustCompile(".*"), Name: "${name}"},
			},
			want: []want{
				{name: "panicf", value: 1},
			},
			logs: func(l log.Logger) {
				l.WithField("name", "panicf").Panicf("panicf")
			},
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			ctx, ctxCancel := context.WithCancel(context.Background())
			defer ctxCancel()
			r := int32(0)
			l, _ := New(log.Debug, Config{
				Metrics:          tt.metrics,
				Interval:         1,
				GraphiteEndpoint: "https://example.com/metrics",
				GraphiteAPIKey:   "test",
				HTTPClient: &http.Client{Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					defer atomic.StoreInt32(&r, 1)

					// Verify URL and auth key:
					assert.Equal(t, "https://example.com/metrics", req.URL.String())
					assert.Equal(t, "Bearer test", req.Header.Get("Authorization"))

					// Verify data:
					var data []metricJSON
					body, _ := io.ReadAll(req.Body)
					require.NoError(t, json.Unmarshal(body, &data))
					require.Len(t, data, len(tt.want))
					sort.Slice(data, func(i, j int) bool { return data[i].Name < data[j].Name })
					for n, w := range tt.want {
						assert.Equal(t, w.name, data[n].Name)
						assert.Equal(t, uint(1), data[n].Interval)
						assert.Equal(t, w.value, data[n].Value)
						assert.Greater(t, data[n].Time, int64(0))
					}

					return nil
				})},
				Logger: null.New(),
			})
			if l, ok := l.(log.LoggerService); ok {
				require.NoError(t, l.Start(ctx))
			}

			// Execute logs:
			func() {
				defer func() { recover() }()
				tt.logs(l)
			}()

			// Wait for result:
			for n := 0; n < 3 && atomic.LoadInt32(&r) == 0; n++ {
				time.Sleep(time.Second * 1)
			}
			require.Equal(t, int32(1), atomic.LoadInt32(&r), "logs has not been sent")
		})
	}
}

func Test_byPath(t *testing.T) {
	tests := []struct {
		value   interface{}
		path    string
		want    interface{}
		invalid bool
	}{
		{
			value: "test",
			path:  "",
			want:  "test",
		},
		{
			value:   "test",
			path:    "abc",
			invalid: true,
		},
		{
			value: struct {
				Field string
			}{
				Field: "test",
			},
			path: "Field",
			want: "test",
		},
		{
			value: map[string]string{"Key": "test"},
			path:  "Key",
			want:  "test",
		},
		{
			value:   map[int]string{42: "test"},
			path:    "42",
			invalid: true,
		},
		{
			value: []string{"test"},
			path:  "0",
			want:  "test",
		},
		{
			value: struct {
				Field map[string][]int
			}{
				Field: map[string][]int{"Field2": {42}},
			},
			path: "Field.Field2.0",
			want: 42,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := byPath(reflect.ValueOf(tt.value), tt.path)
			if tt.invalid {
				assert.False(t, v.IsValid())
			} else {
				assert.Equal(t, tt.want, v.Interface())
			}
		})
	}
}
