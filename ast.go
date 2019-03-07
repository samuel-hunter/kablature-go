package main

import "fmt"

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
