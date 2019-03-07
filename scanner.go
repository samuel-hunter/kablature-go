package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	TOK_NOTE TokenType = iota
	TOK_DURATION
	TOK_PAREN_OPEN
	TOK_PAREN_CLOSE
	TOK_OCTAVE_UPSHIFT
	TOK_OCTAVE_DOWNSHIFT
	TOK_NOTE_RAISE
	TOK_DOT
)

type TokenType int

type Token struct {
	typ     TokenType
	content string
}

type Scanner struct {
	br *bufio.Reader
}

func (typ TokenType) String() string {
	switch typ {
	case TOK_NOTE:
		return "NOTE"
	case TOK_DURATION:
		return "DURATION"
	case TOK_PAREN_OPEN:
		return "PAREN_OPEN"
	case TOK_PAREN_CLOSE:
		return "PAREN_CLOSE"
	case TOK_OCTAVE_UPSHIFT:
		return "OCTAVE_UPSHIFT"
	case TOK_OCTAVE_DOWNSHIFT:
		return "OCTAVE_DOWNSHIFT"
	case TOK_NOTE_RAISE:
		return "NOTE_RAISE"
	case TOK_DOT:
		return "DOT"
	default:
		panic("unrecognized token type")
	}
}

func (tok Token) String() string {
	return fmt.Sprintf("<%s \"%s\">", tok.typ, tok.content)
}

// Return true if the rune is a recognized single-character token
func isSymbol(r rune) bool {
	switch r {
	case '(', ')', '<', '>', '\'', '.':
		return true
	}

	return false
}

// Return true if the rune is a note duration character
func isNoteDuration(r rune) bool {
	return unicode.IsNumber(r)
}

// Return true if the rune is a note character.6
func isNote(r rune) bool {
	r = unicode.To(unicode.LowerCase, r)
	return strings.ContainsRune(NOTES, r)
}

func scan(br *bufio.Reader, checker func(rune) bool, typ TokenType) (*Token, error) {
	content := ""

	for {
		r, _, err := br.ReadRune()
		if err == io.EOF {
			return &Token{typ: typ, content: content}, nil
		} else if err != nil {
			return nil, err
		}

		if checker(r) {
			content += string(r)
		} else {
			br.UnreadRune()
			return &Token{typ: typ, content: content}, nil
		}
	}
}

func (scanner Scanner) Next() (*Token, error) {
	br := scanner.br

	// Skip all the leading whitespace
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			return nil, err
		}

		if !unicode.IsSpace(r) {
			br.UnreadRune()
			break
		}
	}

	r, _, err := br.ReadRune()
	if err != nil {
		return nil, err
	}

	switch r {
	case '(':
		return &Token{typ: TOK_PAREN_OPEN, content: "("}, nil
	case ')':
		return &Token{typ: TOK_PAREN_CLOSE, content: ")"}, nil
	case '>':
		return &Token{typ: TOK_OCTAVE_UPSHIFT, content: ">"}, nil
	case '<':
		return &Token{typ: TOK_OCTAVE_DOWNSHIFT, content: "<"}, nil
	case '\'':
		return &Token{typ: TOK_NOTE_RAISE, content: "'"}, nil
	case '.':
		return &Token{typ: TOK_DOT, content: "."}, nil
	}

	if isNote(r) {
		return &Token{typ: TOK_NOTE, content: string(r)}, nil
	} else if isNoteDuration(r) {
		return &Token{typ: TOK_DURATION, content: string(r)}, nil
	}

	return nil, errors.New(fmt.Sprintf("Unexpected character '%c'.", r))
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{br: bufio.NewReader(r)}
}
