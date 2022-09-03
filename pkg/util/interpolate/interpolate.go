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
	p := &parser{in: s}
	p.parse()
	return p.res
}

// String is a parsed string.
type String []part

const (
	litType = iota
	partVar
)

type part struct {
	typ int
	val string
}

// Interpolate replaces variables in the string based on the mapping function.
func (s String) Interpolate(mapping func(name string) string) string {
	var buf strings.Builder
	for _, v := range s {
		switch v.typ {
		case litType:
			buf.WriteString(v.val)
		case partVar:
			buf.WriteString(mapping(v.val))
		}
	}
	return buf.String()
}

var (
	tokenEscapedDollar = "$$"
	tokenBackslash     = "\\"
	tokenVarBegin      = "${"
	tokenVarEnd        = "}"
)

type parser struct {
	in     string
	res    String
	pos    int
	litBuf strings.Builder
	varBuf strings.Builder
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
			p.appendByte(p.nextByte())
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
	p.varBuf.Reset()
	for p.hasNext() {
		switch {
		case p.nextToken(tokenVarEnd):
			p.appendVariable(p.varBuf.String())
			return
		case p.nextToken(tokenBackslash):
			if !p.hasNext() {
				continue
			}
			p.varBuf.WriteByte(p.nextByte())
		default:
			p.varBuf.WriteByte(p.nextByte())
		}
	}
	// Variable not closed. Treat the whole thing as a literal.
	p.appendLiteral(tokenVarBegin)
	p.pos = pos
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

// nextToken returns true if the next token matches the given string and advances
// the position.
func (p *parser) nextToken(s string) bool {
	if strings.HasPrefix(p.in[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

// appendLiteral appends the given string as a literal to the result. Literals
// are not added immediately, but buffered until appendBuffer is called.
func (p *parser) appendLiteral(s string) {
	p.litBuf.WriteString(s)
	return
}

// appendByte appends the given byte as a literal to the result. Literals are
// not added immediately, but buffered until appendBuffer is called.
func (p *parser) appendByte(b byte) {
	p.litBuf.WriteByte(b)
	return
}

// appendVariable appends the given string as a variable name to the result.
func (p *parser) appendVariable(s string) {
	p.appendBuffer()
	p.res = append(p.res, part{typ: partVar, val: s})
}

// appendBuffer checks if literal buffer is not empty and appends it to the result.
func (p *parser) appendBuffer() {
	if p.litBuf.Len() > 0 {
		p.res = append(p.res, part{typ: litType, val: p.litBuf.String()})
		p.litBuf.Reset()
	}
}
