package rfc5545

import (
	"bytes"
	"io"
)

type Writer interface {
	Write(line *ContentLine) error
}

type writer struct {
	w      *foldingWriter
	buffer *bytes.Buffer
}

func NewWriter(w io.Writer) Writer {
	return &writer{
		w:      newFoldingWriter(w),
		buffer: bytes.NewBuffer(nil),
	}
}

func (w *writer) Write(line *ContentLine) error {
	line.marshal(w.buffer)
	return w.w.writeLine(w.buffer.Bytes())
}
