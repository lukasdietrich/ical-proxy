package rfc5545

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	ErrUnexpectedEOL  = errors.New("rfc5455: unexpected eol")
	ErrUnexpectedChar = errors.New("rfc5455: unexpected char")
)

func unexpectedCharError(line []byte, i int, expected string) error {
	return fmt.Errorf("%w: %x in %q[%d]: expected %s",
		ErrUnexpectedChar, line[i], line, i, expected)
}

type Param struct {
	Name   []byte
	Values [][]byte
}

type ContentLine struct {
	Name   []byte
	Params []Param
	Value  []byte
}

func (c *ContentLine) unmarshal(line []byte) error {
	const (
		_ = iota
		sName
		sParamName
		sParamValueAny
		sParamValueQuote
		sParamValueUnquote
		sValue
	)

	c.Params = c.Params[:0] // reset params

	state := sName
	offset := 0

	for i, b := range line {
		switch state {
		case sName:
			if b == semicolon || b == colon {
				c.Name = line[offset:i]
				offset = i + 1

				if b == semicolon {
					state = sParamName
				} else {
					state = sValue
				}
			}

		case sParamName:
			if b == equals {
				c.Params = append(c.Params, Param{Name: line[offset:i]})
				offset = i + 1
				state = sParamValueAny
			}

		case sParamValueAny:
			if i == offset && b == dquote {
				state = sParamValueQuote
			} else if b == semicolon || b == colon || b == comma {
				p := &c.Params[len(c.Params)-1]
				p.Values = append(p.Values, line[offset:i])
				offset = i + 1

				if b == semicolon {
					state = sParamName
				} else if b == colon {
					state = sValue
				} else {
					state = sParamValueAny
				}
			} else if !isSafeChar(b) {
				return unexpectedCharError(line, i, "SAFE-CHAR")
			}

		case sParamValueQuote:
			if b == dquote {
				state = sParamValueUnquote
			} else if !isQsafeChar(b) {
				return unexpectedCharError(line, i, "QSAFE-CHAR")
			}

		case sParamValueUnquote:
			if b == semicolon || b == colon || b == comma {
				p := &c.Params[len(c.Params)-1]
				p.Values = append(p.Values, line[offset:i])
				offset = i + 1

				if b == semicolon {
					state = sParamName
				} else if b == colon {
					state = sValue
				} else {
					state = sParamValueAny
				}
			} else {
				return unexpectedCharError(line, i, `";", "," or ":"`)
			}

		case sValue:
			if !isSafeChar(b) {
				return unexpectedCharError(line, i, "SAFE-CHAR")
			}
		}
	}

	if state != sValue {
		return ErrUnexpectedEOL
	}

	c.Value = line[offset:]
	return nil
}

func (c *ContentLine) marshal(b *bytes.Buffer) {
	b.Reset()
	b.Write(c.Name)

	for _, param := range c.Params {
		b.WriteByte(semicolon)
		b.Write(param.Name)
		b.WriteByte(equals)

		for i, value := range param.Values {
			if i > 0 {
				b.WriteByte(comma)
			}

			b.Write(value)
		}
	}

	b.WriteByte(colon)
	b.Write(c.Value)
}
