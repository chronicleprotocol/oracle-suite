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

package retry

import (
	"context"
	"time"
)

// Try runs the f function until it returns nil but not more than defined in
// the attempts argument. After reaching the max attempts, it returns the last
// error. The delay argument defines the time between each attempt. If the
// context is canceled, the function stops and returns the error.
func Try(ctx context.Context, f func() error, attempts int, delay time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err = f(); err == nil {
			return nil
		}
		if attempts < 0 || i < attempts {
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
			case <-t.C:
			}
			t.Stop()
		}
	}
	return err
}

// TryForever runs the f function until it returns nil or the context is
// canceled. The delay argument defines the time between each attempt.
func TryForever(ctx context.Context, f func() error, delay time.Duration) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := f(); err == nil {
			return
		}
		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
		case <-t.C:
		}
		t.Stop()
	}
}
