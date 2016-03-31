package parser

import (
	"bytes"
	"fmt"
)

// Parser defines the interface implemented by all combinable parsers.
type Parser interface {
	Parse(sc Scanner) ParseResult
}

func fail(at TextPos, format string, a ...interface{}) ParseResult {
	return &FailedResult{
		err:       fmt.Errorf(format, a...),
		textRange: TextRange{at, at},
	}
}

// EOFParser expects just EOF.
type EOFParser struct{}

// EOF returns a parser that expects just EOF.
func EOF() Parser {
	return &EOFParser{}
}

// Parse parses the input.
func (p *EOFParser) Parse(sc Scanner) ParseResult {
	r, err := sc.Read()
	if err == nil {
		return fail(sc.GetPos(), "expected EOF, got %c", r)
	}
	return &SuccessResult{
		textRange: TextRange{sc.GetPos(), sc.GetPos()},
		result:    "",
	}
}

// CharRangeParser parses any single character in a range, inclusive.
type CharRangeParser struct {
	min rune // both inclusive
	max rune
}

// Char returns a parser that parses a single occurance of that rune.
func Char(c rune) Parser {
	return &CharRangeParser{c, c}
}

// CharRange returns a parser that parses a single occurance of any
// rune in the given range, inclusive.
func CharRange(min, max rune) Parser {
	return &CharRangeParser{min, max}
}

// Parse parses the input.
func (p *CharRangeParser) Parse(sc Scanner) ParseResult {
	start := sc.GetPos()
	r, err := sc.Read()
	if err != nil {
		return fail(sc.GetPos(), "expected a character, got error %v", err)
	}
	if r < p.min || r > p.max {
		return fail(sc.GetPos(), "expected a character in the range, got error %c", r)
	}
	return &SuccessResult{
		textRange: TextRange{start, sc.GetPos()},
		result:    r,
	}
}

// TokenParser works like a series of CharRangeParsers, but is more
// efficient.
type TokenParser struct {
	token string
}

// Token returns a parser that parses the exact string given.
func Token(token string) Parser {
	return &TokenParser{token}
}

// Parse parses the input.
func (p *TokenParser) Parse(sc Scanner) ParseResult {
	start := sc.GetPos()
	seen := []rune{}
	for _, c := range p.token {
		r, err := sc.Read()
		seen = append(seen, r)
		if err != nil {
			return fail(sc.GetPos(), "expected '%s', got error %v", p.token, err)
		}
		if r != c {
			return fail(sc.GetPos(), "expected '%s', got '%s'", p.token, string(seen))
		}
	}
	return &SuccessResult{
		textRange: TextRange{start, sc.GetPos()},
		result:    string(seen),
	}
}

// CharSetParser parses any single character in the set.
type CharSetParser struct {
	allowed map[rune]struct{}
}

// AnyCharIn returns a parser that parses a single occurance of any
// rune in the given string.
func AnyCharIn(s string) Parser {
	allowed := make(map[rune]struct{}, len(s))
	for _, r := range s {
		allowed[r] = struct{}{}
	}
	return &CharSetParser{allowed}
}

// AnyChar returns a parser that parses a single occurance of any rune
// given.
func AnyChar(rs ...rune) Parser {
	allowed := make(map[rune]struct{}, len(rs))
	for _, r := range rs {
		allowed[r] = struct{}{}
	}
	return &CharSetParser{allowed}
}

// Parse parses the input.
func (p *CharSetParser) Parse(sc Scanner) ParseResult {
	start := sc.GetPos()
	r, err := sc.Read()
	if err != nil {
		return fail(sc.GetPos(), "expected a character, got error %v", err)
	}
	if _, ok := p.allowed[r]; !ok {
		return fail(sc.GetPos(), "expected a character in the set, got error %c", r)
	}
	return &SuccessResult{
		textRange: TextRange{start, sc.GetPos()},
		result:    r,
	}
}

// SeqParser combines multiple parsers in sequence.
type SeqParser struct {
	parsers []Parser
}

