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

package origins

import (
	"encoding/json"
	"strconv"
	"time"
)

type stringAsFloat64 float64

func (s *stringAsFloat64) UnmarshalJSON(bytes []byte) error {
	var ss string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	f, err := strconv.ParseFloat(ss, 64)
	if err != nil {
		return err
	}
	*s = stringAsFloat64(f)
	return nil
}
func (s *stringAsFloat64) val() float64 {
	return float64(*s)
}

type firstStringFromSliceAsFloat64 float64

func (s *firstStringFromSliceAsFloat64) UnmarshalJSON(bytes []byte) error {
	var ss []string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	f, err := strconv.ParseFloat(ss[0], 64)
	if err != nil {
		return err
	}
	*s = firstStringFromSliceAsFloat64(f)
	return nil
}

func (s *firstStringFromSliceAsFloat64) val() float64 {
	return float64(*s)
}

type stringAsUnixTimestamp time.Time

func (s *stringAsUnixTimestamp) UnmarshalJSON(bytes []byte) error {
	var ss string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	i, err := strconv.ParseInt(ss, 10, 64)
	if err != nil {
		return err
	}
	*s = stringAsUnixTimestamp(time.Unix(i, 0))
	return nil
}

func (s *stringAsUnixTimestamp) val() time.Time {
	return time.Time(*s)
}

type stringAsInt64 int64

func (s *stringAsInt64) UnmarshalJSON(bytes []byte) error {
	var ss string
	if err := json.Unmarshal(bytes, &ss); err != nil {
		return err
	}
	i, err := strconv.ParseInt(ss, 10, 64)
	if err != nil {
		return err
	}
	*s = stringAsInt64(i)
	return nil
}

func (s *stringAsInt64) val() int64 {
	return int64(*s)
}

type intAsUnixTimestamp time.Time

func (s *intAsUnixTimestamp) UnmarshalJSON(bytes []byte) error {
	var i int64
	if err := json.Unmarshal(bytes, &i); err != nil {
		return err
	}
	*s = intAsUnixTimestamp(time.Unix(i, 0))
	return nil
}
func (s *intAsUnixTimestamp) val() time.Time {
	return time.Time(*s)
}

type intAsUnixTimestampMs time.Time

func (s *intAsUnixTimestampMs) UnmarshalJSON(bytes []byte) error {
	var i int64
	if err := json.Unmarshal(bytes, &i); err != nil {
		return err
	}
	*s = intAsUnixTimestampMs(time.Unix(i/1000, 0))
	return nil
}
func (s *intAsUnixTimestampMs) val() time.Time {
	return time.Time(*s)
}
