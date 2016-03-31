package parser

// TextPos is a single character position in some text. Both line and
// col start from 0.
type TextPos struct {
	line int
	col  int
}

// TextRange is an (inclusive) range between two TextPos.
type TextRange struct {
	start TextPos
	end   TextPos
}

// StartingPos returns the 0 position.
func StartingPos() TextPos {
	return TextPos{
		line: 0,
		col:  0,
	}
}

// Pos is a shorthand for creating a TextPos.
func Pos(line, col int) TextPos {
	return TextPos{line, col}
}

// AdvanceCol return a new TextPos with the column advanced by one.
func (pos TextPos) AdvanceCol() TextPos {
	return TextPos{
		col:  pos.col + 1,
		line: pos.line,
	}
}

// AdvanceLine returns a new TextPos with the line advanced by one.
func (pos TextPos) AdvanceLine() TextPos {
	return TextPos{
		col:  0,
		line: pos.line + 1,
	}
}

// Advance returns a new TextPos advanced by the given character.
func (pos TextPos) Advance(c rune) TextPos {
	if c == '\n' {
		return pos.AdvanceLine()
	}
	return pos.AdvanceCol()
}
