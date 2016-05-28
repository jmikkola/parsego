package result

import (
	"fmt"

	"github.com/jmikkola/parsego/parser/textpos"
)

// ParseResult defines the values that can result from parsing.
type ParseResult interface {
	Matched() bool
	Result() interface{}
	TextRange() textpos.TextRange
	Error() error
}

// Success returns a SuccessResult
func Success(textRange textpos.TextRange, result interface{}) ParseResult {
	return &SuccessResult{textRange, result}
}

// SuccessResult is returned by parsers on a successful parser.
type SuccessResult struct {
	textRange textpos.TextRange
	result    interface{}
}

// Matched returns whether the parser matched the input (true in this
// case).
func (r *SuccessResult) Matched() bool {
	return true
}

// Result returns the parsed value (usually a string).
func (r *SuccessResult) Result() interface{} {
	return r.result
}

// TextRange returns the range of text parsed.
func (r *SuccessResult) TextRange() textpos.TextRange {
	return r.textRange
}

// Error returns the reason for failing (nil in this case).
func (r *SuccessResult) Error() error {
	return nil
}

// FailedResult is returned by parsers when they fail to parse the
// input.
type FailedResult struct {
	textRange textpos.TextRange
	err       error
}

// Failed returns a FailedResult
func Failed(textRange textpos.TextRange, err error) ParseResult {
	return &FailedResult{textRange, err}
}

// Matched returns whether the parser matched the input (false in this
// case).
func (r *FailedResult) Matched() bool {
	return false
}

// Result returns the parsed value (always nil, in this case).
func (r *FailedResult) Result() interface{} {
	return nil
}

// TextRange returns the range of text parsed.
func (r *FailedResult) TextRange() textpos.TextRange {
	return r.textRange
}

// Error returns the reason for failing.
func (r *FailedResult) Error() error {
	end := r.TextRange().End()
	return fmt.Errorf("%v at line %d, col %d", r.err, end.Line(), end.Col())
}
