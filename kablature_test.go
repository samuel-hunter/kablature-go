package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func reportLex(s string, expected []Token) {
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

		if tok != expected[i] {
			panic(fmt.Sprintf("Real doesn't match expected token %s.", expected[i].String()))
		}
		i++
	}

	if i != len(expected) {
		panic("Less tokens than expected.")
	}
}

func TestLexer(t *testing.T) {
	reportLex("a b c", []Token{
		Token{typ: TOK_NOTE, content: "a"},
		Token{typ: TOK_NOTE, content: "b"},
		Token{typ: TOK_NOTE, content: "c"},
	})
	reportLex("(e' e) (e g b)", []Token{
		Token{typ: TOK_PAREN_OPEN, content: "("},
		Token{typ: TOK_NOTE, content: "e'"},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_PAREN_CLOSE, content: ")"},
		Token{typ: TOK_PAREN_OPEN, content: "("},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE, content: "g"},
		Token{typ: TOK_NOTE, content: "b"},
		Token{typ: TOK_PAREN_CLOSE, content: ")"},
	})
	reportLex("1/2 e' c  2e", []Token{
		Token{typ: TOK_DURATION, content: "1/2"},
		Token{typ: TOK_NOTE, content: "e'"},
		Token{typ: TOK_NOTE, content: "c"},
		Token{typ: TOK_DURATION, content: "2"},
		Token{typ: TOK_NOTE, content: "e"},
	})
	reportLex("e g > e g < c", []Token{
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE, content: "g"},
		Token{typ: TOK_OCTAVE_UPSHIFT, content: ">"},
		Token{typ: TOK_NOTE, content: "e"},
		Token{typ: TOK_NOTE, content: "g"},
		Token{typ: TOK_OCTAVE_DOWNSHIFT, content: "<"},
		Token{typ: TOK_NOTE, content: "c"},
	})
}
