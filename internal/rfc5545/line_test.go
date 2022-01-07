package rfc5545

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentLineMarshalUnmarshal(t *testing.T) {
	var (
		actualBuffer bytes.Buffer
		actualLine   ContentLine
	)

	for expectedText, expectedLine := range map[string]ContentLine{
		`BEGIN:VCALENDAR`: ContentLine{
			Name:  []byte(`BEGIN`),
			Value: []byte(`VCALENDAR`),
		},
		`PRODID:-//xyz Corp//NONSGML PDA Calendar Version 1.0//EN`: ContentLine{
			Name:  []byte(`PRODID`),
			Value: []byte(`-//xyz Corp//NONSGML PDA Calendar Version 1.0//EN`),
		},
		`DTSTART;TZID=America/New_York:19980312T083000`: ContentLine{
			Name: []byte(`DTSTART`),
			Params: []Param{
				{
					Name:   []byte(`TZID`),
					Values: [][]byte{[]byte(`America/New_York`)},
				},
			},
			Value: []byte(`19980312T083000`),
		},
		`DTSTART;TZID="America;New_York":19980312T083000`: ContentLine{
			Name: []byte(`DTSTART`),
			Params: []Param{
				{
					Name:   []byte(`TZID`),
					Values: [][]byte{[]byte(`"America;New_York"`)},
				},
			},
			Value: []byte(`19980312T083000`),
		},
	} {
		expectedLine.marshal(&actualBuffer)
		assert.Equal(t, expectedText, actualBuffer.String())

		// depending on the order the map is iterated, Params may have been truncated to Params[:0]
		// which assert.Equal interprets as unequal to nil
		actualLine.Params = nil

		err := actualLine.unmarshal([]byte(expectedText))
		assert.NoError(t, err)
		assert.Equal(t, expectedLine, actualLine)
	}
}

func TestContentLineError(t *testing.T) {
	for text, expected := range map[string]error{
		"BEGIN":             ErrUnexpectedEOL,
		"NAME;PARAM":        ErrUnexpectedEOL,
		"NAME;PARAM=\"\b\"": ErrUnexpectedChar,
		"NAME;PARAM=\b":     ErrUnexpectedChar,
	} {
		var line ContentLine
		err := line.unmarshal([]byte(text))
		assert.ErrorIs(t, err, expected)
	}
}
