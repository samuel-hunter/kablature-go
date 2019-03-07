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
	TAB_WIDTH    = TABNOTE_WIDTH * NUM_NOTES
	TAB_MARGIN_X = 50
	TAB_MARGIN_Y = 10
	TAB_NOTES    = "DBGECAFDCEGBDFACE"

	TABNOTE_OFFSET_Y = 5
	TABNOTE_WIDTH    = 15
	TABNOTE_COLOR    = "white"
	TABNOTE_MARKED   = "#ee7c80"

	MEASURE_THICKNESS = 3
	FONT_SIZE         = 10

	NOTE_RADIUS   = 4
	SYMBOL_HEIGHT = TABNOTE_WIDTH // Spacing a general musical symbol would have allocated.

	// Calculated constants
	MEASURES_PER_TAB = 7
	MAX_TAB_HEIGHT   = (9*MEASURES_PER_TAB + 1) * SYMBOL_HEIGHT
	NUM_NOTES        = len(TAB_NOTES)
	HALF_NOTES       = NUM_NOTES / 2
)

var (
	THIN_STYLE    = "stroke-width:1;stroke:black"
	MEASURE_STYLE = "stroke-width:" + strconv.Itoa(MEASURE_THICKNESS) + ";stroke:black"
	TEXT_STYLE    = "font-size:" + strconv.Itoa(FONT_SIZE) + ";fill:black"
)

type TabScore struct {
	canvas      *svg.SVG
	cur_tab     int
	max_tabs    int
	tab_started bool
	current_y   int
	measure     int
}

func NewScore(w io.Writer, total_measures int) *TabScore {
	tablatures := int(math.Ceil(float64(total_measures) / MEASURES_PER_TAB))

	width := tablatures*TAB_WIDTH + (tablatures+1)*TAB_MARGIN_X
	height := MAX_TAB_HEIGHT + TAB_MARGIN_Y*2 + HALF_NOTES*TABNOTE_OFFSET_Y + FONT_SIZE

	canvas := svg.New(w)
	canvas.Start(width, height)

	// Fill canvas with white background
	canvas.Rect(0, 0, width, height, "fill:white")

	return &TabScore{canvas: canvas, max_tabs: tablatures}
}

func (tab *TabScore) Close() {
	if tab.tab_started {
		tab.canvas.Gend()
	}

	tab.canvas.End()
}

func (score *TabScore) drawTabNote(tab_height, note, offset_y int, marked bool) {
	canvas := score.canvas
	x := note * TABNOTE_WIDTH
	rect_height := tab_height + offset_y*TABNOTE_OFFSET_Y
	rect_style := THIN_STYLE
	text_style := TEXT_STYLE + ";text-anchor:middle"

	if marked {
		rect_style += ";fill:" + TABNOTE_MARKED
	} else {
		rect_style += ";fill:" + TABNOTE_COLOR
	}

	canvas.Rect(x, 0, TABNOTE_WIDTH, rect_height, rect_style)
	canvas.Text(x+TABNOTE_WIDTH/2, rect_height+FONT_SIZE,
		string(TAB_NOTES[note]), text_style)
}

func (score *TabScore) NewTablature(measures int) {
	if score.tab_started {
		score.canvas.Gend()
	}

	score.cur_tab++

	tab_height := measures * 9 * SYMBOL_HEIGHT
	if score.cur_tab == score.max_tabs {
		tab_height += SYMBOL_HEIGHT
	}

	offset_x := score.cur_tab*TAB_MARGIN_X + (score.cur_tab-1)*TAB_WIDTH
	offset_y := TAB_MARGIN_Y + MAX_TAB_HEIGHT - tab_height

	score.canvas.Translate(offset_x, offset_y)

	// Draw tablature spaces
	for i := 0; i < NUM_NOTES; i++ {
		if i < HALF_NOTES {
			// Draw notes going down
			score.drawTabNote(tab_height, i, i, (i+1)%3 == 0)
		} else {
			// Draw notes going up
			score.drawTabNote(tab_height, i, NUM_NOTES-i-1, (i+1)%3 == 0)
		}
	}

	// Draw central line
	center_x := HALF_NOTES*TABNOTE_WIDTH + MEASURE_THICKNESS/2
	line_height := tab_height + HALF_NOTES*TABNOTE_OFFSET_Y
	score.canvas.Line(center_x, 0, center_x, line_height, MEASURE_STYLE)

	score.current_y = tab_height
	score.tab_started = true
}

