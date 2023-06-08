package datapoint

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stringValue string

func (s stringValue) Print() string {
	return string(s)
}

func (s stringValue) MarshalBinary() (data []byte, err error) {
	return nil, errors.New("not implemented")
}

func (s *stringValue) UnmarshalBinary(_ []byte) error {
	return errors.New("not implemented")
}

func TestDataPoint_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		dataPoint     Point
		expectError   bool
		errorContains string
	}{
		{
			name: "valid data point",
			dataPoint: Point{
				Value: stringValue("value"),
				Time:  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
			},
			expectError: false,
		},
		{
			name: "error is set",
			dataPoint: Point{
				Value: stringValue("value"),
				Time:  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				Error: errors.New("some error"),
			},
			expectError:   true,
			errorContains: "some error",
		},
		{
			name: "value is nil",
			dataPoint: Point{
				Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
			},
			expectError:   true,
			errorContains: "value is not set",
		},
		{
			name: "time is not set",
			dataPoint: Point{
				Value: stringValue("value"),
			},
			expectError:   true,
			errorContains: "time is not set",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dataPoint.Validate()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
