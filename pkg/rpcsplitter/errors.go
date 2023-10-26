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

package rpcsplitter

import (
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
)

type errorList []error

func (e errorList) ErrorCode() int {
	codes := make([]int, 0, len(e))
	for _, err := range e {
		if err, ok := err.(rpc.Error); ok {
			codes = append(codes, err.ErrorCode())
		}
	}
	if len(codes) == 0 {
		return -32000
	}
	return mostCommon(codes)
}

func (e errorList) ErrorData() any {
	data := make([]any, 0, len(e))
	for _, err := range e {
		if err, ok := err.(rpc.DataError); ok {
			data = append(data, err.ErrorData())
		}
	}
	if len(data) == 0 {
		return nil
	}
	return mostCommon(data)
}

func (e errorList) Error() string {
	switch len(e) {
	case 0:
		return "unknown error"
	case 1:
		return e[0].Error()
	default:
		s := strings.Builder{}
		s.WriteString("the following errors occurred: ")
		s.WriteString("[")
		for n, err := range e {
			s.WriteString(err.Error())
			if n < len(e)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteString("]")
		return s.String()
	}
}

// addError adds an error to an error slice. If errs is not an error slice it
// will be converted into one. If there is already an error with the same
// message in the slice, it will not be added.
func addError(err error, errs ...error) error {
	if err == nil {
		err = errorList{}
	}
	if _, ok := err.(errorList); !ok {
		err = errorList{err}
	}
	errList := err.(errorList)
	for _, e := range errs {
		if e == nil {
			continue
		}
		f := false
		for _, e2 := range errList {
			if e.Error() == e2.Error() {
				f = true
				break
			}
		}
		if !f {
			errList = append(errList, e)
		}
	}
	return errList
}

func mostCommon[T comparable](s []T) T {
	if len(s) == 0 {
		return *new(T)
	}
	m := make(map[T]int)
	for _, v := range s {
		m[v]++
	}
	var max T
	var maxCount int
	for k, v := range m {
		if v > maxCount {
			maxCount = v
			max = k
		}
	}
	return max
}
