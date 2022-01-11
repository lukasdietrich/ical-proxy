package rfc5545

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// <https://datatracker.ietf.org/doc/html/rfc5545#section-3.1>
//
// Lines of text SHOULD NOT be longer than 75 octets, excluding the line
// break.  Long content lines SHOULD be split into a multiple line
// representations using a line "folding" technique.  That is, a long
// line can be split between any two characters by inserting a CRLF
// immediately followed by a single linear white-space character (i.e.,
// SPACE or HTAB).  Any sequence of CRLF followed immediately by a
// single linear white-space character is ignored (i.e., removed) when
// processing the content type.

const (
	foldingLength    = 75
	maximumRawLength = 4096
	crlf             = "\r\n"
)

var (
	ErrContentLineTooLong = errors.New("content-line too long")
)

type foldingReader struct {
	scanner *bufio.Scanner
	peeked  bool
	buffer  *bytes.Buffer
}

func newFoldingReader(r io.Reader) *foldingReader {
	s := bufio.NewScanner(r)
	s.Buffer(nil, 1024)

	b := bytes.NewBuffer(nil)
	b.Grow(maximumRawLength)

	return &foldingReader{
		scanner: s,
		peeked:  false,
		buffer:  b,
	}
}

func (r *foldingReader) readLine() ([]byte, error) {
	if err := r.unfold(); err != nil {
		return nil, err
	}

	return r.buffer.Bytes(), nil
}

func (r *foldingReader) push(b []byte) error {
	if r.buffer.Len()+len(b) > maximumRawLength {
		return ErrContentLineTooLong
	}

	r.buffer.Write(b)
	return nil
}

func (r *foldingReader) unfold() error {
	r.buffer.Reset()

	if !r.peeked && !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return err
		}

		return io.EOF
	}

	if err := r.push(r.scanner.Bytes()); err != nil {
		return err
	}

	for r.scanner.Scan() {
		r.peeked = true
		b := r.scanner.Bytes()

		if len(b) == 0 || !isWhitespace(b[0]) {
			return nil
		}

		if err := r.push(b[1:]); err != nil {
			return err
		}
	}

	r.peeked = false
	return r.scanner.Err()
}

type foldingWriter struct {
	writer *bufio.Writer
}

func newFoldingWriter(w io.Writer) *foldingWriter {
	return &foldingWriter{
		writer: bufio.NewWriterSize(w, maximumRawLength),
	}
}

func (w *foldingWriter) writeLine(line []byte) error {
	for offset := 0; offset < len(line); offset += foldingLength {
		if offset > 0 {
			if err := w.writer.WriteByte(space); err != nil {
				return err
			}
		}

		length := len(line) - offset
		if length > foldingLength {
			length = foldingLength
		}

		if _, err := w.writer.Write(line[offset : offset+length]); err != nil {
			return err
		}

		if _, err := w.writer.WriteString(crlf); err != nil {
			return err
		}
	}

	return w.writer.Flush()
}
