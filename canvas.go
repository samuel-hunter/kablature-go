package main

import (
	"bytes"
	"errors"
	"fmt"
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

	BEATS_PER_MEASURE = 8
	MEASURES_PER_TAB  = 7

	// Calculated constants
	MAX_TAB_HEIGHT = (9*MEASURES_PER_TAB + 1) * SYMBOL_HEIGHT
	NUM_NOTES      = len(TAB_NOTES)
	HALF_NOTES     = NUM_NOTES / 2
	TAB_CENTER     = HALF_NOTES*TABNOTE_WIDTH + MEASURE_THICKNESS/2
)

var (
	THIN_STYLE    = "stroke-width:1;stroke:black"
	MEASURE_STYLE = "stroke-width:" + strconv.Itoa(MEASURE_THICKNESS) + ";stroke:black"
	TEXT_STYLE    = "font-size:" + strconv.Itoa(FONT_SIZE) + ";fill:black"
)

type TabScore struct {
	canvas         *svg.SVG
	cur_tab        int
	total_measures int

	tab_measures_left int
	tab_started       bool
	measure_beats     int
	current_y         int
	measure           int

	has_lonely_eighth bool // Whether the last note was an eighth note
	eighth_pos        int  // Position of previous eighth note
}

func NewScore(w io.Writer, total_measures int) *TabScore {
	tablatures := int(math.Ceil(float64(total_measures) / MEASURES_PER_TAB))

	width := tablatures*TAB_WIDTH + (tablatures+1)*TAB_MARGIN_X
	height := MAX_TAB_HEIGHT + TAB_MARGIN_Y*2 + HALF_NOTES*TABNOTE_OFFSET_Y + FONT_SIZE

	canvas := svg.New(w)
	canvas.Start(width, height)

	// Fill canvas with white background
	canvas.Rect(0, 0, width, height, "fill:white")

	score := &TabScore{canvas: canvas, total_measures: total_measures}
	score.AddMeasure()

	return score
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

func (score *TabScore) NewTablature() {
	if score.tab_started {
		score.canvas.Gend()
	}

	last_measure := false

	score.tab_measures_left = score.total_measures - MEASURES_PER_TAB*score.cur_tab
	if score.tab_measures_left >= MEASURES_PER_TAB {
		score.tab_measures_left = MEASURES_PER_TAB
		last_measure = true
	}

	score.cur_tab++

	tab_height := score.tab_measures_left * 9 * SYMBOL_HEIGHT
	if last_measure {
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
	line_height := tab_height + HALF_NOTES*TABNOTE_OFFSET_Y
	score.canvas.Line(TAB_CENTER, 0, TAB_CENTER, line_height, MEASURE_STYLE)

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

func (score *TabScore) AddMeasure() {
	if score.tab_measures_left == 0 {
		score.NewTablature()
	}

	bar_y := score.current_y - MEASURE_THICKNESS/2
	text_style := TEXT_STYLE + ";dominant-baseline:central"
	text_margin_left := 2
	score.measure++

	score.canvas.Line(0, bar_y, TAB_WIDTH, bar_y, MEASURE_STYLE)
	score.canvas.Text(TAB_WIDTH+text_margin_left, bar_y,
		strconv.Itoa(score.measure), text_style)

	score.tab_measures_left--
	score.current_y -= SYMBOL_HEIGHT
}

func (score *TabScore) MoveForward(sym Symbol) error {

	// Draw lonely taper when necessary
	if score.has_lonely_eighth && score.current_y != score.eighth_pos {
		score.DrawLonelyTaper()
	}

	length := findTrueLength(sym)
	score.current_y -= length * SYMBOL_HEIGHT

	score.measure_beats += length
	if score.measure_beats > 8 {
		return errors.New(fmt.Sprintf(
			"Expected %d beats in measure, received %d",
			BEATS_PER_MEASURE, score.measure_beats))
	} else if score.measure_beats == 8 {
		if score.has_lonely_eighth {
			score.DrawLonelyTaper()
		}
		score.AddMeasure()
		score.measure_beats = 0
	}

	return nil
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

func (score *TabScore) DrawLonelyTaper() {
	score.has_lonely_eighth = false
	score.canvas.Line(-20, score.eighth_pos-NOTE_RADIUS,
		-15, score.eighth_pos-NOTE_RADIUS-5, THIN_STYLE)
}

// Draw a note's stem and taper if appropriate.
func (score *TabScore) DrawStem(note_x int, length byte) {
	with_stem := length != WHOLE_NOTE
	tapered := length == EIGHTH_NOTE

	if with_stem {
		line_y := score.current_y - NOTE_RADIUS
		score.canvas.Line(-20, line_y, note_x, line_y, THIN_STYLE)

		if tapered {
			if score.has_lonely_eighth {
				score.canvas.Line(-20, line_y, -20, score.eighth_pos-NOTE_RADIUS, THIN_STYLE)
				score.has_lonely_eighth = false
			} else {
				score.has_lonely_eighth = true
				score.eighth_pos = score.current_y
			}
		}
	}
}

// Draw a note with a stem and taper when appropriate.
func (score *TabScore) AddNote(note Note) error {
	note_x, err := score.DrawPitch(note)
	if err != nil {
		return err
	}

	score.DrawStem(note_x, note.length)
	score.MoveForward(note)
	return nil
}

func (score *TabScore) AddChord(chord Chord) error {
	rightmost_x := 0

	for _, pitch := range chord.pitches {
		note_x, err := score.DrawPitch(Note{
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

	score.DrawStem(rightmost_x, chord.length)
	score.MoveForward(chord)
	return nil
}

func (score *TabScore) AddRest(rest Rest) {
	y := score.current_y

	switch rest.Length() {
	case WHOLE_NOTE:
		score.canvas.Rect(TAB_CENTER, y-NOTE_RADIUS/2,
			NOTE_RADIUS/2, NOTE_RADIUS, "fill:black")
	case HALF_NOTE:
		score.canvas.Rect(TAB_CENTER-NOTE_RADIUS/2, y-NOTE_RADIUS/2,
			NOTE_RADIUS/2, NOTE_RADIUS, "fill:black")
	case QUARTER_NOTE:
		score.canvas.Polyline(
			[]int{
				TAB_CENTER,
				TAB_CENTER + TABNOTE_WIDTH/2,
				TAB_CENTER + TABNOTE_WIDTH,
				TAB_CENTER + TABNOTE_WIDTH*4/3,
				TAB_CENTER + TABNOTE_WIDTH*5/3,
				TAB_CENTER + 2*TABNOTE_WIDTH,
			}, []int{
				y,
				y - TABNOTE_WIDTH/2,
				y,
				y - TABNOTE_WIDTH/2,
				y,
				y - TABNOTE_WIDTH/4,
			}, "stroke-width:2;stroke:black;fill:none")
	case EIGHTH_NOTE:
		score.canvas.Line(TAB_CENTER+3*TABNOTE_WIDTH/2, y-NOTE_RADIUS,
			TAB_CENTER+5*TABNOTE_WIDTH/2, y+NOTE_RADIUS,
			"stroke-width:2;stroke:black;fill:none")
		score.canvas.Circle(TAB_CENTER+7*TABNOTE_WIDTH/4, y+NOTE_RADIUS/2,
			NOTE_RADIUS*.75, "fill:black")
	}

	// score.canvas.Circle(TAB_CENTER, score.current_y, NOTE_RADIUS, "fill:green")
	score.MoveForward(rest)
}

func DrawScore(w io.Writer, symbols []Symbol) error {
	score := NewScore(w, findMeasures(symbols))
	defer score.Close()

	for _, symb := range symbols {
		switch symb.(type) {
		case Note:
			err := score.AddNote(symb.(Note))
			if err != nil {
				return err
			}
		case Chord:
			err := score.AddChord(symb.(Chord))
			if err != nil {
				return err
			}
		case Rest:
			score.AddRest(symb.(Rest))
		default:
			return errors.New(fmt.Sprintf("Unrecognized symbol %s.", symb))
		}
	}

	// Add ending lines
	end_y := MEASURE_THICKNESS / 2
	score.canvas.Line(0, end_y, TAB_WIDTH, end_y, MEASURE_STYLE)
	end_y += SYMBOL_HEIGHT / 2
	score.canvas.Line(0, end_y, TAB_WIDTH, end_y, THIN_STYLE)

	return nil
}
