package main

import (
	"bytes"
	"errors"
	"github.com/ajstarks/svgo"
	"io"
	"math"
	"strconv"
)

const (
	TAB_WIDTH    = TABNOTE_WIDTH * NUM_TABNOTES
	TAB_MARGIN_X = 50
	TAB_MARGIN_Y = 10
	TAB_NOTES    = "BGECAFDCEGBDFAC"

	NUM_TABNOTES     = 15
	TABNOTE_OFFSET_Y = 5
	TABNOTE_WIDTH    = 15 // Also used for spacing between eighth notes.
	TABNOTE_COLOR    = "white"
	TABNOTE_MARKED   = "#ee7c80"

	MEASURE_THICKNESS = 3
	FONTSIZE          = 10

	NOTE_RADIUS   = 4
	NUM_NOTES     = len(TAB_NOTES)
	HALF_NOTES    = NUM_NOTES / 2
	SYMBOL_HEIGHT = TABNOTE_WIDTH // Spacing a general musical symbol would have allocated.

	// Calculated constatns
	TAB_TOP   = TAB_MARGIN_Y
	TAB_LEFT  = TAB_MARGIN_X
	TAB_RIGHT = TAB_LEFT + TAB_WIDTH
)

var (
	THIN_STYLE    = "stroke-width:1;stroke:black"
	MEASURE_STYLE = "stroke-width:" + strconv.Itoa(MEASURE_THICKNESS) + ";stroke:black"
	TEXT_STYLE    = "font-size:" + strconv.Itoa(FONTSIZE) + ";fill:black"
)

type Tablature struct {
	canvas    *svg.SVG
	current_y int
	measure   int
}

func drawTabNote(canvas *svg.SVG, tab_height, note, offset_y int, marked bool) {
	x := TAB_MARGIN_X + note*TABNOTE_WIDTH
	rect_height := tab_height + offset_y*TABNOTE_OFFSET_Y
	rect_style := THIN_STYLE
	text_style := TEXT_STYLE + ";text-anchor:middle"

	if marked {
		rect_style += ";fill:" + TABNOTE_MARKED
	} else {
		rect_style += ";fill:" + TABNOTE_COLOR
	}

	canvas.Rect(x, TAB_MARGIN_Y, TABNOTE_WIDTH, rect_height, rect_style)
	canvas.Text(x+TABNOTE_WIDTH/2, TAB_MARGIN_Y+rect_height+FONTSIZE,
		string(TAB_NOTES[note]), text_style)
}

func NewTablature(w io.Writer, tab_height int) *Tablature {
	width := TAB_WIDTH + TAB_MARGIN_X*2
	height := tab_height + TAB_MARGIN_Y*2 + HALF_NOTES*TABNOTE_OFFSET_Y + FONTSIZE

	canvas := svg.New(w)
	canvas.Start(width, height)

	// Frame a rectangle around the canvas to show border
	canvas.Rect(0, 0, width, height, "fill:none;stroke-width:1;stroke:green")

	for i := 0; i < NUM_NOTES; i++ {
		if i < HALF_NOTES {
			// Draw notes going down
			drawTabNote(canvas, tab_height, i, i, (i+2)%3 == 0)
		} else {
			// Draw notes going up
			drawTabNote(canvas, tab_height, i, NUM_NOTES-i-1, (i+2)%3 == 0)
		}
	}

	center_x := TAB_MARGIN_X + HALF_NOTES*TABNOTE_WIDTH + MEASURE_THICKNESS/2
	line_height := tab_height + HALF_NOTES*TABNOTE_OFFSET_Y
	canvas.Line(center_x, TAB_MARGIN_Y, center_x, TAB_MARGIN_Y+line_height,
		MEASURE_STYLE)

	return &Tablature{
		canvas:    canvas,
		current_y: TAB_MARGIN_Y + tab_height,
	}
}

func findTabHeight(symbols []Symbol) int {
	eighth_beats := 0

	for _, symb := range symbols {
		eighth_beats += symb.Length()
	}

	measure_bars := eighth_beats / 8
	if eighth_beats%8 == 0 {
		measure_bars--
	}

	// +1 for measure end
	return (eighth_beats + measure_bars + 1) * SYMBOL_HEIGHT
}

func (tab *Tablature) DrawMeasureBar() {
	bar_y := tab.current_y - MEASURE_THICKNESS/2
	text_style := TEXT_STYLE + ";dominant-baseline:central"
	text_margin_left := 2
	tab.measure++

	tab.canvas.Line(TAB_LEFT, bar_y, TAB_RIGHT, bar_y, MEASURE_STYLE)
	tab.canvas.Text(TAB_RIGHT+text_margin_left, bar_y,
		strconv.Itoa(tab.measure), text_style)

	tab.current_y -= SYMBOL_HEIGHT
}

