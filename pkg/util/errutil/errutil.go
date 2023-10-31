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

package errutil

import "strings"

// Append combines the provided error with a list of errors.
func Append(err error, errs ...error) error {
	if err == nil && len(errs) == 0 {
		return nil
	}
	var mErr MultiError
	if e, ok := err.(MultiError); ok {
		mErr = e
	} else if err != nil {
		mErr = MultiError{err}
	}
	for _, e := range errs {
		if e == nil {
			continue
		}
		if m, ok := e.(MultiError); ok {
			mErr = append(mErr, m...)
		} else {
			mErr = append(mErr, e)
		}
	}
	switch len(mErr) {
	case 0:
		return nil
	case 1:
		return mErr[0]
	default:
		return mErr
	}
}

// MultiError is a collection of errors.
type MultiError []error

// Error implements the error interface.
func (m MultiError) Error() string {
	if len(m) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("following errors occurred: [")
	for i, err := range m {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(err.Error())
	}
	b.WriteString("]")
	return b.String()
}

// Ignore is a helper function that returns the first argument and ignores the
// error.
func Ignore[T any](v T, _ error) T {
	return v
}

// Must is a helper function that panics when the error is not nil. Otherwise,
// it returns the first argument. It is intended for use with functions that
// should never return an error when called.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
