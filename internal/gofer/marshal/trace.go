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

package marshal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
)

type traceItem struct {
	writer io.Writer
	item   []byte
}

type trace struct {
	items []traceItem
}

func newTrace() *trace {
	return &trace{}
}

// Write implements the Marshaller interface.
func (t *trace) Write(writer io.Writer, item interface{}) error {
	var i []byte
	switch typedItem := item.(type) {
	case *gofer.Price:
		i = t.handlePrice(typedItem)
	case *gofer.Model:
		i = t.handleModel(typedItem)
	case error:
		i = []byte(fmt.Sprintf("Error: %s", typedItem.Error()))
	default:
		return fmt.Errorf("unsupported data type")
	}

	t.items = append(t.items, traceItem{writer: writer, item: i})
	return nil
}

// Flush implements the Marshaller interface.
func (t *trace) Flush() error {
	var err error
	for _, i := range t.items {
		_, err = i.writer.Write(i.item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*trace) handlePrice(price *gofer.Price) []byte {
	tree := renderTree(func(node interface{}) ([]byte, []interface{}) {
		t := node.(*gofer.Price)
		var tErr error
		if t.Error != "" {
			tErr = errors.New(t.Error)
		}

		s := renderNode(
			t.Type,
			mergeKVMap(
				[]param{
					{key: "pair", value: t.Pair.String()},
					{key: "price", value: t.Price},
					{key: "timestamp", value: t.Time.In(time.UTC).Format(time.RFC3339Nano)},
				},
				t.Parameters,
			),
			tErr,
		)

		var c []interface{}
		for _, tc := range t.Prices {
			c = append(c, tc)
		}

		return s, c
	}, []interface{}{price}, 0)

	buf := bytes.Buffer{}
	buf.Write([]byte(fmt.Sprintf("Price for %s:\n", price.Pair)))
	buf.Write(tree)
	return buf.Bytes()
}

func (t *trace) handleModel(node *gofer.Model) []byte {
	tree := renderTree(func(node interface{}) ([]byte, []interface{}) {
		n := node.(*gofer.Model)
		s := renderNode(
			n.Type,
			mergeKVMap(
				[]param{
					{key: "pair", value: n.Pair.String()},
				},
				n.Parameters,
			),
			nil,
		)

		var c []interface{}
		for _, nc := range n.Models {
			c = append(c, nc)
		}

		return s, c
	}, []interface{}{node}, 0)

	buf := bytes.Buffer{}
	buf.Write([]byte(fmt.Sprintf("Graph for %s:\n", node.Pair)))
	buf.Write(tree)
	return buf.Bytes()
}

// param is used to work with lists of sorted key/value pairs.
type param struct {
	key   string
	value interface{}
}

// mergeKVMap merges map[string]string into []param.
func mergeKVMap(target []param, kv map[string]string) []param {
	for k, v := range kv {
		target = append(target, param{key: k, value: v})
	}
	sort.Slice(target, func(i, j int) bool {
		return target[i].key < target[j].key
	})
	return target
}

// renderNode renders graph node which may be used as nodes for renderTree
// method. An example node may look like this: Type(param:value, param2:value2).
// If an err argument is provided, the node will be prepended with an [ERROR]
// label and a message will be printed in a new line.
func renderNode(typ string, params []param, err error) []byte {
	str := bytes.Buffer{}

	if err != nil {
		str.WriteString(color(typ, red))
	} else {
		str.WriteString(typ)
	}

	str.WriteString("(")
	for i, p := range params {
		str.WriteString(color(p.key, green))
		str.WriteString(":")
		str.WriteString(fmt.Sprintf("%v", p.value))
		if i != len(params)-1 {
			str.WriteString(", ")
		}
	}
	str.WriteString(")")
	if err != nil {
		str.WriteString("\n")
		str.WriteString(color("Error: "+strings.TrimSpace(err.Error()), red))
	}

	return str.Bytes()
}

// renderTree renders graphical tree for the CLI output.
//
// The printer argument defines a function which returns node name and list of
// child nodes.
// The nodes arguments is a initial list of nodes to render.
// The level is used internally and needs to be always 0.
//
//nolint:gocyclo
func renderTree(printer func(interface{}) ([]byte, []interface{}), nodes []interface{}, level int) []byte {
	const (
		first  = "┌──"
		middle = "├──"
		last   = "└──"
		vline  = "│  "
		hline  = "───"
		empty  = "   "
	)

	s := bytes.Buffer{}
	for i, node := range nodes {
		nodeStr, nodeChildren := printer(node)
		isFirst := i == 0
		isLast := i == len(nodes)-1
		hasChild := len(nodeChildren) > 0
		firstLinePrefix := color("", reset)
		restLinesPrefix := color("", reset)

		switch {
		case level == 0 && isFirst && isLast:
			firstLinePrefix += color(hline, green)
		case level == 0 && isFirst:
			firstLinePrefix += color(first, green)
		case isLast:
			firstLinePrefix += color(last, green)
		default:
			firstLinePrefix += color(middle, green)
		}

		switch {
		case isLast && hasChild:
			restLinesPrefix += color(empty+vline, green)
		case !isLast && hasChild:
			restLinesPrefix += color(vline+vline, green)
		case isLast && !hasChild:
			restLinesPrefix += color(empty+empty, green)
		case !isLast && !hasChild:
			restLinesPrefix += color(vline+empty, green)
		}

		s.Write(prependLines(nodeStr, firstLinePrefix, restLinesPrefix))
		s.WriteByte('\n')

		if len(nodeChildren) > 0 {
			subTree := renderTree(printer, nodeChildren, level+1)

			if isLast {
				subTree = prependLines(subTree, empty, empty)
			} else {
				subTree = prependLines(subTree, color(vline, green), color(vline, green))
			}

			s.Write(subTree)
			s.WriteByte('\n')
		}
	}

	return s.Bytes()
}

// prependLines prepends all lines in given bytes slice.
func prependLines(s []byte, first, rest string) []byte {
	bts := bytes.Buffer{}
	bts.WriteString(first)
	bts.Write(bytes.ReplaceAll(bytes.TrimRight(s, "\n"), []byte{'\n'}, append([]byte{'\n'}, rest...)))
	return bts.Bytes()
}
