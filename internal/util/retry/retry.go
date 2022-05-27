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

package retry

import (
	"context"
	"time"
)

// Retry runs the f function until it returns nil.
func Retry(ctx context.Context, f func() error, attempts int, delay time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = f()
		if err == nil {
			return nil
		}
		if i != attempts-1 {
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
