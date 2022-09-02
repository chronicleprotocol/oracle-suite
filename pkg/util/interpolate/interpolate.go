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

package interpolate

import (
	"bytes"
	"strings"
)

// Parse parses the string and returns a parsed representation. The variables
// may be interpolated later using the Interpolate method.
//
// The syntax is similar to shell variable expansion. The following rules apply:
//
// - Variables are enclosed in ${...} and may contain any character.
//
// - To include a literal $ in the output, escape it with a backslash or
// another $. For example, \$ and $$ are both interpreted as a literal $.
// The latter does not work inside a variable.
//
// - If a variable is not closed, it is treated as a literal.
func Parse(s string) String {
	return parse(tokenize(s))
}

// String is a parsed string.
type String []any

// Interpolate replaces variables in the string based on the mapping function.
func (s String) Interpolate(mapping func(name string) string) string {
	var buf strings.Builder
	for _, v := range s {
		switch v := v.(type) {
		case partLiteral:
			buf.WriteString(string(v))
		case partVariable:
			buf.WriteString(mapping(string(v)))
		}
	}
	return buf.String()
}

type partLiteral string
type partVariable string

var (
	tokenEscapedDollar = []byte("$$")
	tokenBackslash     = []byte("\\")
	tokenVarBegin      = []byte("${")
	tokenVarEnd        = []byte("}")
)

func parse(t [][]byte) String {
	p := &parser{tokens: t}
	p.parse()
	return p.result
}

type parser struct {
	result String
	tokens [][]byte
	pos    int
}

func (p *parser) parse() {
	for {
		t := p.nextToken()
		switch {
		case t == nil:
			return
		case bytes.Equal(t, tokenEscapedDollar):
			p.appendLiteral([]byte("$"))
		case bytes.Equal(t, tokenVarBegin):
			p.parseVar()
		default:
			p.appendLiteral(t)
		}
	}
}

func (p *parser) parseVar() {
	var (
		varName []byte
		literal []byte
	)
	literal = tokenVarBegin
	for {
		t := p.nextToken()
		literal = append(literal, t...)
		switch {
		case t == nil:
			// If the variable is not closed, treat the whole thing as a literal.
			p.appendLiteral(literal)
			return
		case bytes.Equal(t, tokenVarEnd):
			p.appendVariable(varName)
			return
		default:
			varName = append(varName, t...)
		}
	}
}

func (p *parser) nextToken() []byte {
	if p.pos >= len(p.tokens) {
		return nil
	}
	p.pos++
	return p.tokens[p.pos-1]
}

func (p *parser) appendLiteral(v []byte) {
	if len(p.result) == 0 {
		p.result = String{partLiteral(v)}
		return
	}
	// If the last part is a literal, append to it. Having smaller number of
	// parts is better for performance.
	if last, ok := p.result[len(p.result)-1].(partLiteral); ok {
		p.result[len(p.result)-1] = last + partLiteral(v)
		return
	}
	p.result = append(p.result, partLiteral(v))
}

func (p *parser) appendVariable(v []byte) {
	p.result = append(p.result, partVariable(v))
}

func tokenize(s string) [][]byte {
	p := &tokenizer{input: []byte(s)}
	p.tokenize()
	return p.tokens
}

const (
	stateLiteral = iota
	stateVariable
	stateEscapedInLiteral
	stateEscapedInVariable
)

type tokenizer struct {
	input  []byte
	state  int
	pos    int
	tokens [][]byte
	// Buffer for literal values. The bytes are added to it every time
	// a value that is not a token is encountered. The buffer is appended
	// to the tokens slice when a token is encountered.
	literal []byte
}

func (p *tokenizer) tokenize() {
	for {
		switch p.state {
		case stateLiteral:
			if t, ok := p.nextToken(tokenEscapedDollar); ok {
				p.appendToken(t)
				continue
			}
			if _, ok := p.nextToken(tokenBackslash); ok {
				p.state = stateEscapedInLiteral
				continue
			}
			if t, ok := p.nextToken(tokenVarBegin); ok {
				p.appendToken(t)
				p.state = stateVariable
				continue
			}
		case stateVariable:
			if _, ok := p.nextToken(tokenBackslash); ok {
				p.state = stateEscapedInVariable
				continue
			}
			if t, ok := p.nextToken(tokenVarEnd); ok {
				p.appendToken(t)
				p.state = stateLiteral
				continue
			}
		case stateEscapedInLiteral:
			if b, ok := p.nextLiteral(); ok {
				p.appendLiteral(b)
			}
			p.state = stateLiteral
			continue
		case stateEscapedInVariable:
			if b, ok := p.nextLiteral(); ok {
				p.appendLiteral(b)
			}
			p.state = stateVariable
			continue
		}
		if b, ok := p.nextLiteral(); ok {
			p.appendLiteral(b)
			continue
		}
		break
	}
	if len(p.literal) != 0 {
		p.tokens = append(p.tokens, p.literal)
		p.literal = nil
	}
}

func (p *tokenizer) nextToken(token []byte) ([]byte, bool) {
	if p.pos+len(token) > len(p.input) {
		return nil, false
	}
	ok := bytes.Equal(p.input[p.pos:p.pos+len(token)], token)
	if ok {
		p.pos += len(token)
	}
	return token, ok
}

func (p *tokenizer) nextLiteral() (byte, bool) {
	if p.pos >= len(p.input) {
		return 0, false
	}
	p.pos++
	return p.input[p.pos-1], true
}

func (p *tokenizer) appendToken(token []byte) {
	// If literal buffer is not empty, it needs to be appended as a token first.
	if len(p.literal) != 0 {
		p.tokens = append(p.tokens, p.literal, token)
		p.literal = nil
		return
	}
	p.tokens = append(p.tokens, token)
}

func (p *tokenizer) appendLiteral(b byte) {
	p.literal = append(p.literal, b)
}
