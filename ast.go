package main

const (
	EIGHTH_NOTE  = 1
	QUARTER_NOTE = 2
	HALF_NOTE    = 4
	WHOLE_NOTE   = 8
)

type Symbol interface {
	Length() int  // Number of eighth beats a note holds
	Dotted() bool // Whether the length is dotted (i.e. x1.5)
}

type Note struct {
	length int
	dotted bool
	pitch  byte // 0 = C, 7 = G, 8 = C at a higher octave, etc.
}

type Chord struct {
	length  int
	dotted  bool
	pitches []byte
}

type Rest struct {
	length int
	dotted bool
}

func (n Note) Length() int  { return n.length }
func (n Note) Dotted() bool { return n.dotted }

func (c Chord) Length() int  { return c.length }
func (c Chord) Dotted() bool { return c.dotted }

func (r Rest) Length() int  { return r.length }
func (r Rest) Dotted() bool { return r.dotted }
