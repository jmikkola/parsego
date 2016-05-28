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
	"io/ioutil"

	"github.com/jmikkola/parsego/parser/scanner"
)

// ParseString parses the text in a string.
func ParseString(parser Parser, str string) (interface{}, error) {
	result := parser.Parse(scanner.FromString(str))
	return result.Result(), result.Error()
}

// ParseScanner parses the text from a scanner.
// TODO: this currently just reads the whole scanner in to memory
// instead of reading as the text is parsed.
func ParseScanner(parser Parser, reader io.Reader) (interface{}, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ParseString(parser, string(bytes))
}
