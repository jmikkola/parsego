package parser

func Digit() Parser {
	return CharRange('0', '9')
}

func LowerLetter() Parser {
	return CharRange('a', 'z')
}

func UpperLetter() Parser {
	return CharRange('A', 'Z')
}

func Letter() Parser {
	return Or(LowerLetter(), UpperLetter())
}

func AlphaNum() Parser {
	return Or(LowerLetter(), UpperLetter(), Digit())
}

// Many1 makes 1+ occurrences
func Many1(inner Parser) Parser {
	return Sequence(inner, Many(inner))
}

func Digits() Parser {
	return Many1(Digit())
}
