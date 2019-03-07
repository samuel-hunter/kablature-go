package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	NOTES         = "cdefgab"
	OCTAVE_LENGTH = len(NOTES)

	EIGHTH_NOTE  = 1
	QUARTER_NOTE = 2
	HALF_NOTE    = 4
	WHOLE_NOTE   = 8
)

var NOTE_LENGTHS = []byte{
	EIGHTH_NOTE,
	QUARTER_NOTE,
	HALF_NOTE,
	WHOLE_NOTE,
}

type Symbol interface {
	Length() byte // Number of eighth beats a note holds
	Dotted() bool // Whether the length is dotted (i.e. x1.5)
	Equal(Symbol) bool
}

type Note struct {
	length byte
	dotted bool
	pitch  byte // 0 = C, 7 = G, 8 = C at a higher octave, etc.
}

type Chord struct {
	length  byte
	dotted  bool
	pitches []byte
}

type Rest struct {
	length byte
	dotted bool
}

type Parser struct {
	scanner *Scanner
	tokbuf  *Token
	length  byte
	dotted  bool
	octave  int
}

func (n Note) Length() byte { return n.length }
func (n Note) Dotted() bool { return n.dotted }
func (n Note) String() string {
	return fmt.Sprintf("<NOTE %d %t %d>", n.length, n.dotted, n.pitch)
}
func (n Note) Equal(s Symbol) bool {
	that, ok := s.(Note)

	return ok &&
		n.length == that.length &&
		n.dotted == that.dotted &&
		n.pitch == that.pitch
}

func (c Chord) Length() byte { return c.length }
func (c Chord) Dotted() bool { return c.dotted }
func (c Chord) String() string {
	return fmt.Sprintf("<CHORD %d %t %s>", c.length, c.dotted, c.pitches)
}
func (c Chord) Equal(s Symbol) bool {
	// TODO check pitches equality
	that, ok := s.(Chord)

	return ok &&
		c.length == that.length &&
		c.dotted == that.dotted
}

func (r Rest) Length() byte { return r.length }
func (r Rest) Dotted() bool { return r.dotted }
func (r Rest) String() string {
	return fmt.Sprintf("<REST %d %t>", r.length, r.dotted)
}
func (r Rest) Equal(s Symbol) bool {
	that, ok := s.(Rest)

	return ok &&
		r.length == that.length &&
		r.dotted == that.dotted
}

func NewParser(scanner *Scanner) *Parser {
	return &Parser{
		scanner: scanner,
		tokbuf:  nil,
		length:  QUARTER_NOTE,
		dotted:  false,
	}
}

func (parser *Parser) peekToken() (*Token, error) {
	if parser.tokbuf == nil {
		tok, err := parser.scanner.Next()
		if err != nil {
			return nil, err
		}
		parser.tokbuf = tok
	}

	return parser.tokbuf, nil
}

func (parser *Parser) nextToken() (*Token, error) {
	if parser.tokbuf == nil {
		return parser.scanner.Next()
	}

	result := parser.tokbuf
	parser.tokbuf = nil
	return result, nil
}

func (p *Parser) scanPitch() (byte, error) {
	// Consume note token
	tok, err := p.nextToken()
	if err != nil {
		return 0, err
	}

	pitch := strings.Index(NOTES, strings.ToLower(tok.content))
	if pitch < 0 {
		return 0, errors.New(fmt.Sprintf("Note '%s' doesn't exist.", tok.content))
	}

	pitch += p.octave * OCTAVE_LENGTH

	// Find raise token if it exists.
	tok, err = p.peekToken()
	if err != nil && err != io.EOF {
		return 0, err
	}

	if err != io.EOF && tok.typ == TOK_NOTE_RAISE {
		// Consume the token.
		p.nextToken()

		pitch += OCTAVE_LENGTH
	}

	return byte(pitch), nil
}

func (p *Parser) scanNote() (Note, error) {
	pitch, err := p.scanPitch()
	if err != nil {
		return Note{}, err
	}

	return Note{length: p.length, dotted: p.dotted, pitch: pitch}, nil
}

func (p *Parser) scanChord() (Chord, error) {
	// Consume the open paren token
	tok, err := p.nextToken()
	if err != nil {
		return Chord{}, err
	}

	if tok.typ != TOK_PAREN_OPEN {
		return Chord{}, errors.New("Expected open parenthesis")
	}

	pitches := []byte{}

	for {
		tok, err = p.peekToken()
		switch tok.typ {
		case TOK_NOTE:
			pitch, err := p.scanPitch()
			if err != nil {
				return Chord{}, err
			}

			pitches = append(pitches, pitch)
		case TOK_PAREN_CLOSE:
			p.nextToken() // consume token
			return Chord{length: p.length, dotted: p.dotted, pitches: pitches}, nil
		default:
			return Chord{}, errors.New(fmt.Sprintf("Unexpected token '%s'.", tok))
		}
	}
}

func (parser *Parser) takeLength() error {
	tok, err := parser.nextToken()
	if err != nil {
		return err
	}

	l, err := strconv.ParseInt(tok.content, 10, 8)
	if err != nil {
		return err
	}
	length := byte(l)

	if bytes.IndexByte(NOTE_LENGTHS, length) < 0 {
		return errors.New("Invalid note length " + tok.content)
	}

	parser.length = length

	tok, err = parser.peekToken()
	if err != nil && err != io.EOF {
		return err
	}

	if err != io.EOF && tok.typ == TOK_DOT {
		parser.nextToken() // Consume token
		parser.dotted = true
	}

	return nil
}

func (parser *Parser) Next() (Symbol, error) {
	for {
		tok, err := parser.peekToken()
		if err != nil {
			return nil, err
		}

		switch tok.typ {
		case TOK_NOTE:
			return parser.scanNote()
		case TOK_DURATION:
			err := parser.takeLength()
			if err != nil {
				return nil, err
			}
		case TOK_PAREN_OPEN:
			return parser.scanChord()
		case TOK_OCTAVE_UPSHIFT:
			parser.octave++
			parser.nextToken() // Consume the token
		case TOK_OCTAVE_DOWNSHIFT:
			if parser.octave == 0 {
				return nil, errors.New("Unexpected '<': Can't downshift from octave 0")
			}
			parser.octave--
			parser.nextToken() // Consume the token
		default:
			return nil, errors.New("Unexpected token: " + tok.String())
		}
	}
}
