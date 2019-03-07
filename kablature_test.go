package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func testParse(s string, expected []Symbol) {
	parser := NewParser(strings.NewReader(s))
	i := 0

	fmt.Printf("\nSymbols from \"%s\":\n", s)
	for {
		sym, err := parser.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		fmt.Println(sym)
		if i == len(expected) {
			panic("More symbols than expected.")
		}

		if !sym.Equal(expected[i]) {
			panic(fmt.Sprintf("Real doesn't match expected symbol %s.", expected[i]))
		}
		i++
	}

	if i != len(expected) {
		panic("Less symbols than expected.")
	}
}

func TestParserParsesNotes(t *testing.T) {
	testParse("c d e", []Symbol{
		Note{length: QUARTER_NOTE, dotted: false, pitch: 0},
		Note{length: QUARTER_NOTE, dotted: false, pitch: 1},
		Note{length: QUARTER_NOTE, dotted: false, pitch: 2},
	})
}

func TestParserParsesDurations(t *testing.T) {
	testParse("c 4 d 8. e", []Symbol{
		Note{length: QUARTER_NOTE, dotted: false, pitch: 0},
		Note{length: HALF_NOTE, dotted: false, pitch: 1},
		Note{length: WHOLE_NOTE, dotted: true, pitch: 2},
	})
}

func TestParserParsesChords(t *testing.T) {
	testParse("(c d e) (f g a)", []Symbol{
		Chord{length: 2, dotted: false, pitches: []byte{0, 1, 2}},
		Chord{length: 2, dotted: false, pitches: []byte{3, 4, 5}},
	})
}

func TestParserParsesOctaves(t *testing.T) {
	testParse("c > c c' < c'", []Symbol{
		Note{length: 2, dotted: false, pitch: 0},
		Note{length: 2, dotted: false, pitch: 7},
		Note{length: 2, dotted: false, pitch: 14},
		Note{length: 2, dotted: false, pitch: 7},
	})
}

func TestParserParsesMixed(t *testing.T) {
	testParse("4 e 1 c 2 (c e g)", []Symbol{
		Note{length: 4, dotted: false, pitch: 2},
		Note{length: 1, dotted: false, pitch: 0},
		Chord{length: 2, dotted: false, pitches: []byte{0, 2, 4}},
	})
}
