/*
Package parser defines methods for building an running parser combinators.

For actually running the resulting parsers, check out ParseScanner and ParseString.

For building parsers, look at any method returning a Parser.

Example usage:

    p := parser.Sequence(
        parser.Digits(),
        parser.Maybe(
            parser.Sequence(
                parser.Char('.'),
                parser.Digits())))
    result, err := parser.ParseString(p, "1234.567")
*/
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
