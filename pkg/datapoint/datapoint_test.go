//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package datapoint

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
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

func TestDataPoint_LogFields(t *testing.T) {
	testCases := []struct {
		name      string
		dataPoint Point
		expected  log.Fields
	}{
		{
			name: "valid data point",
			dataPoint: Point{
				Value: stringValue("value"),
				Time:  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
			},
			expected: log.Fields{
				"point.time":  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				"point.value": "value",
			},
		},
		{
			name: "error is set",
			dataPoint: Point{
				Value: stringValue("value"),
				Time:  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				Error: errors.New("some error"),
			},
			expected: log.Fields{
				"point.time":  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				"point.value": "value",
				"point.error": "some error",
			},
		},
		{
			name: "meta is set",
			dataPoint: Point{
				Value: stringValue("value"),
				Time:  time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				Meta: log.Fields{
					"key": "value",
				},
			},
			expected: log.Fields{
				"point.time":     time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				"point.value":    "value",
				"point.meta.key": "value",
			},
		},
		{
			name:      "empty data point",
			dataPoint: Point{},
			expected: log.Fields{
				"point.error": "value is not set",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields := PointLogFields(tc.dataPoint)
			assert.Equal(t, tc.expected, fields)
		})
	}
}
