package parser

import (
	"io"
)

func parseWith(parser Parser, scanner Scanner) (interface{}, error) {
	result := parser.Parse(scanner)
	return result.Result(), result.Error()
}

func ParseScanner(parser Parser, reader io.RuneReader) (interface{}, error) {
	return parseWith(parser, FromReader(reader))
}

func ParseString(parser Parser, str string) (interface{}, error) {
	return parseWith(parser, FromString(str))
}
