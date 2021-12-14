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
	"fmt"
	"io"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
)

type plainItem struct {
	writer io.Writer
	item   []byte
}

type plain struct {
	items []plainItem
}

func newPlain() *plain {
	return &plain{}
}

// Write implements the Marshaller interface.
func (p *plain) Write(writer io.Writer, item interface{}) error {
	var i []byte
	switch typedItem := item.(type) {
	case *gofer.Price:
		i = p.handlePrice(typedItem)
	case *gofer.Model:
		i = p.handleModel(typedItem)
	case error:
		i = []byte(fmt.Sprintf("Error: %s", typedItem.Error()))
	default:
		return fmt.Errorf("unsupported data type")
	}

	p.items = append(p.items, plainItem{writer: writer, item: i})
	return nil
}

// Flush implements the Marshaller interface.
func (p *plain) Flush() error {
	var err error
	for _, i := range p.items {
		_, err = i.writer.Write(i.item)
		if err != nil {
			return err
		}
		_, err = i.writer.Write([]byte{'\n'})
		if err != nil {
			return err
		}
	}
	return nil
}

func (*plain) handlePrice(price *gofer.Price) []byte {
	if price.Error != "" {
		return []byte(fmt.Sprintf("%s - %s", price.Pair, strings.TrimSpace(price.Error)))
	}
	return []byte(fmt.Sprintf("%s %f", price.Pair, price.Price))
}

func (*plain) handleModel(node *gofer.Model) []byte {
	return []byte(node.Pair.String())
}
