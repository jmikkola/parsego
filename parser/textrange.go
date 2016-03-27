package parser

type TextPos struct {
	line int
	col  int
}

type TextRange struct {
	start TextPos
	end   TextPos
}

func StartingPos() TextPos {
	return TextPos{
		line: 0,
		col:  0,
	}
}

func Pos(line, col int) TextPos {
	return TextPos{line, col}
}

func (pos TextPos) AdvanceCol() TextPos {
	return TextPos{
		col:  pos.col + 1,
		line: pos.line,
	}
}

func (pos TextPos) AdvanceLine() TextPos {
	return TextPos{
		col:  0,
		line: pos.line + 1,
	}
}

func (pos TextPos) Advance(c rune) TextPos {
	if c == '\n' {
		return pos.AdvanceLine()
	} else {
		return pos.AdvanceCol()
	}
}
