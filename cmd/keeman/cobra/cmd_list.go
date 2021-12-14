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

package cobra

import (
	"github.com/spf13/cobra"
)

func NewList(opts *Options) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "list [--all]",
		Short: "List word count and first word from the input, omitting the comments",
		RunE: func(_ *cobra.Command, args []string) error {
			if all {
				lines, err := linesFromFile(opts.InputFile)
				if err != nil {
					return err
				}
				for _, l := range lines {
					printLine(l)
				}
				return nil
			}
			l, err := lineFromFile(opts.InputFile, opts.Index)
			if err != nil {
				return err
			}
			printLine(l)
			return nil
		},
	}
	cmd.Flags().BoolVarP(
		&all,
		"all",
		"a",
		false,
		"all data",
	)
	return cmd
}