func findTrueLength(symb Symbol) int {
	if symb.Dotted() {
		return int(float64(symb.Length()) * 1.5)
	} else {
		return int(symb.Length())
	}
}

func findMeasures(symbols []Symbol) int {
	eighth_beats := 0

	for _, symb := range symbols {
		eighth_beats += findTrueLength(symb)
	}

	return int(math.Ceil(float64(eighth_beats) / 8))
}

func (tab *TabScore) DrawMeasureBar() {
	bar_y := tab.current_y - MEASURE_THICKNESS/2
	text_style := TEXT_STYLE + ";dominant-baseline:central"
	text_margin_left := 2
	tab.measure++

	tab.canvas.Line(0, bar_y, TAB_WIDTH, bar_y, MEASURE_STYLE)
	tab.canvas.Text(TAB_WIDTH+text_margin_left, bar_y,
		strconv.Itoa(tab.measure), text_style)

	tab.current_y -= SYMBOL_HEIGHT
}

// Return the x position that the provided pitch would be on.
func findNotePosition(pitch byte) (int, error) {
	notes := []byte{15, 13, 11, 9, 7, 5, 3, 1, 0, 2, 4, 6, 8, 10, 12, 14, 16}
	index := bytes.IndexByte(notes, pitch)
	if index < 0 {
		return -1, errors.New("Pitch out of range.")
	}

	return int(math.Ceil((float64(index) + 0.5) * TABNOTE_WIDTH)), nil
}

// Draw a note without any stem or taper and return its x position.
func (tab *TabScore) DrawPitch(note Note) (int, error) {
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
func (tab *TabScore) DrawStem(note_x int, length byte) {
	with_stem := length != WHOLE_NOTE
	tapered := length == EIGHTH_NOTE

	if with_stem {
		line_y := tab.current_y - NOTE_RADIUS
		tab.canvas.Line(-20, line_y, note_x, line_y, THIN_STYLE)

		if tapered {
			tab.canvas.Line(-20, line_y, -15, line_y-5, THIN_STYLE)
		}
	}
}

// Draw a note with a stem and taper when appropriate.
func (tab *TabScore) DrawNote(note Note) error {
	note_x, err := tab.DrawPitch(note)
	if err != nil {
		return err
	}

	tab.DrawStem(note_x, note.length)

	tab.current_y -= SYMBOL_HEIGHT * findTrueLength(note)
	return nil
}

func (tab *TabScore) DrawChord(chord Chord) error {
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

	tab.current_y -= SYMBOL_HEIGHT * findTrueLength(chord)
	return nil
}

func DrawScore(w io.Writer, symbols []Symbol) error {
	measures_left := findMeasures(symbols)
	score := NewScore(w, measures_left)
	score.NewTablature(MEASURES_PER_TAB)
	defer score.Close()

	eighth_beats := 0
	measures_done := -1
	for _, symb := range symbols {
		// Add a measure bar when necessary.
		if eighth_beats%8 == 0 {
			measures_done++

			if measures_done == MEASURES_PER_TAB {
				measures_left -= measures_done
				measures_done = 0

				if measures_left < MEASURES_PER_TAB {
					score.NewTablature(measures_left)
				} else {
					score.NewTablature(MEASURES_PER_TAB)
				}
			}

			score.DrawMeasureBar()
			eighth_beats = 0
		}

		// Add empty space for a note.
		switch symb.(type) {
		case Note:
			err := score.DrawNote(symb.(Note))
			if err != nil {
				return err
			}
		case Chord:
			err := score.DrawChord(symb.(Chord))
			if err != nil {
				return err
			}
		default:
			score.canvas.Circle(TAB_WIDTH/2, score.current_y, NOTE_RADIUS, "fill:green")
			score.current_y -= int(symb.Length()) * SYMBOL_HEIGHT
		}

		eighth_beats += int(symb.Length())
		if symb.Dotted() {
			eighth_beats += int(symb.Length()) / 2
		}

		if eighth_beats > 8 {
			return errors.New("Uneven beats in measure (expected 8 eighth beats, received " +
				strconv.Itoa(eighth_beats) + ").")
		}
	}

	// Add ending lines
	end_y := MEASURE_THICKNESS / 2
	score.canvas.Line(0, end_y, TAB_WIDTH, end_y, MEASURE_STYLE)
	end_y += SYMBOL_HEIGHT / 2
	score.canvas.Line(0, end_y, TAB_WIDTH, end_y, THIN_STYLE)

	return nil
}
