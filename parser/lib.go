package parser

import (
	"io"
)

func parseWith(parser Parser, scanner Scanner) (interface{}, error) {
	result := parser.Parse(scanner)
	return result.Result(), result.Error()
}

// ParseScanner parses the text from a scanner.
func ParseScanner(parser Parser, reader io.RuneReader) (interface{}, error) {
	return parseWith(parser, FromReader(reader))
}

// ParseString parses the text in a string.
func ParseString(parser Parser, str string) (interface{}, error) {
	return parseWith(parser, FromString(str))
}
