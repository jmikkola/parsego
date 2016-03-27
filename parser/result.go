package parser

type ParseResult interface {
	Matched() bool
	Result() interface{}
	TextRange() TextRange
	Error() error
}

type SuccessRusult struct {
	textRange TextRange
	result    interface{}
}

func (r *SuccessRusult) Matched() bool {
	return true
}

func (r *SuccessRusult) Result() interface{} {
	return r.result
}

func (r *SuccessRusult) TextRange() TextRange {
	return r.textRange
}

func (r *SuccessRusult) Error() error {
	return nil
}

type FailedResult struct {
	textRange TextRange
	err       error
}

func (r *FailedResult) Matched() bool {
	return false
}

func (r *FailedResult) Result() interface{} {
	return nil
}

func (r *FailedResult) TextRange() TextRange {
	return r.textRange
}

func (r *FailedResult) Error() error {
	return r.err
}
