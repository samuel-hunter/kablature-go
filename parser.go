// parser - converting text source file to a more computer-friendly
// data structure.

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const (
	TOK_CHORD_START      = '('
	TOK_CHORD_END        = ')'
	TOK_OCTAVE_UPSHIFT   = '>'
	TOK_OCTAVE_DOWNSHIFT = '<'
	TOK_NOTE_RAISE       = '\''
	TOK_DOT              = '.'
	TOK_REST             = 'r'
)

type Parser struct {
	reader *bufio.Reader
	length byte
	dotted bool
	octave int
}

func isNote(r rune) bool {
	r = unicode.To(unicode.LowerCase, r)
	return strings.ContainsRune(NOTES, r)
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(reader),
		length: QUARTER_NOTE,
		dotted: false,
	}
}

func (p *Parser) next() (rune, error) {
	for {
		r, _, err := p.reader.ReadRune()
		if err != nil {
			return -1, err
		}
		if !unicode.IsSpace(r) {
			return r, nil
		}
	}
}

// Skip the rest of the line.
func (p *Parser) skipLine() error {
	for {
		r, _, err := p.reader.ReadRune()
		if err != nil {
			return err
		}
		if r == '\n' {
			return nil
		}
	}
}

func (p *Parser) unread() {
	p.reader.UnreadRune()
}

func (p *Parser) scanPitch() (byte, error) {
	// Consume note token
	r, err := p.next()
	if err != nil {
		return 0, err
	}

	pitch := strings.Index(NOTES, string(unicode.To(unicode.LowerCase, r)))
	if pitch < 0 {
		return 0, errors.New(fmt.Sprintf("Note '%c' doesn't exist.", r))
	}

	pitch += p.octave * OCTAVE_LENGTH

	// Find raise token if it exists.
	r, err = p.next()
	if err != nil && err != io.EOF {
		return 0, err
	}

	if err != io.EOF && r == TOK_NOTE_RAISE {
		pitch += OCTAVE_LENGTH
	} else {
		// Not the token we're looking for; unread.
		p.unread()
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
	var pitches []byte

	for {
		r, err := p.next()
		if err != nil {
			return Chord{}, err
		}

		if r == TOK_CHORD_END {
			return Chord{length: p.length, dotted: p.dotted, pitches: pitches}, nil
		} else if isNote(r) {
			// Unread rune for pitch reader to consume it
			p.unread()

			pitch, err := p.scanPitch()
			if err != nil {
				return Chord{}, err
			}

			pitches = append(pitches, pitch)
		} else {
			return Chord{}, errors.New(fmt.Sprintf("Unexpected character '%c'.", r))
		}
	}
}

func (parser *Parser) takeLength() error {
	r, err := parser.next()
	if err != nil {
		return err
	}

	l, err := strconv.ParseInt(string(r), 10, 8)
	if err != nil {
		return err
	}
	length := byte(l)

	if bytes.IndexByte(NOTE_LENGTHS, length) < 0 {
		return errors.New("Invalid note length " + string(r))
	}

	parser.length = length

	r, err = parser.next()
	if err != nil && err != io.EOF {
		return err
	}

	if err != io.EOF && r == TOK_DOT {
		parser.dotted = true
	} else {
		parser.unread()
		parser.dotted = false
	}

	return nil
}

func (parser *Parser) Next() (Symbol, error) {
	for {
		r, err := parser.next()
		if err != nil {
			return nil, err
		}

		if isNote(r) {
			parser.unread()
			return parser.scanNote()
		} else if unicode.IsNumber(r) {
			parser.unread()
			err := parser.takeLength()
			if err != nil {
				return nil, err
			}
		} else {
			switch r {
			case '#':
				// Skip the rest of the line.
				err = parser.skipLine()
				if err != nil {
					return nil, err
				}
			case TOK_CHORD_START:
				return parser.scanChord()
			case TOK_OCTAVE_UPSHIFT:
				parser.octave++
			case TOK_OCTAVE_DOWNSHIFT:
				if parser.octave == 0 {
					return nil, errors.New(fmt.Sprintf(
						"Unexpected '%c'; Can't downshift from octave 0", r))
				}
				parser.octave--
			case TOK_REST:
				return Rest{length: parser.length, dotted: parser.dotted}, nil
			default:
				return nil, errors.New(fmt.Sprintf("Unexpected character '%c'", r))
			}
		}
	}
}
