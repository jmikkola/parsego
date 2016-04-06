package scanner

import (
	"io"

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

// BacktrackingScanner defines methods for a scanner that supports
// buffering and walking back over input.
type BacktrackingScanner interface {
	ReadRune
	StartBuffer()
	ReadBuffer(int) rune
	BufSize() int
	DropBuffer()
}

// RuneBacktrackingScanner implements BacktrackingScanner for a slice
// of runes.
type RuneBacktrackingScanner struct {
	runes        []rune
	pos          int
	recordOffset int
	recording    bool
}

// RBSFromString creates a RuneBacktrackingScanner from a string.
func RBSFromString(s string) *RuneBacktrackingScanner {
	return &RuneBacktrackingScanner{
		runes:        []rune(s),
		pos:          0,
		recordOffset: 0,
		recording:    false,
	}
}

// Read a rune if one is available, otherwise return an EOFError.
func (s *RuneBacktrackingScanner) Read() (rune, error) {
	if s.pos >= len(s.runes) {
		return 0, &EOFError{}
	}
	r := s.runes[s.pos]
	s.pos++
	return r, nil
}

// StartBuffer starts recording runs returned from Read() into the
// buffer, if not already recording.
func (s *RuneBacktrackingScanner) StartBuffer() {
	// Not much to do, the entire string is saved anyway.
	// Just save the offset for math later.
	if !s.recording {
		s.recordOffset = s.pos
		s.recording = true
	}
}

// ReadBuffer reads a buffered rune out of a buffer. The buffer is
// empty unless StartBuffer() then one or more calls to Read()
// happens.
func (s *RuneBacktrackingScanner) ReadBuffer(bufIdx int) rune {
	return s.runes[s.recordOffset+bufIdx]
}

// DropBuffer stops recording and empties the buffer.
func (s *RuneBacktrackingScanner) DropBuffer() {
	s.recordOffset = 0
	s.recording = false
}

// BufSize returns the current size of the buffer.
func (s *RuneBacktrackingScanner) BufSize() int {
	if s.recording {
		return s.pos - s.recordOffset
	}
	return 0
}

// ScannerBacktrackingScanner implements BacktrackingScanner for an
// io.RuneReader.
type ScannerBacktrackingScanner struct {
	wrapped      io.RuneReader
	buffer       []rune
	bufferOffset int
	recording    bool
}

// SBSFromReader creates a ScannerBacktrackingScanner from an
// io.RuneReader.
func SBSFromReader(reader io.RuneReader) *ScannerBacktrackingScanner {
	return &ScannerBacktrackingScanner{
		wrapped:      reader,
		buffer:       []rune{},
		bufferOffset: 0,
		recording:    false,
	}
}

// Read a rune if one is available, otherwise return an EOFError.
func (s *ScannerBacktrackingScanner) Read() (rune, error) {
	r, _, err := s.wrapped.ReadRune()
	if err != nil {
		return 0, err
	}
	if s.recording {
		s.buffer = append(s.buffer, r)
	}
	return r, nil
}

// StartBuffer starts recording runs returned from Read() into the
// buffer, if not already recording.
func (s *ScannerBacktrackingScanner) StartBuffer() {
	s.recording = true
}

// ReadBuffer reads a buffered rune out of a buffer. The buffer is
// empty unless StartBuffer() then one or more calls to Read()
// happens.
func (s *ScannerBacktrackingScanner) ReadBuffer(bufIdx int) rune {
	return s.buffer[bufIdx]
}

// DropBuffer stops recording and empties the buffer.
func (s *ScannerBacktrackingScanner) DropBuffer() {
	s.buffer = []rune{}
	s.recording = false
}

// BufSize returns the current size of the buffer.
func (s *ScannerBacktrackingScanner) BufSize() int {
	return len(s.buffer)
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

// Snapshot records the state of a snapshot taken by a scanner.
type Snapshot struct {
	buffOffset int
	position   textpos.TextPos
	next       *Snapshot
}

// RewindableScanner is an implementation of Scanner based on a
// BacktrackingScanner.
type RewindableScanner struct {
	source      BacktrackingScanner
	currentPos  textpos.TextPos
	lastSnap    *Snapshot
	isReplaying bool
	replayPos   int
}

// NewRewindableScanner creates a RewindableScanner from a
// BacktrackingScanner.
func NewRewindableScanner(source BacktrackingScanner) *RewindableScanner {
	return &RewindableScanner{
		source:      source,
		currentPos:  textpos.StartingPos(),
		lastSnap:    nil,
		isReplaying: false,
		replayPos:   0,
	}
}

// FromString creates a Scanner from a string.
func FromString(str string) Scanner {
	return NewRewindableScanner(RBSFromString(str))
}

// FromReader creates a Scanner from an io.RuneReader.
func FromReader(reader io.RuneReader) Scanner {
	return NewRewindableScanner(SBSFromReader(reader))
}

// Read a rune if one is available, otherwise return an EOFError.
func (s *RewindableScanner) Read() (rune, error) {
	if s.isReplaying {
		if s.replayPos < s.source.BufSize() {
			r := s.source.ReadBuffer(s.replayPos)
			s.replayPos++
			s.currentPos = s.currentPos.Advance(r)
			return r, nil
		}
		s.isReplaying = false
	}

	r, err := s.source.Read()
	if err != nil {
		return r, err
	}
	s.currentPos = s.currentPos.Advance(r)
	return r, nil
}

// GetPos returns the position of the next character Read() will
// return.
func (s *RewindableScanner) GetPos() textpos.TextPos {
	return s.currentPos
}

// StartSnapshot takes a new snapshot that can be rolled back to
// later.
func (s *RewindableScanner) StartSnapshot() {
	// Make sure the current data is recorded
	if s.lastSnap == nil {
		s.source.StartBuffer()
	}
	var offset int
	if s.isReplaying {
		offset = s.replayPos
	} else {
		offset = s.source.BufSize()
	}
	// Capture the position information necessary to know where to
	// start replaying from.
	s.lastSnap = &Snapshot{
		buffOffset: offset,
		position:   s.currentPos,
		next:       s.lastSnap,
	}
}

// RewindSnapshot reverts the scanner back to the state it was in when
// StartSnapshot() was last called.
func (s *RewindableScanner) RewindSnapshot() {
	if s.lastSnap == nil {
		panic("Bug: rewinding to a snapshot that was never started")
	}

	s.currentPos = s.lastSnap.position
	s.replayPos = s.lastSnap.buffOffset
	s.lastSnap = s.lastSnap.next
	s.isReplaying = true
}

// PopSnapshot drops a snapshot when it is no longer needed.
func (s *RewindableScanner) PopSnapshot() {
	if s.lastSnap == nil {
		panic("Bug: popped a snapshot that was never started")
	}
	s.lastSnap = s.lastSnap.next
	if s.lastSnap == nil {
		// Drop buffered data because nothing references it
		s.source.DropBuffer()
	}
}
