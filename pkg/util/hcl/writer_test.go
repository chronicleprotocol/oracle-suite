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

package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlock_Bytes(t *testing.T) {
	testCases := []struct {
		name     string
		block    *Block
		expected string
		hasError bool
	}{
		{
			name:     "empty root block",
			block:    &Block{},
			expected: "",
			hasError: false,
		},
		{
			name: "attributes",
			block: &Block{
				Attributes: []*Attribute{
					{Name: "attr1", Value: "value1"},
					{Name: "attr2", Value: "value2"},
				},
			},
			expected: "attr1 = \"value1\"\nattr2 = \"value2\"\n",
			hasError: false,
		},
		{
			name: "empty block",
			block: &Block{
				Blocks: []*Block{
					{TypeName: "child", Labels: []string{"label1"}},
				},
			},
			expected: "child \"label1\" {}\n",
		},
		{
			name: "block with attributes",
			block: &Block{
				Blocks: []*Block{
					{
						TypeName: "child",
						Labels:   []string{"label1"},
						Attributes: []*Attribute{
							{Name: "attr1", Value: "value1"},
						},
					},
				},
			},
			expected: "child \"label1\" { attr1 = \"value1\" }\n",
		},
		{
			name: "block with blocks",
			block: &Block{
				Blocks: []*Block{
					{
						TypeName: "child",
						Labels:   []string{"label1"},
						Blocks: []*Block{
							{TypeName: "grandchild", Labels: []string{"label2"}},
						},
					},
				},
			},
			expected: "child \"label1\" {\n  grandchild \"label2\" {}\n}\n",
		},
		{
			name: "block with attributes and blocks",
			block: &Block{
				Blocks: []*Block{
					{
						TypeName: "child",
						Labels:   []string{"label1"},
						Attributes: []*Attribute{
							{Name: "attr1", Value: "value1"},
						},
						Blocks: []*Block{
							{TypeName: "grandchild", Labels: []string{"label2"}},
						},
					},
				},
			},
			expected: "child \"label1\" {\n  attr1 = \"value1\"\n\n  grandchild \"label2\" {}\n}\n",
		},
		{
			name: "multiline block",
			block: &Block{
				Blocks: []*Block{
					{
						TypeName: "child",
						Labels:   []string{"label1"},
						Attributes: []*Attribute{
							{Name: "attr1", Value: "value1"},
							{Name: "attr2", Value: "value2"},
						},
					},
				},
			},
			expected: "child \"label1\" {\n  attr1 = \"value1\"\n  attr2 = \"value2\"\n}\n",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hcl, diags := tc.block.Bytes()
			assert.Equal(t, tc.hasError, diags.HasErrors())
			assert.Equal(t, tc.expected, string(hcl))
		})
	}
}
