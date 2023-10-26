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

package main

import (
	"bufio"
	"io"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

var completeDataModels = map[string]string{}

func init() {
	readFile, err := os.Open("./testdata/models.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)
	modelName := ""
	modelTree := strings.Builder{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Model for ") {
			if modelName != "" {
				completeDataModels[modelName] = modelTree.String()
				modelTree.Reset()
			}
			modelName = line[10 : len(line)-1]
			continue
		}
		modelTree.WriteString(line)
		modelTree.WriteString("\n")
	}
	completeDataModels[modelName] = modelTree.String()
	modelTree.Reset()
	readFile.Close()
}

func TestNewModelsCmd_List(t *testing.T) {
	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()

	err := os.Setenv("ETH_RPC_URL", "http://localhost:8545")
	require.NoErrorf(t, err, "failed to set ETH_RPC_URL")

	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"gofer", "-v", "warning", "models"}
	main()

	_ = w.Close()
	out, _ := io.ReadAll(r)

	keys := maputil.Keys(completeDataModels)
	sort.Strings(keys)

	respKeys := strings.Split(strings.TrimSpace(string(out)), "\n")
	sort.Strings(respKeys)

	assert.Equal(t, keys, respKeys)
}

func TestNewModelsCmd(t *testing.T) {
	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()

	err := os.Setenv("ETH_RPC_URL", "http://localhost:8545")
	require.NoErrorf(t, err, "failed to set ETH_RPC_URL")

	keys := maputil.Keys(completeDataModels)
	sort.Strings(keys)
	modelList := strings.Join(keys, "\n") + "\n"

	for _, model := range strings.Split(strings.Trim(modelList, "\n"), "\n") {
		t.Run("Model for "+model, func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stdout = w

			os.Args = []string{"gofer", "models", "-v", "warning", "-o", "trace", "--no-color", model}
			main()

			_ = w.Close()
			out, _ := io.ReadAll(r)

			s, ok := completeDataModels[model]
			out = out[strings.Index(string(out), "\n")+1:]
			require.Truef(t, ok, "missing model for %s", model)
			assert.Equal(t, strings.Trim(s, "\n")+"\n\n", string(out))
		})
	}
}
