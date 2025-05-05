package rfc5545

const (
	space     = ' '
	htab      = '\t'
	dquote    = '"'
	semicolon = ';'
	colon     = ':'
	comma     = ','
	equals    = '='
)

func isWhitespace(b byte) bool {
	return b == space || b == htab
}

func isQsafeChar(b byte) bool {
	return isValueChar(b) && b != dquote
}

func isSafeChar(b byte) bool {
	return isQsafeChar(b) && b != semicolon && b != colon && b != comma
}

func isValueChar(b byte) bool {
	return !isControl(b)
}

func isControl(b byte) bool {
	return b <= 0x08 || (b >= 0x0a && b <= 0x1f) || b == 0x7f
}
