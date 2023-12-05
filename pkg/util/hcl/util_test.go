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

import "github.com/zclconf/go-cty/cty"

type textUnmarshaler struct {
	Val string
}

func (t *textUnmarshaler) UnmarshalText(text []byte) error {
	t.Val = string(text)
	return nil
}

func (t textUnmarshaler) MarshalText() ([]byte, error) {
	return []byte(t.Val), nil
}

type hclUnmarshaler struct {
	Val string
}

func (t *hclUnmarshaler) UnmarshalHCL(cty cty.Value) error {
	t.Val = cty.AsString()
	return nil
}

func (t hclUnmarshaler) MarshalHCL() (cty.Value, error) {
	return cty.StringVal(t.Val), nil
}

type basicTypes struct {
	String          string           `hcl:"string,optional"`
	Int             int32            `hcl:"int,optional"`
	Float           float32          `hcl:"float,optional"`
	Bool            bool             `hcl:"bool,optional"`
	Slice           []int            `hcl:"slice,optional"`
	Map             map[string]int   `hcl:"map,optional"`
	CTY             cty.Value        `hcl:"cty,optional"`
	TextUnmarshaler *textUnmarshaler `hcl:"text_unmarshaler,optional"`
	HCLUnmarshaler  *hclUnmarshaler  `hcl:"hcl_unmarshaler,optional"`
}

type block struct {
	Label string `hcl:",label"`
	Attr  string `hcl:"attr,optional"`
}

type emptyBlock struct {
	Label string `hcl:",label"`
}

type blocks struct {
	Single      block              `hcl:"single,block"`
	SinglePtr   *block             `hcl:"single_ptr,block"`
	Slice       []block            `hcl:"slice,block"`
	SlicePtr    []*block           `hcl:"slice_ptr,block"`
	Map         map[string]block   `hcl:"map,block"`
	MapPtr      map[string]*block  `hcl:"map_ptr,block"`
	PtrSlice    *[]block           `hcl:"ptr_slice,block"`
	PtrSlicePtr *[]*block          `hcl:"ptr_slice_ptr,block"`
	PtrMap      *map[string]block  `hcl:"ptr_map,block"`
	PtrMapPtr   *map[string]*block `hcl:"ptr_map_ptr,block"`
}

type singleBlock struct {
	Block block `hcl:"block,block"`
}

type singleEmptyBlock struct {
	Block emptyBlock `hcl:"block,block"`
}

type requiredAttrs struct {
	Var    string  `hcl:"var"`
	VarPtr *string `hcl:"var_ptr"`
}

type optionalAttrs struct {
	Var    string  `hcl:"var,optional"`
	VarPtr *string `hcl:"var_ptr,optional"`
}

type requiredBlocks struct {
	Block    block  `hcl:"block,block"`
	BlockPtr *block `hcl:"block_ptr,block"`
}

type optionalBlocks struct {
	Block    *block `hcl:"block,block,optional"`
	BlockPtr *block `hcl:"block_ptr,block,optional"`
}

type blockSlice struct {
	Slice []block `hcl:"slice,block"`
}

type ignoredField struct {
	Var string `hcl:"var,ignore"`
}

type anyField struct {
	Var any `hcl:"var"`
}
