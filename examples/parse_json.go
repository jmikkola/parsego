package main

import (
	"fmt"
	"strconv"

	"github.com/jmikkola/parsego/parser"
)

func charWHS(c rune) parser.Parser {
	return parser.Sequence(
		parser.Whitespace(),
		parser.Char(c),
		parser.Whitespace())
}

func jsonString() parser.Parser {
	hexChar := parser.Or(
		parser.CharRange('a', 'f'),
		parser.CharRange('A', 'F'),
		parser.Digit())
	unicodeEscapeSeq := parser.Sequence(
		parser.Char('u'), hexChar, hexChar, hexChar)

	escapedChar := parser.Sequence(
		parser.Char('\\'),
		parser.Or(
			parser.AnyChar('b', 'f', 'n', 'r', 't', '\\', '/', '"'),
			unicodeEscapeSeq))

	stringChars := parser.Many(parser.Or(
		parser.NoneOf('"'), escapedChar))
	quote := parser.Char('"')

	return parser.ParseWith(
		parser.Sequence(quote, stringChars, quote),
		func(str interface{}) interface{} {
			t, _ := strconv.Unquote(str.(string))
			return t
		})
}

type pair struct {
	key   string
	value interface{}
}

func objectPair() parser.Parser {
	return parser.Map([]parser.Named{
		{"key", jsonString()},
		{"", charWHS(':')},
		{"value", jsonParser()},
	}, func(m map[string]interface{}) interface{} {
		return pair{
			key:   m["key"].(string),
			value: m["value"],
		}
	})
}

func listParser() parser.Parser {
	values := parser.ManySepBy(jsonParser(), charWHS(','))
	return parser.Surround(charWHS('['), values, charWHS(']'))
}

func floatParser() parser.Parser {
	decimalPart := parser.Sequence(
		parser.Char('.'),
		parser.Digits())

	exponentPart := parser.Sequence(
		parser.AnyChar('e', 'E'),
		parser.Maybe(parser.AnyChar('-', '+')),
		parser.Digits())

	floatP := parser.Sequence(
		parser.Maybe(parser.Char('-')),
		parser.Digits(),
		parser.Maybe(decimalPart),
		parser.Maybe(exponentPart))

	return parser.ParseWith(
		floatP,
		func(floatVal interface{}) interface{} {
			f, _ := strconv.ParseFloat(floatVal.(string), 64)
			return f
		})
}

func objectParser() parser.Parser {
	pairs := parser.ManySepBy(objectPair(), charWHS(','))

	return parser.ParseWith(
		parser.Surround(charWHS('{'), pairs, charWHS('}')),
		func(ps interface{}) interface{} {
			object := map[string]interface{}{}
			for _, part := range ps.([]interface{}) {
				pair, _ := part.(pair)
				object[pair.key] = pair.value
			}
			return object
		})
}

func jsonParser() parser.Parser {
	trueParser := parser.TokenAs("true", true)
	falseParser := parser.TokenAs("false", false)
	nullParser := parser.TokenAs("null", nil)
	return parser.Lazy(func() parser.Parser {
		return parser.Or(
			objectParser(), listParser(), trueParser, nullParser,
			falseParser, floatParser(), jsonString())
	})
}

func main() {
	// See http://www.json.org/
	json := `{"a key": -123.45E+8, "b": [true, false, null], "c": {"in\\ner": "yup"}}`
	result, err := parser.ParseString(jsonParser(), json)
	if err != nil {
		fmt.Println("error parsing", err)
	} else {
		fmt.Println("parsed", result)
	}
}