// Return the x position that the provided pitch would be on.
func findNotePosition(pitch byte) (int, error) {
	notes := []byte{15, 13, 11, 9, 7, 5, 3, 1, 0, 2, 4, 6, 8, 10, 12, 14}
	index := bytes.IndexByte(notes, pitch)
	if index < 0 {
		return -1, errors.New("Pitch out of range.")
	}

	return TAB_LEFT + int(math.Ceil((float64(index)-0.5)*TABNOTE_WIDTH)), nil
}

// Draw a note without any stem or taper and return its x position.
func (tab *Tablature) DrawPitch(note Note) (int, error) {
	note_x, err := findNotePosition(note.pitch)

	if err != nil {
		return -1, err
	}

	filled := note.length == HALF_NOTE || note.length == WHOLE_NOTE

	circle_style := THIN_STYLE
	if filled {
		circle_style += ";fill:white"
	} else {
		circle_style += ";fill:black"
	}

	tab.canvas.Circle(note_x, tab.current_y, NOTE_RADIUS, circle_style)

	// Add dotted circle if required
	if note.dotted {
		tab.canvas.Circle(note_x+NOTE_RADIUS+2, tab.current_y-NOTE_RADIUS-3,
			2)
	}
	return note_x, nil
}

// Draw a note's stem and taper if appropriate.
func (tab *Tablature) DrawStem(note_x int, length int) {
	with_stem := length != WHOLE_NOTE
	tapered := length == EIGHTH_NOTE

	if with_stem {
		line_y := tab.current_y - NOTE_RADIUS
		tab.canvas.Line(TAB_LEFT-20, line_y, note_x, line_y, THIN_STYLE)

		if tapered {
			tab.canvas.Line(TAB_LEFT-20, line_y, TAB_LEFT-15, line_y-5, THIN_STYLE)
		}
	}
}

// Draw a note with a stem and taper when appropriate.
func (tab *Tablature) DrawNote(note Note) error {
	note_x, err := tab.DrawPitch(note)
	if err != nil {
		return err
	}

	tab.DrawStem(note_x, note.length)

	tab.current_y -= SYMBOL_HEIGHT * note.length
	return nil
}

func (tab *Tablature) DrawChord(chord Chord) error {
	rightmost_x := 0

	for _, pitch := range chord.pitches {
		note_x, err := tab.DrawPitch(Note{
			length: chord.length,
			dotted: chord.dotted,
			pitch:  pitch,
		})
		if err != nil {
			return err
		}

		if rightmost_x < note_x {
			rightmost_x = note_x
		}
	}

	tab.DrawStem(rightmost_x, chord.length)

	tab.current_y -= SYMBOL_HEIGHT * chord.length
	return nil
}

func DrawTablature(w io.Writer, symbols []Symbol) error {
	tablature := NewTablature(w, findTabHeight(symbols))
	defer tablature.canvas.End()

	eighth_beats := 0

	for _, symb := range symbols {
		// Add a measure bar when necessary.
		if eighth_beats%8 == 0 {
			tablature.DrawMeasureBar()
			eighth_beats = 0
		}

		// Add empty space for a note.
		switch symb.(type) {
		case Note:
			err := tablature.DrawNote(symb.(Note))
			if err != nil {
				return err
			}
		case Chord:
			err := tablature.DrawChord(symb.(Chord))
			if err != nil {
				return err
			}
		default:
			tablature.canvas.Circle(TAB_LEFT+TAB_WIDTH/2, tablature.current_y, NOTE_RADIUS, "")
			tablature.current_y -= symb.Length() * SYMBOL_HEIGHT
		}

		eighth_beats += symb.Length()
		if symb.Dotted() {
			eighth_beats += symb.Length() / 2
		}

		if eighth_beats > 8 {
			return errors.New("Uneven beats in measure (expected 8 eighth beats, received " +
				strconv.Itoa(eighth_beats) + ").")
		}
	}

	// Add ending lines
	end_y := TAB_TOP + MEASURE_THICKNESS/2
	tablature.canvas.Line(TAB_LEFT, end_y, TAB_RIGHT, end_y, MEASURE_STYLE)
	end_y += SYMBOL_HEIGHT / 2
	tablature.canvas.Line(TAB_LEFT, end_y, TAB_RIGHT, end_y, THIN_STYLE)

	return nil
}
