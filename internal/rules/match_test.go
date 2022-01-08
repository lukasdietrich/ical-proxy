package rules

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchScope(t *testing.T) {
	data := []byte("test")

	for pattern, expected := range map[string]bool{
		`te.t`:     true,
		`\d+`:      false,
		`(?i)TEST`: true,
	} {
		assert.Equal(t, expected, MatchScope(&data, regexp.MustCompile(pattern)))
	}
}
