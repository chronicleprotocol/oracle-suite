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
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// Block represents a single HCL block.
//
// If typeName is empty, the block is a root block.
//
// It is used to build an HCL configuration.
type Block struct {
	TypeName   string
	Labels     []string
	Attributes []*Attribute
	Blocks     []*Block
}

// Attribute represents a single attribute in a HCL block.
//
// It is used to build am HCL configuration.
type Attribute struct {
	Name  string
	Value any
}

// AppendAttribute appends a new attribute to the given block.
func (b *Block) AppendAttribute(name string, value any) {
	b.Attributes = append(b.Attributes, &Attribute{
		Name:  name,
		Value: value,
	})
}

// AppendBlock appends a new block to the given block.
func (b *Block) AppendBlock(typeName string, labels []string) *Block {
	block := &Block{
		TypeName: typeName,
		Labels:   labels,
	}
	b.Blocks = append(b.Blocks, block)
	return block
}

// Bytes returns the HCL configuration.
func (b *Block) Bytes() ([]byte, hcl.Diagnostics) {
	file := hclwrite.NewEmptyFile()
	if diags := b.Write(file.Body()); diags.HasErrors() {
		return nil, diags
	}
	return file.Bytes(), nil
}

// Write writes the block to the given HCL body.
func (b *Block) Write(hclBody *hclwrite.Body) hcl.Diagnostics {
	blockTokens, diags := tokensForBlock(b)
	if diags.HasErrors() {
		return diags
	}
	hclBody.AppendUnstructuredTokens(blockTokens)
	return nil
}

// tokensForValue returns tokens for the given value.
func tokensForValue(val any) (hclwrite.Tokens, hcl.Diagnostics) {
	var ctyVal cty.Value
	if err := mapper.Map(val, &ctyVal); err != nil {
		return nil, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Encode error",
			Detail:   err.Error(),
		}}
	}
	return hclwrite.TokensForValue(ctyVal), nil
}

// tokensForAttribute returns tokens for the given attribute.
func tokensForAttribute(attr *Attribute) (hclwrite.Tokens, hcl.Diagnostics) {
	attrTokens, diags := tokensForValue(attr.Value)
	if diags.HasErrors() {
		return nil, diags
	}
	return appendTokens(
		hclwrite.TokensForIdentifier(attr.Name),
		hclwrite.Tokens{
			{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}},
		},
		attrTokens,
		hclwrite.Tokens{
			{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
		},
	), nil
}

// tokensForBlock returns tokens for the given block. Blocks are written in
// a single line if possible.
//
//nolint:gocyclo
func tokensForBlock(block *Block) (hclwrite.Tokens, hcl.Diagnostics) {
	// Special case for root block.
	if block.TypeName == "" {
		tokens := hclwrite.Tokens{}

		// Top level attributes.
		for _, attr := range block.Attributes {
			attrTokens, diags := tokensForAttribute(attr)
			if diags.HasErrors() {
				return nil, diags
			}
			tokens = appendTokens(tokens, attrTokens)
		}

		// New line between attributes and blocks.
		if len(block.Attributes) > 0 && len(block.Blocks) > 0 {
			tokens = appendTokens(
				tokens,
				hclwrite.Tokens{
					{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
				},
			)
		}

		// Top level blocks.
		for _, block := range block.Blocks {
			blockTokens, diags := tokensForBlock(block)
			if diags.HasErrors() {
				return nil, diags
			}
			tokens = appendTokens(
				tokens,
				blockTokens,
			)
		}

		return tokens, nil
	}

	// Block type.
	tokens := hclwrite.TokensForIdentifier(block.TypeName)

	// Block labels.
	for _, label := range block.Labels {
		tokens = append(tokens, hclwrite.TokensForValue(cty.StringVal(label))...)
	}

	// Empty single line block.
	if len(block.Attributes) == 0 && len(block.Blocks) == 0 {
		tokens = appendTokens(
			tokens,
			hclwrite.Tokens{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
				{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
			},
		)
		return tokens, nil
	}

	// Single line block with one attribute.
	if len(block.Attributes) == 1 && len(block.Blocks) == 0 {
		attrTokens, diags := tokensForValue(block.Attributes[0].Value)
		if diags.HasErrors() {
			return nil, diags
		}
		if multilineTokens(attrTokens) {
			// If block contains multiline attribute, then it must be
			// written as a multiline block.
			goto multiline
		}
		tokens = appendTokens(
			tokens,
			hclwrite.Tokens{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
			},
			hclwrite.TokensForIdentifier(block.Attributes[0].Name),
			hclwrite.Tokens{
				{Type: hclsyntax.TokenEqual, Bytes: []byte{'='}},
			},
			attrTokens,
			hclwrite.Tokens{
				{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
			},
		)
		return tokens, nil
	}

	// Multiline block.
multiline:

	// Opening brace.
	tokens = appendTokens(
		tokens,
		hclwrite.Tokens{
			{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}},
			{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
		},
	)

	// Block attributes.
	for _, attr := range block.Attributes {
		attrTokens, diags := tokensForAttribute(attr)
		if diags.HasErrors() {
			return nil, diags
		}
		tokens = appendTokens(tokens, attrTokens)
	}

	// New line between attributes and blocks.
	if len(block.Attributes) > 0 && len(block.Blocks) > 0 {
		tokens = appendTokens(
			tokens,
			hclwrite.Tokens{
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
			},
		)
	}

	// Nested blocks.
	for _, nestedBlock := range block.Blocks {
		nestedBlockTokens, diags := tokensForBlock(nestedBlock)
		if diags.HasErrors() {
			return nil, diags
		}
		tokens = appendTokens(
			tokens,
			nestedBlockTokens,
		)
	}

	// Closing brace.
	tokens = appendTokens(
		tokens,
		hclwrite.Tokens{
			{Type: hclsyntax.TokenCBrace, Bytes: []byte{'}'}},
			{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}},
		},
	)

	return tokens, nil
}

// multilineTokens returns true if the given tokens contain a newline.
func multilineTokens(tokens hclwrite.Tokens) bool {
	for _, token := range tokens {
		if token.Type == hclsyntax.TokenNewline {
			return true
		}
	}
	return false
}

// appendTokens appends tokens to the given tokens.
func appendTokens(tokens ...hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) == 0 {
		return nil
	}
	result := tokens[0]
	for _, tokens := range tokens[1:] {
		result = append(result, tokens...)
	}
	return result
}
