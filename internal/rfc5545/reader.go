package rfc5545

import "io"

type Reader interface {
	Read() (*ContentLine, error)
}

type reader struct {
	r    *foldingReader
	line *ContentLine
}

func NewReader(r io.Reader) Reader {
	return &reader{
		r:    newFoldingReader(r),
		line: new(ContentLine),
	}
}

func (r *reader) Read() (*ContentLine, error) {
	rawLine, err := r.r.readLine()
	if err != nil {
		return nil, err
	}

	if err := r.line.unmarshal(rawLine); err != nil {
		return nil, err
	}

	return r.line, nil
}
