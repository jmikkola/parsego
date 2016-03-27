package parser

import (
	"bytes"
	"fmt"
)

type Parser interface {
	Parse(sc Scanner) ParseResult
}

func fail(at TextPos, format string, a ...interface{}) ParseResult {
	return &FailedResult{
		err:       fmt.Errorf(format, a...),
		textRange: TextRange{at, at},
	}
}

type EOFParser struct{}

func EOF() Parser {
	return &EOFParser{}
}

func (p *EOFParser) Parse(sc Scanner) ParseResult {
	r, err := sc.Read()
	if err == nil {
		return fail(sc.GetPos(), "expected EOF, got %c", r)
	}
	return &SuccessRusult{
		textRange: TextRange{sc.GetPos(), sc.GetPos()},
		result:    "",
	}
}

type CharRangeParser struct {
	min rune // both inclusive
	max rune
}

func Char(c rune) Parser {
	return &CharRangeParser{c, c}
}

func CharRange(min, max rune) Parser {
	return &CharRangeParser{min, max}
}

func (p *CharRangeParser) Parse(sc Scanner) ParseResult {
	start := sc.GetPos()
	r, err := sc.Read()
	if err != nil {
		return fail(sc.GetPos(), "expected a character, got error %v", err)
	}
	if r < p.min || r > p.max {
		return fail(sc.GetPos(), "expected a character in the range, got error %c", r)
	}
	return &SuccessRusult{
		textRange: TextRange{start, sc.GetPos()},
		result:    r,
	}
}

// TokenParser works like a series of CharRangeParsers, but is more efficient
type TokenParser struct {
	token string
}

func Token(token string) Parser {
	return &TokenParser{token}
}

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
	return &SuccessRusult{
		textRange: TextRange{start, sc.GetPos()},
		result:    string(seen),
	}
}

type CharSetParser struct {
	allowed map[rune]struct{}
}

func AnyCharIn(s string) Parser {
	allowed := make(map[rune]struct{}, len(s))
	for _, r := range s {
		allowed[r] = struct{}{}
	}
	return &CharSetParser{allowed}
}

func AnyChar(rs ...rune) Parser {
	allowed := make(map[rune]struct{}, len(rs))
	for _, r := range rs {
		allowed[r] = struct{}{}
	}
	return &CharSetParser{allowed}
}

func (p *CharSetParser) Parse(sc Scanner) ParseResult {
	start := sc.GetPos()
	r, err := sc.Read()
	if err != nil {
		return fail(sc.GetPos(), "expected a character, got error %v", err)
	}
	if _, ok := p.allowed[r]; !ok {
		return fail(sc.GetPos(), "expected a character in the set, got error %c", r)
	}
	return &SuccessRusult{
		textRange: TextRange{start, sc.GetPos()},
		result:    r,
	}
}

type SeqParser struct {
	parsers []Parser
}

func Sequence(parsers ...Parser) Parser {
	return &SeqParser{parsers}
}

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

	return &SuccessRusult{
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

type Wrapper struct {
	inner Parser
	fn    func(interface{}) interface{}
}

func ParseWith(p Parser, fn func(interface{}) interface{}) Parser {
	return &Wrapper{inner: p, fn: fn}
}

func (p *Wrapper) Parse(sc Scanner) ParseResult {
	innerResult := p.inner.Parse(sc)
	if innerResult.Matched() {
		return &SuccessRusult{
			textRange: innerResult.TextRange(),
			result:    p.fn(innerResult.Result()),
		}
	} else {
		return innerResult
	}
}

type MaybeParser struct {
	inner Parser
}

func Maybe(inner Parser) Parser {
	return &MaybeParser{inner}
}

func (p *MaybeParser) Parse(sc Scanner) ParseResult {
	sc.StartSnapshot()

	innerResult := p.inner.Parse(sc)
	if innerResult.Matched() {
		sc.PopSnapshot()
		return innerResult
	}

	sc.RewindSnapshot()
	start := sc.GetPos()
	return &SuccessRusult{
		textRange: TextRange{start, start},
		result:    "",
	}
}

// ManyParser Matches 0+ occurrences
type ManyParser struct {
	inner Parser
}

func Many(inner Parser) Parser {
	return &ManyParser{inner}
}

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
	return &SuccessRusult{
		textRange: textRange,
		result:    cleanupResult(results),
	}
}

type OrParser struct {
	parsers []Parser
}

func Or(parsers ...Parser) Parser {
	return &OrParser{parsers}
}

func (p *OrParser) Parse(sc Scanner) ParseResult {
	for _, inner := range p.parsers {
		sc.StartSnapshot()
		innerResult := inner.Parse(sc)

		if innerResult.Matched() {
			sc.PopSnapshot()
			return innerResult
		} else {
			sc.RewindSnapshot()
		}
	}

	return fail(sc.GetPos(), "no parser matched")
}
