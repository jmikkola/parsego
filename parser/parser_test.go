package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmikkola/parsego/parser"
)

func TestParseEOF(t *testing.T) {
	_, err := parser.ParseString(parser.EOF(), "")
	assert.NoError(t, err, "Expected successful parse")

	_, err = parser.ParseString(parser.EOF(), "not EOF just yet...")
	assert.Error(t, err, "Expected an error")
}

func TestAnyChar(t *testing.T) {
	result, err := parser.ParseString(parser.AnyChar('a', '\n', '☃'), "☃")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "☃", result)

	_, err2 := parser.ParseString(parser.AnyChar('a', '\n', '☃'), "x")
	assert.Error(t, err2, "Expected an error when the character doesn't match")

	_, err3 := parser.ParseString(parser.AnyChar('a', '\n', '☃'), "")
	assert.Error(t, err3, "Expected an error when the input ends")
}

func TestAnyCharIn(t *testing.T) {
	result, err := parser.ParseString(parser.AnyCharIn("a\n☃"), "☃")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "☃", result)
}

func TestChar(t *testing.T) {
	result, err := parser.ParseString(parser.Char('☃'), "☃")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "☃", result)

	_, err2 := parser.ParseString(parser.Char('☃'), "x")
	assert.Error(t, err2, "Expected an error when the character doesn't match")

	_, err3 := parser.ParseString(parser.Char('☃'), "")
	assert.Error(t, err3, "Expected an error when the input ends")
}

func TestCharRange(t *testing.T) {
	result, err := parser.ParseString(parser.CharRange('0', '9'), "5")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "5", result)

	_, err2 := parser.ParseString(parser.CharRange('0', '9'), "x")
	assert.Error(t, err2, "Expected an error when the character doesn't match")

	_, err3 := parser.ParseString(parser.CharRange('0', '9'), "")
	assert.Error(t, err3, "Expected an error when the input ends")
}

func TestParseSequence(t *testing.T) {
	p := parser.Sequence(
		parser.AnyChar('a'),
		parser.AnyChar('b'),
		parser.EOF())
	result, err := parser.ParseString(p, "ab")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "ab", result)

	_, err2 := parser.ParseString(p, "aa")
	assert.Error(t, err2, "Expected error when string doesn't match")
}

type runepair struct {
	a, b rune
}

func TestWrapper(t *testing.T) {
	ab := parser.Sequence(parser.AnyChar('a'), parser.AnyChar('☃'))
	p := parser.ParseWith(ab, func(val interface{}) interface{} {
		rs := []rune(val.(string))
		return runepair{a: rs[0], b: rs[1]}
	})
	result, err := parser.ParseString(p, "a☃")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, runepair{'a', '☃'}, result)

	_, err2 := parser.ParseString(p, "x")
	assert.Error(t, err2, "Expected error when string doesn't match")
}

func TestParseToken(t *testing.T) {
	floatTok := parser.Token("float64")
	result, err := parser.ParseString(floatTok, "float64")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "float64", result)

	_, err1 := parser.ParseString(floatTok, "xfloat64")
	assert.Error(t, err1, "Expected error")

	_, err2 := parser.ParseString(floatTok, "")
	assert.Error(t, err2, "Expected error")
}

func TestMaybe(t *testing.T) {
	p := parser.Sequence(parser.Maybe(parser.Char('a')), parser.Char('b'))

	result1, err1 := parser.ParseString(p, "ab")
	assert.NoError(t, err1, "Expected successful parse")
	assert.Equal(t, "ab", result1)

	result2, err2 := parser.ParseString(p, "b")
	assert.NoError(t, err2, "Expected successful parse")
	assert.Equal(t, "b", result2)
}

func TestOr(t *testing.T) {
	p := parser.Or(
		parser.Token("int32"), parser.Token("int64"),
		parser.Token("float32"), parser.Token("float64"))

	result1, err1 := parser.ParseString(p, "float32")
	assert.NoError(t, err1, "Expected successful parse")
	assert.Equal(t, "float32", result1)

	_, err2 := parser.ParseString(p, "foo")
	assert.Error(t, err2, "Expected error when no options match")
}

func TestMany(t *testing.T) {
	p := parser.Many(parser.Digit())

	result, err := parser.ParseString(p, "1234x")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "1234", result)
}

func TestListOf(t *testing.T) {
	p := parser.ListOf(parser.Digit())

	result, err := parser.ParseString(p, "1234x")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, []interface{}{"1", "2", "3", "4"}, result)
}

func TestMap(t *testing.T) {
	p := parser.Map([]parser.Named{
		{"value", parser.Many(parser.Letter())},
		{"", parser.Char('[')},
		{"index", parser.Many(parser.Digit())},
		{"", parser.Char(']')},
	}, func(m map[string]interface{}) interface{} {
		return []interface{}{m["value"], m["index"]}
	})

	result, err := parser.ParseString(p, "myVar[123]")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, []interface{}{"myVar", "123"}, result)
}