// Sequence returns a parser that runs each given parser in series and
// combines the result.
func Sequence(parsers ...Parser) Parser {
	return &SeqParser{parsers}
}

// Parse parses the input.
func (p *SeqParser) Parse(sc Scanner) ParseResult {
	var textRange TextRange
	textRange.start = sc.GetPos()
	results := []interface{}{}

	for _, inner := range p.parsers {
		innerResult := inner.Parse(sc)
		// Return errors right away
		if !innerResult.Matched() {
			return innerResult
		}

		textRange.end = innerResult.TextRange().end
		results = append(results, innerResult.Result())
	}

	return &SuccessResult{
		textRange: textRange,
		result:    cleanupResult(results),
	}
}

func cleanupResult(results []interface{}) interface{} {
	var buffer bytes.Buffer
	allStr := true
	for _, result := range results {
		if result == "" {
			continue
		}
		if r, ok := result.(rune); ok {
			buffer.WriteRune(r)
		} else if s, ok := result.(string); ok {
			buffer.WriteString(s)
		} else {
			allStr = false
			break
		}
	}
	if allStr {
		return buffer.String()
	}
	return results
}

// Wrapper modifies the result of a parser with a function.
type Wrapper struct {
	inner Parser
	fn    func(interface{}) interface{}
}

// ParseWith returns a parser that will apply the given function to
// the result of parsing, if the parser was successful.
func ParseWith(p Parser, fn func(interface{}) interface{}) Parser {
	return &Wrapper{inner: p, fn: fn}
}

// Parse parses the input.
func (p *Wrapper) Parse(sc Scanner) ParseResult {
	innerResult := p.inner.Parse(sc)
	if innerResult.Matched() {
		return &SuccessResult{
			textRange: innerResult.TextRange(),
			result:    p.fn(innerResult.Result()),
		}
	}
	return innerResult
}

// MaybeParser tries to run the inner parser, but allows the inner
// parser to fail.
type MaybeParser struct {
	inner Parser
}

// Maybe returns a parser that parses 0 or 1 occurances of the given
// parser.
func Maybe(inner Parser) Parser {
	return &MaybeParser{inner}
}

// Parse parses the input.
func (p *MaybeParser) Parse(sc Scanner) ParseResult {
	sc.StartSnapshot()

	innerResult := p.inner.Parse(sc)
	if innerResult.Matched() {
		sc.PopSnapshot()
		return innerResult
	}

	sc.RewindSnapshot()
	start := sc.GetPos()
	return &SuccessResult{
		textRange: TextRange{start, start},
		result:    "",
	}
}

// ManyParser Matches 0+ occurrences
type ManyParser struct {
	inner Parser
}

// Many returns a parser that matches the given parser zero or more
// times, and combines the results.
func Many(inner Parser) Parser {
	return &ManyParser{inner}
}

// Parse parses the input.
func (p *ManyParser) Parse(sc Scanner) ParseResult {
	var textRange TextRange
	textRange.start = sc.GetPos()
	results := []interface{}{}

	for true {
		sc.StartSnapshot()
		innerResult := p.inner.Parse(sc)

		if innerResult.Matched() {
			sc.PopSnapshot()
			results = append(results, innerResult.Result())
		} else {
			sc.RewindSnapshot()
			break
		}
	}

	textRange.end = sc.GetPos()
	return &SuccessResult{
		textRange: textRange,
		result:    cleanupResult(results),
	}
}

// OrParser parses at most one of the inner parses.
type OrParser struct {
	parsers []Parser
}

// Or returns a parser that accepts the union of the languages
// accepted by the given parsers.
func Or(parsers ...Parser) Parser {
	return &OrParser{parsers}
}

// Parse parses the input.
func (p *OrParser) Parse(sc Scanner) ParseResult {
	for _, inner := range p.parsers {
		sc.StartSnapshot()
		innerResult := inner.Parse(sc)

		if innerResult.Matched() {
			sc.PopSnapshot()
			return innerResult
		}
		sc.RewindSnapshot()
	}

	return fail(sc.GetPos(), "no parser matched")
}
