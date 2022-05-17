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

	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/txt"
)

type Options struct {
	InputFile  string
	OutputFile string
	Verbose    bool
}

func Command() (*Options, *cobra.Command) {
	return &Options{}, &cobra.Command{
		Use: "keeman",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
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
	return txt.ReadNonEmptyLines(file, 0, false)
}

func selectLine(lines []string, lineIdx int) (string, error) {
	if len(lines) <= lineIdx {
		return "", fmt.Errorf("data needs %d line(s)", lineIdx+1)
	}
	return lines[lineIdx], nil
}

func inputFileOrStdin(inputFilePath string) (*os.File, func() error, error) {
	if inputFilePath != "" {
		f, err := os.Open(inputFilePath)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to open file: %w", err)
		}
		return f, f.Close, nil
	}
	stdin, err := NonEmptyStdIn()
	return stdin, func() error { return nil }, err
}
func NonEmptyStdIn() (*os.File, error) {
	if fi, err := os.Stdin.Stat(); err != nil {
		return nil, fmt.Errorf("unable to stat stdin: %w", err)
	} else if fi.Size() <= 0 && fi.Mode()&os.ModeNamedPipe == 0 {
		return nil, errors.New("stdin is empty")
	}
	return os.Stdin, nil
}
func printLine(l string) {
	split := strings.Split(l, " ")
	fmt.Println(len(split), split[0])
}
