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

package txt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	type args struct {
		filename     string
		limit        int
		withComments bool
	}
	tests := []struct {
		name               string
		args               args
		want               []string
		wantedErrorMessage string
	}{
		{
			name: "no comments,no limit",
			args: args{
				filename:     "testdata/file.txt",
				limit:        0,
				withComments: false,
			},
			want: []string{"one", "two", "five", "seven", "nine"},
		},
		{
			name: "no comments,limit 1",
			args: args{
				filename:     "testdata/file.txt",
				limit:        1,
				withComments: false,
			},
			want: []string{"one"},
		},
		{
			name: "no comments,limit over comment",
			args: args{
				filename:     "testdata/file.txt",
				limit:        3,
				withComments: false,
			},
			want: []string{"one", "two", "five"},
		},
		{
			name: "no comments,limit equals line count",
			args: args{
				filename:     "testdata/file.txt",
				limit:        9,
				withComments: false,
			},
			want: []string{"one", "two", "five", "seven", "nine"},
		},
		{
			name: "no comments,limit over line count",
			args: args{
				filename:     "testdata/file.txt",
				limit:        10,
				withComments: false,
			},
			want: []string{"one", "two", "five", "seven", "nine"},
		},
		{
			name: "with comments,no limit",
			args: args{
				filename:     "testdata/file.txt",
				limit:        0,
				withComments: true,
			},
			want: []string{"one", "two", "#three", "# four", "five # comment", "# six", "seven", "# eight", "nine"},
		},
		{
			name: "with comments,limit 1",
			args: args{
				filename:     "testdata/file.txt",
				limit:        1,
				withComments: true,
			},
			want: []string{"one"},
		},
		{
			name: "with comments,limit over comment",
			args: args{
				filename:     "testdata/file.txt",
				limit:        3,
				withComments: true,
			},
			want: []string{"one", "two", "#three"},
		},
		{
			name: "with comments,limit equals line count",
			args: args{
				filename:     "testdata/file.txt",
				limit:        9,
				withComments: true,
			},
			want: []string{"one", "two", "#three", "# four", "five # comment", "# six", "seven", "# eight", "nine"},
		},
		{
			name: "with comments,limit over line count",
			args: args{
				filename:     "testdata/file.txt",
				limit:        10,
				withComments: true,
			},
			want: []string{"one", "two", "#three", "# four", "five # comment", "# six", "seven", "# eight", "nine"},
		},
		{
			name: "file does not exist",
			args: args{
				filename: "testdata/no-file.txt",
			},
			wantedErrorMessage: "open testdata/no-file.txt: no such file or directory",
		},
		{
			name: "file is empty",
			args: args{
				filename: "testdata/empty.txt",
			},
			wantedErrorMessage: "file is empty: testdata/empty.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadLines(tt.args.filename, tt.args.limit, tt.args.withComments)
			if tt.wantedErrorMessage != "" && assert.Error(t, err) {
				assert.Equal(t, tt.wantedErrorMessage, err.Error())
			} else {
				assert.NoError(t, err, "ReadFileLines")
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
func TestFileReadLines(t *testing.T) {
	type args struct {
		file         *os.File
		limit        int
		withComments bool
	}
	tests := []struct {
		name               string
		args               args
		want               []string
		wantedErrorMessage string
	}{
		{
			name: "file is empty",
			args: args{
				file: os.Stdin,
			},
			wantedErrorMessage: "file is empty: /dev/stdin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileReadLines(tt.args.file, tt.args.limit, tt.args.withComments)
			if tt.wantedErrorMessage != "" && assert.Error(t, err) {
				assert.Equal(t, tt.wantedErrorMessage, err.Error())
			} else {
				assert.NoError(t, err, "ReadFileLines")
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
