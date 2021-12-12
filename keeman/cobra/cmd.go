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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/keeman/txt"
)

type Options struct {
	InputFile  string
	OutputFile string
	Index      int
}

func Execute() error {
	var opts Options
	cmd := &cobra.Command{Use: "keeman"}
	cmd.PersistentFlags().StringVarP(
		&opts.InputFile,
		"input",
		"i",
		"",
		"input file path",
	)
	cmd.PersistentFlags().StringVarP(
		&opts.OutputFile,
		"output",
		"o",
		"",
		"output file path",
	)
	cmd.PersistentFlags().IntVarP(
		&opts.Index,
		"index",
		"n",
		0,
		"data index",
	)
	cmd.AddCommand(
		NewHd(&opts),
		NewList(&opts),
	)
	return cmd.Execute()
}

func lineFromFile(filename string, idx int) (string, error) {
	lines, err := linesFromFile(filename)
	if err != nil {
		return "", err
	}
	return selectLine(lines, idx)
}

func linesFromFile(filename string) ([]string, error) {
	file, fileClose, err := inputFileOrStdin(filename)
	if err != nil {
		return nil, err
	}
	defer func() { err = fileClose() }()
	lines, err := txt.ReadNonEmptyLines(file, 0, false)
	return lines, err
}

func selectLine(lines []string, lineIdx int) (string, error) {
	if len(lines) <= lineIdx {
		return "", fmt.Errorf("data needs %d line(s)", lineIdx+1)
	}
	return lines[lineIdx], nil
}

func inputFileOrStdin(inputFilePath string) (*os.File, func() error, error) {
	if inputFilePath != "" {
		file, err := os.Open(inputFilePath)
		if err != nil {
			return nil, nil, err
		}
		return file, file.Close, nil
	} else if fi, err := os.Stdin.Stat(); err != nil {
		return nil, nil, err
	} else if fi.Size() <= 0 && fi.Mode()&os.ModeNamedPipe == 0 {
		return nil, nil, errors.New("no input file provided and stdin is empty")
	}
	return os.Stdin, func() error { return nil }, nil
}

func printLine(l string) {
	split := strings.Split(l, " ")
	fmt.Println(len(split), split[0])
}
