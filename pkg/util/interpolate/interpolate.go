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
	"strings"
)

// Parsed is a parsed string.
type Parsed []part

type Variable struct {
	Name       string // Name of the variable.
	Default    string // Default value if the variable is not set.
	HasDefault bool   // True if the variable has a default value.
}

// Parse parses the string and returns a parsed representation. The variables
// may be interpolated later using the Interpolate method.
//
// The syntax is similar to shell variable expansion. The following rules apply:
//
// - Variables are enclosed in ${...} and may contain any character.
//
// - Variables may have a default value separated by :=, eq. ${VAR:=default}.
//
// - To include a literal $ in the output, escape it with a backslash or
// another $. For example, \$ and $$ are both interpreted as a literal $.
// The latter does not work inside a variable.
//
// - If a variable is not closed, it is treated as a literal.
func Parse(s string) Parsed {
	p := parser{in: s, out: make([]part, 0, 1)}
	p.parse()
	return p.out
}

// Interpolate replaces variables in the string based on the mapping function.
func (s Parsed) Interpolate(mapping func(variable Variable) string) string {
	var buf strings.Builder
	for _, v := range s {
		switch v.typ {
		case litType:
			buf.WriteString(v.literal)
		case varType:
			buf.WriteString(mapping(v.variable))
		}
	}
	return buf.String()
}

// HasVars returns true if the string contain at least one variable.
func (s Parsed) HasVars() bool {
	for _, v := range s {
		if v.typ == varType {
			return true
		}
	}
	return false
}

const (
	litType = iota
	varType
)

type part struct {
	typ      int
	literal  string
	variable Variable
}

var (
	tokenEscapedDollar = "$$"
	tokenBackslash     = "\\"
	tokenVarBegin      = "${"
	tokenVarEnd        = "}"
	tokenDefaultVal    = "-"
)

type parser struct {
	in     string
	out    Parsed
	pos    int
	litBuf strings.Builder
	varBuf strings.Builder
	defBuf strings.Builder
}

func (p *parser) parse() {
	for p.hasNext() {
		switch {
		case p.nextToken(tokenBackslash):
			p.parseBackslash()
		case p.nextToken(tokenEscapedDollar):
			p.appendByte('$')
		case p.nextToken(tokenVarBegin):
			p.parseVariable()
		default:
			// Add all characters to the first character that may start the token.
			p.appendLiteral(p.nextBytesUntilAnyOf("\\$"))
		}
	}
	p.appendBuffer()
}

func (p *parser) parseBackslash() {
	if !p.hasNext() {
		p.appendLiteral(tokenBackslash)
		return
	}
	p.appendByte(p.nextByte())
}

func (p *parser) parseVariable() {
	pos := p.pos
	def := false
	p.varBuf.Reset()
	p.defBuf.Reset()
	for p.hasNext() {
		switch {
		case p.nextToken(tokenBackslash):
			if !p.hasNext() {
				continue
			}
			p.varBuf.WriteByte(p.nextByte())
		case p.nextToken(tokenDefaultVal):
			def = true
			p.parseDefault()
		case p.nextToken(tokenVarEnd):
			p.appendVariable(Variable{
				Name:       p.varBuf.String(),
				Default:    p.defBuf.String(),
				HasDefault: def,
			})
			return
		default:
			// Add all characters to the first character that may start the token.
			p.varBuf.WriteString(p.nextBytesUntilAnyOf("-\\}"))
		}
	}
	// Variable not closed. Treat the whole thing as a literal.
	p.appendLiteral(tokenVarBegin)
	p.pos = pos
}

func (p *parser) parseDefault() {
	for p.hasNext() {
		switch {
		case p.nextToken(tokenBackslash):
			if !p.hasNext() {
				continue
			}
			p.defBuf.WriteByte(p.nextByte())
		case p.nextToken(tokenVarEnd):
			// Move the position back so that the closing token is not
			// consumed and can be parsed by the parent parser.
			p.pos -= len(tokenVarEnd)
			return
		default:
			// Add all characters to the first character that may start the token.
			p.defBuf.WriteString(p.nextBytesUntilAnyOf("\\}"))
		}
	}
}

// hasNext returns true if there are more bytes to read.
func (p *parser) hasNext() bool {
	return p.pos < len(p.in)
}

// nextByte returns the next byte and advances the position.
func (p *parser) nextByte() byte {
	p.pos++
	return p.in[p.pos-1]
}

// nextBytesUntilAnyOf returns the next bytes until any of the given characters
// is encountered but not less than one. Only ASCII characters are supported.
func (p *parser) nextBytesUntilAnyOf(s string) string {
	pos := p.pos
	p.pos++
	for p.pos < len(p.in) {
		for n := 0; n < len(s); n++ {
			if p.in[p.pos] == s[n] {
				return p.in[pos:p.pos]
			}
		}
		p.pos++
	}
	return p.in[pos:p.pos]
}

// nextToken returns true if the next token matches the given string and advances
// the position.
func (p *parser) nextToken(s string) bool {
	if strings.HasPrefix(p.in[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

// appendVariable appends the given string as a variable name to the result.
func (p *parser) appendVariable(v Variable) {
	p.appendBuffer()
	p.out = append(p.out, part{typ: varType, variable: v})
}

// appendLiteral appends the given string as a literal to the result. Literals
// are not added immediately, but buffered until appendBuffer is called.
func (p *parser) appendLiteral(s string) {
	p.litBuf.WriteString(s)
}

// appendByte appends the given byte as a literal to the result. Literals are
// not added immediately, but buffered until appendBuffer is called.
func (p *parser) appendByte(b byte) {
	p.litBuf.WriteByte(b)
}

// appendBuffer checks if literal buffer is not empty and appends it to the result.
func (p *parser) appendBuffer() {
	if p.litBuf.Len() > 0 {
		p.out = append(p.out, part{typ: litType, literal: p.litBuf.String()})
		p.litBuf.Reset()
	}
}
