package scanner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmikkola/parsego/parser/scanner"
	"github.com/jmikkola/parsego/parser/textpos"
)

func assertReads(t *testing.T, sc scanner.ReadRune, c rune) {
	r, err := sc.Read()
	assert.NoError(t, err, "Expected successful read")
	assert.Equal(t, c, r, "Expected char")
}

func TestRBS(t *testing.T) {
	sc := scanner.RBSFromString("some text")
	assertReads(t, sc, 's')
	assertReads(t, sc, 'o')

	sc.StartBuffer()
	assert.Equal(t, 0, sc.BufSize())
	assertReads(t, sc, 'm')
	assertReads(t, sc, 'e')

	assert.Equal(t, 2, sc.BufSize())
	assert.Equal(t, 'm', sc.ReadBuffer(0))
	assert.Equal(t, 'e', sc.ReadBuffer(1))

	sc.DropBuffer()
	assert.Equal(t, 0, sc.BufSize())

	assertReads(t, sc, ' ')
	assertReads(t, sc, 't')
}

func TestSimpleRewind(t *testing.T) {
	sc := scanner.FromString("abcdefgh")

	assert.Equal(t, textpos.Pos(0, 0), sc.GetPos())
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	assert.Equal(t, textpos.Pos(0, 2), sc.GetPos())

	sc.StartSnapshot()
	assertReads(t, sc, 'c')
	assertReads(t, sc, 'd')
	assert.Equal(t, textpos.Pos(0, 4), sc.GetPos())

	sc.RewindSnapshot()
	assert.Equal(t, textpos.Pos(0, 2), sc.GetPos())
	assertReads(t, sc, 'c')
	assertReads(t, sc, 'd')
	assert.Equal(t, textpos.Pos(0, 4), sc.GetPos())
}

func TestSimplePop(t *testing.T) {
	sc := scanner.FromString("abcdefgh")

	assert.Equal(t, textpos.Pos(0, 0), sc.GetPos())
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	assert.Equal(t, textpos.Pos(0, 2), sc.GetPos())

	sc.StartSnapshot()
	assertReads(t, sc, 'c')
	assertReads(t, sc, 'd')
	assert.Equal(t, textpos.Pos(0, 4), sc.GetPos())

	sc.PopSnapshot()
	assert.Equal(t, textpos.Pos(0, 4), sc.GetPos())
	assertReads(t, sc, 'e')
	assertReads(t, sc, 'f')
	assert.Equal(t, textpos.Pos(0, 6), sc.GetPos())
}

func TestRecursiveSnapshots(t *testing.T) {
	sc := scanner.FromString("abcdefgh")

	assert.Equal(t, textpos.Pos(0, 0), sc.GetPos())
	assertReads(t, sc, 'a')
	assert.Equal(t, textpos.Pos(0, 1), sc.GetPos())

	sc.StartSnapshot()
	assertReads(t, sc, 'b')
	assert.Equal(t, textpos.Pos(0, 2), sc.GetPos())

	sc.StartSnapshot()
	assertReads(t, sc, 'c')
	assert.Equal(t, textpos.Pos(0, 3), sc.GetPos())

	sc.StartSnapshot()
	assertReads(t, sc, 'd')
	assert.Equal(t, textpos.Pos(0, 4), sc.GetPos())

	sc.RewindSnapshot()
	assertReads(t, sc, 'd')
	assert.Equal(t, textpos.Pos(0, 4), sc.GetPos())
	assertReads(t, sc, 'e')
	assert.Equal(t, textpos.Pos(0, 5), sc.GetPos())

	sc.RewindSnapshot()
	sc.RewindSnapshot()
	assert.Equal(t, textpos.Pos(0, 1), sc.GetPos())
	assertReads(t, sc, 'b')
}

func TestRepeatedRetry(t *testing.T) {
	sc := scanner.FromString("abcdefgh")
	sc.StartSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	sc.RewindSnapshot()

	sc.StartSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	sc.RewindSnapshot()

	sc.StartSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	sc.PopSnapshot()

	assertReads(t, sc, 'c')
	assertReads(t, sc, 'd')
}

func TestTwoSnapshotsInTheSamePlace(t *testing.T) {
	sc := scanner.FromString("abcdefgh")
	sc.StartSnapshot()
	sc.StartSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	sc.RewindSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	sc.RewindSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
}

func TestStartingSnapshotWhileReplaying(t *testing.T) {
	sc := scanner.FromString("abcdefgh")
	sc.StartSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	assertReads(t, sc, 'c')
	assertReads(t, sc, 'd')
	assertReads(t, sc, 'e')
	sc.RewindSnapshot()
	assertReads(t, sc, 'a')
	assertReads(t, sc, 'b')
	assertReads(t, sc, 'c')
	sc.StartSnapshot()
	assertReads(t, sc, 'd')
	assertReads(t, sc, 'e')
	assertReads(t, sc, 'f')
	sc.RewindSnapshot()
	assertReads(t, sc, 'd')
	assertReads(t, sc, 'e')
	assertReads(t, sc, 'f')
}
