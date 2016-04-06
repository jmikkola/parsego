package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmikkola/parsego/parser"
)

func expectParses(t *testing.T, p parser.Parser, s string) {
	_, err := parser.ParseString(p, s)
	assert.NoError(t, err, "expected successful parse")
}

func expectFails(t *testing.T, p parser.Parser, s string) {
	_, err := parser.ParseString(p, s)
	assert.Error(t, err, "expected error")
}

func TestDigit(t *testing.T) {
	expectParses(t, parser.Digit(), "8")
	expectFails(t, parser.Digit(), "x")
}

func TestLowerLetter(t *testing.T) {
	expectParses(t, parser.LowerLetter(), "x")
	expectFails(t, parser.LowerLetter(), "X")
}

func TestUpperLetter(t *testing.T) {
	expectParses(t, parser.UpperLetter(), "X")
	expectFails(t, parser.UpperLetter(), "x")
}

func TestMaybe1(t *testing.T) {
	result, err := parser.ParseString(parser.Many1(parser.Digit()), "56789xxx")
	assert.NoError(t, err, "Expected successful parse")
	assert.Equal(t, "56789", result)

	_, err2 := parser.ParseString(parser.Many1(parser.Digit()), "xxx")
	assert.Error(t, err2, "Expected an error when the character doesn't match")
}

func TestAlphaNum(t *testing.T) {
	expectParses(t, parser.AlphaNum(), "A")
	expectParses(t, parser.AlphaNum(), "a")
	expectParses(t, parser.AlphaNum(), "9")
	expectFails(t, parser.AlphaNum(), "-")
	expectFails(t, parser.AlphaNum(), ".")
}

func TestManySepBy(t *testing.T) {
	p := parser.ManySepBy(parser.Digits(), parser.Whitespace1())
	result1, err1 := parser.ParseString(p, "")
	assert.NoError(t, err1, "Expected successful parse")
	assert.Equal(t, []interface{}{}, result1)

	result2, err2 := parser.ParseString(p, "1234")
	assert.NoError(t, err2, "Expected successful parse")
	assert.Equal(t, []interface{}{"1234"}, result2)

	result3, err3 := parser.ParseString(p, "12 34   56")
	assert.NoError(t, err3, "Expected successful parse")
	assert.Equal(t, []interface{}{"12", "34", "56"}, result3)
}
