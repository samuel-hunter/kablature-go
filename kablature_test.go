package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func testLex(s string, expected []Token) {
	scanner := NewScanner(strings.NewReader(s))
	i := 0

	fmt.Printf("\nTokens from \"%s\":\n", s)
	for {
		tok, err := scanner.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		fmt.Println(tok)
		if i == len(expected) {
			panic("More tokens than expected.")
		}

		if *tok != expected[i] {
			panic(fmt.Sprintf("Real doesn't match expected token %s.", expected[i]))
		}
		i++
	}

	if i != len(expected) {
		panic("Less tokens than expected.")
	}
}

func testParse(s string, expected []Symbol) {
	parser := NewParser(NewScanner(strings.NewReader(s)))
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

func TestLexerLexesNotes(t *testing.T) {
	testLex("a b c", []Token{
		Token{typ: TOK_NOTE, content: "a"},
		Token{typ: TOK_NOTE, content: "b"},
		Token{typ: TOK_NOTE, content: "c"},
	})
}

func TestLexerLexesChords(t *testing.T) {
	testLex("(e' e) (e g b)", []Token{
		Token{typ: TOK_PAREN_OPEN, content: "("},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE_RAISE, content: "'"},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_PAREN_CLOSE, content: ")"},
		Token{typ: TOK_PAREN_OPEN, content: "("},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE, content: "g"},
		Token{typ: TOK_NOTE, content: "b"},
		Token{typ: TOK_PAREN_CLOSE, content: ")"},
	})
}

func TestLexerLexesDurations(t *testing.T) {
	testLex("8 e' c  2e", []Token{
		Token{typ: TOK_DURATION, content: "8"},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE_RAISE, content: "'"},
		Token{typ: TOK_NOTE, content: "c"},
		Token{typ: TOK_DURATION, content: "2"},
		Token{typ: TOK_NOTE, content: "e"},
	})
}

func TestLexerLexesOctaveShifts(t *testing.T) {
	testLex("e g > c e < e", []Token{
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE, content: "g"},
		Token{typ: TOK_OCTAVE_UPSHIFT, content: ">"},
		Token{typ: TOK_NOTE, content: "c"},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_OCTAVE_DOWNSHIFT, content: "<"},
		Token{typ: TOK_NOTE, content: "e"},
	})
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
