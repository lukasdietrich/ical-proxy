package rfc5545

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFoldingReader(t *testing.T) {
	var raw = "normal line\r\n" +
		"folded\r\n" +
		"  with\r\n" +
		" and without extra space\r\n" +
		"final line\r\n"

	r := newFoldingReader(strings.NewReader(raw))

	for _, expected := range []string{
		"normal line",
		"folded withand without extra space",
		"final line",
	} {
		line, err := r.readLine()

		assert.NoError(t, err)
		assert.Equal(t, expected, string(line))
	}

	_, err := r.readLine()
	assert.ErrorIs(t, err, io.EOF)
}

func TestFoldingWriter(t *testing.T) {
	var expected = "short line\r\n" +
		"this is a very long .......................................................\r\n" +
		" ........ line\r\n"

	var actual bytes.Buffer

	w := newFoldingWriter(&actual)

	for _, line := range []string{
		"short line",
		"this is a very long ............................................................... line",
	} {
		assert.NoError(t, w.writeLine([]byte(line)))
	}

	assert.Equal(t, expected, actual.String())
}
