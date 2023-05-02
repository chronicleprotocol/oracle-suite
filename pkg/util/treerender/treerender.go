package treerender

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

// NodeData contains data for a single node in the tree.
type NodeData struct {
	Name      string
	Params    map[string]any
	Ancestors []any
	Error     error
}

// RenderTree renders graphical tree for the CLI output.
//
// The callback is called for each node in the tree. It receives a NodeData
// structure that contains information about the node.
//
// The nodes arguments is an initial list of nodes to render.
//
// The level is used internally. It needs to be always 0.
//
//nolint:gocyclo
func RenderTree(callback func(any) NodeData, nodes []any, level int) []byte {
	const (
		first  = "┌──"
		middle = "├──"
		last   = "└──"
		vline  = "│  "
		hline  = "───"
		empty  = "   "
	)
	buf := bytes.Buffer{}
	for i, node := range nodes {
		data := callback(node)
		isFirst := i == 0
		isLast := i == len(nodes)-1
		hasAncestors := len(data.Ancestors) > 0

		// Determine the style of the branch line for the first line.
		firstLinePrefix := color("", reset)
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

		// Determine the style of the branch line for the rest of the lines
		// (if there are any).
		restLinesPrefix := color("", reset)
		switch {
		case isLast && hasAncestors:
			restLinesPrefix += color(empty+vline, green)
		case !isLast && hasAncestors:
			restLinesPrefix += color(vline+vline, green)
		case isLast && !hasAncestors:
			restLinesPrefix += color(empty+empty, green)
		case !isLast && !hasAncestors:
			restLinesPrefix += color(vline+empty, green)
		}

		// Render the node.
		buf.Write(prependLines(renderNode(data.Name, data.Params, data.Error), firstLinePrefix, restLinesPrefix))
		buf.WriteByte('\n')

		// Render the ancestors.
		if len(data.Ancestors) > 0 {
			subTree := RenderTree(callback, data.Ancestors, level+1)
			if isLast {
				subTree = prependLines(subTree, empty, empty)
			} else {
				subTree = prependLines(subTree, color(vline, green), color(vline, green))
			}
			buf.Write(subTree)
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes()
}

func renderNode(typ string, params map[string]any, err error) []byte {
	buf := bytes.Buffer{}
	if err != nil {
		buf.WriteString(color(typ, red))
	} else {
		buf.WriteString(typ)
	}
	buf.WriteString("(")
	for i, key := range maputil.SortKeys(params, sort.Strings) {
		buf.WriteString(color(key, green))
		buf.WriteString(":")
		buf.WriteString(fmt.Sprintf("%v", params[key]))
		if i != len(params)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(")")
	if err != nil {
		buf.WriteString("\n")
		buf.WriteString(color("Error: "+strings.TrimSpace(err.Error()), red))
	}
	return buf.Bytes()
}

// prependLines prepends all lines in given bytes slice.
func prependLines(s []byte, first, rest string) []byte {
	buf := bytes.Buffer{}
	buf.WriteString(first)
	buf.Write(bytes.ReplaceAll(bytes.TrimRight(s, "\n"), []byte{'\n'}, append([]byte{'\n'}, rest...)))
	return buf.Bytes()
}

// colorCode represents ANSII escape code for color formatting.
type colorCode string

const (
	reset colorCode = "\033[0m"
	red   colorCode = "\033[31m"
	green colorCode = "\033[32m"
)

// color adds given ANSII escape code at beginning of every line.
func color(str string, color colorCode) string {
	return string(color) + strings.ReplaceAll(str, "\n", "\n"+string(reset+color)) + string(reset)
}
