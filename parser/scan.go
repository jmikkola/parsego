package parser

import (
	"io"
)

type EOFError struct{}

func (_ *EOFError) Error() string {
	return "Reached end of input"
}

type ReadRune interface {
	Read() (rune, error)
}

type BacktrackingScanner interface {
	ReadRune
	StartBuffer()
	ReadBuffer(int) rune
	BufSize() int
	DropBuffer()
}

type RuneBacktrackingScanner struct {
	runes        []rune
	pos          int
	recordOffset int
	recording    bool
}

func RBSFromString(s string) *RuneBacktrackingScanner {
	return &RuneBacktrackingScanner{
		runes:        []rune(s),
		pos:          0,
		recordOffset: 0,
		recording:    false,
	}
}

func (s *RuneBacktrackingScanner) Read() (rune, error) {
	if s.pos >= len(s.runes) {
		return 0, &EOFError{}
	}
	r := s.runes[s.pos]
	s.pos++
	return r, nil
}

func (s *RuneBacktrackingScanner) StartBuffer() {
	// Not much to do, the entire string is saved anyway.
	// Just save the offset for math later.
	if !s.recording {
		s.recordOffset = s.pos
		s.recording = true
	}
}

func (s *RuneBacktrackingScanner) ReadBuffer(bufIdx int) rune {
	return s.runes[s.recordOffset+bufIdx]
}

func (s *RuneBacktrackingScanner) DropBuffer() {
	s.recordOffset = 0
	s.recording = false
}

func (s *RuneBacktrackingScanner) BufSize() int {
	if s.recording {
		return s.pos - s.recordOffset
	} else {
		return 0
	}
}

type ScannerBacktrackingScanner struct {
	wrapped      io.RuneReader
	buffer       []rune
	bufferOffset int
	recording    bool
}

func SBSFromReader(reader io.RuneReader) *ScannerBacktrackingScanner {
	return &ScannerBacktrackingScanner{
		wrapped:      reader,
		buffer:       []rune{},
		bufferOffset: 0,
		recording:    false,
	}
}

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

func (s *ScannerBacktrackingScanner) StartBuffer() {
	s.recording = true
}

func (s *ScannerBacktrackingScanner) ReadBuffer(bufIdx int) rune {
	return s.buffer[bufIdx]
}

func (s *ScannerBacktrackingScanner) DropBuffer() {
	s.buffer = []rune{}
	s.recording = false
}

func (s *ScannerBacktrackingScanner) BufSize() int {
	return len(s.buffer)
}

type Scanner interface {
	ReadRune
	GetPos() TextPos
	StartSnapshot()
	RewindSnapshot()
	PopSnapshot()
}

type Snapshot struct {
	buffOffset int
	position   TextPos
	next       *Snapshot
}

type RewindableScanner struct {
	source      BacktrackingScanner
	currentPos  TextPos
	lastSnap    *Snapshot
	isReplaying bool
	replayPos   int
}

func NewRewindableScanner(source BacktrackingScanner) *RewindableScanner {
	return &RewindableScanner{
		source:      source,
		currentPos:  StartingPos(),
		lastSnap:    nil,
		isReplaying: false,
		replayPos:   0,
	}
}

func FromString(str string) Scanner {
	return NewRewindableScanner(RBSFromString(str))
}

func FromReader(reader io.RuneReader) Scanner {
	return NewRewindableScanner(SBSFromReader(reader))
}

func (s *RewindableScanner) Read() (rune, error) {
	if s.isReplaying {
		if s.replayPos < s.source.BufSize() {
			r := s.source.ReadBuffer(s.replayPos)
			s.replayPos++
			s.currentPos = s.currentPos.Advance(r)
			return r, nil
		} else {
			s.isReplaying = false
		}
	}

	r, err := s.source.Read()
	if err != nil {
		return r, err
	}
	s.currentPos = s.currentPos.Advance(r)
	return r, nil
}

func (s *RewindableScanner) GetPos() TextPos {
	return s.currentPos
}

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

func (s *RewindableScanner) RewindSnapshot() {
	if s.lastSnap == nil {
		panic("Bug: rewinding to a snapshot that was never started")
	}

	s.currentPos = s.lastSnap.position
	s.replayPos = s.lastSnap.buffOffset
	s.lastSnap = s.lastSnap.next
	s.isReplaying = true
}

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
