package scanner

import (
	"github.com/jmikkola/parsego/parser/textpos"
)

// EOFError is returned when RuneBacktrackingScanner reaches the end of a string.
type EOFError struct{}

// Error() renders and EOFError to a string.
func (e *EOFError) Error() string {
	return "Reached end of input"
}

// ReadRune defines a simpler version of io.RuneReader.
type ReadRune interface {
	Read() (rune, error)
}

// Scanner provides a way to scan input with the ability to undo
// reading some input (with multiple levels of undo).
type Scanner interface {
	ReadRune
	GetPos() textpos.TextPos
	StartSnapshot()
	RewindSnapshot()
	PopSnapshot()
}

// snapshot records the state of a snapshot taken by a scanner.
type snapshot struct {
	idx        int
	currentPos textpos.TextPos
	next       *snapshot
}

// StringScanner is an implementation of Scanner.
type StringScanner struct {
	rs         []rune
	idx        int
	currentPos textpos.TextPos
	lastSnap   *snapshot
}

// FromString creates a Scanner from a string.
func FromString(str string) Scanner {
	return &StringScanner{
		rs:         []rune(str),
		currentPos: textpos.StartingPos(),
	}
}

// Read a rune if one is available, otherwise return an EOFError.
func (self *StringScanner) Read() (rune, error) {
	var r rune
	if self.idx >= len(self.rs) {
		return r, &EOFError{}
	}

	r = self.rs[self.idx]
	self.idx++
	self.currentPos = self.currentPos.Advance(r)
	return r, nil
}

// GetPos returns the position of the next character Read() will
// return.
func (self *StringScanner) GetPos() textpos.TextPos {
	return self.currentPos
}

// StartSnapshot takes a new snapshot that can be rolled back to
// later.
func (self *StringScanner) StartSnapshot() {
	self.lastSnap = &snapshot{
		idx:        self.idx,
		currentPos: self.currentPos,
		next:       self.lastSnap,
	}
}

// RewindSnapshot reverts the scanner back to the state it was in when
// StartSnapshot() was last called.
func (s *StringScanner) RewindSnapshot() {
	if s.lastSnap == nil {
		panic("Bug: rewinding to a snapshot that was never started")
	}

	s.currentPos = s.lastSnap.currentPos
	s.idx = s.lastSnap.idx
	s.lastSnap = s.lastSnap.next
}

// PopSnapshot drops a snapshot when it is no longer needed.
func (s *StringScanner) PopSnapshot() {
	if s.lastSnap == nil {
		panic("Bug: popped a snapshot that was never started")
	}
	s.lastSnap = s.lastSnap.next
}
