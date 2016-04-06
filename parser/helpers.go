package parser

// Digit parses a single digit.
func Digit() Parser {
	return CharRange('0', '9')
}

// LowerLetter parses a single lower case letter.
func LowerLetter() Parser {
	return CharRange('a', 'z')
}

// UpperLetter parses a single upper case letter.
func UpperLetter() Parser {
	return CharRange('A', 'Z')
}

// Letter parses a single letter (upper or lower case).
func Letter() Parser {
	return Or(LowerLetter(), UpperLetter())
}

// AlphaNum parse a letter or a digit.
func AlphaNum() Parser {
	return Or(LowerLetter(), UpperLetter(), Digit())
}

// Many1 makes 1+ occurrences.
func Many1(inner Parser) Parser {
	return Sequence(inner, Many(inner))
}

// Many1SepBy parses a list of 1+ things separated by some separator.
// E.g. parser.ManySepBy(parser.Digits(), parser.Whitespace1()) would
// parse "123 4   456" as []interface{"123", "4", "456"}
func Many1SepBy(inner, separator Parser) Parser {
	pairs := Map([]Named{
		{"", separator},
		{"inner", inner},
	}, func(m map[string]interface{}) interface{} {
		return m["inner"]
	})
	return Map([]Named{
		{"first", inner},
		{"rest", ListOf(pairs)},
	}, func(m map[string]interface{}) interface{} {
		first := m["first"]
		rest := m["rest"].([]interface{})
		out := make([]interface{}, 1+len(rest))
		out[0] = first
		for i, val := range rest {
			out[i+1] = val
		}
		return out
	})
}

// ManySepBy parses a list of 0+ things separated by some separator.
func ManySepBy(inner, separator Parser) Parser {
	return ParseWith(
		Maybe(Many1SepBy(inner, separator)),
		func(inner interface{}) interface{} {
			// Make sure the return type is a list even when nothing matches
			if _, ok := inner.([]interface{}); ok {
				return inner
			}
			return []interface{}{}
		})
}

// Digits parses one or more digits.
func Digits() Parser {
	return Many1(Digit())
}

// WhitespaceChar parses a single whitespace character
func WhitespaceChar() Parser {
	return AnyChar(' ', '\n', '\t', '\b', '\v')
}

// Whitespace parses zero or more whitespace characters
func Whitespace() Parser {
	return Many(WhitespaceChar())
}

// Whitespace1 parses one or more whitespace characters
func Whitespace1() Parser {
	return Many1(WhitespaceChar())
}
