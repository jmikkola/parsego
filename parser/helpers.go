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

// Digits parses one or more digits.
func Digits() Parser {
	return Many1(Digit())
}
