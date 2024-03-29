// canvas - Transcribing the tablature to SVG.

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
	TABNOTE_MARKED   = "salmon"

	MEASURE_THICKNESS = 3
	FONT_SIZE         = 10

	NOTE_RADIUS   = 4
	SYMBOL_HEIGHT = TABNOTE_WIDTH // Spacing a general musical symbol would have allocated.

	// Calculated constants
	NUM_NOTES  = len(TAB_NOTES)
	HALF_NOTES = NUM_NOTES / 2
	TAB_CENTER = HALF_NOTES*TABNOTE_WIDTH + MEASURE_THICKNESS/2
)

var (
	THIN_STYLE    = "stroke-width:1;stroke:black"
	MEASURE_STYLE = "stroke-width:" + strconv.Itoa(MEASURE_THICKNESS) + ";stroke:black"
	TEXT_STYLE    = "font-size:" + strconv.Itoa(FONT_SIZE) + ";fill:black"
)

type tabScore struct {
	canvas         *svg.SVG
	cur_tab        int
	total_measures int // total measures the score will account for

	tab_measures_left int
	tab_started       bool
	measure_beats     int
	current_y         int
	measure           int

	has_lonely_eighth bool // Whether the last note was an eighth note
	eighth_pos        int  // Position of previous eighth note
}

// Return the total tablatures the tab score should have.
func (score *tabScore) totalTabs() int {
	return int(math.Ceil(float64(score.total_measures) / float64(GlobalConfig.MeasuresPerTab)))
}

// Return the calculated height of a tablature.
func tabHeight(measures int, withEndSymbol bool) int {
	result := ((GlobalConfig.BeatsPerMeasure + 1) * measures) * SYMBOL_HEIGHT
	if withEndSymbol {
		result += SYMBOL_HEIGHT
	}

	return result
}

// Return the highest possible measure of a tablature.
func maxTabHeight() int {
	return tabHeight(GlobalConfig.MeasuresPerTab, true)
}

// Construct a new tablature score.
func newScore(w io.Writer, total_measures int) *tabScore {
	canvas := svg.New(w)
	score := &tabScore{canvas: canvas, total_measures: total_measures}

	width := score.totalTabs()*TAB_WIDTH + (score.totalTabs()+1)*TAB_MARGIN_X
	height := maxTabHeight() + TAB_MARGIN_Y*2 + HALF_NOTES*TABNOTE_OFFSET_Y + FONT_SIZE

	canvas.Start(width, height)

	// Fill canvas with white background
	canvas.Rect(0, 0, width, height, "fill:white")

	return score
}

// Add the ending notes to the tablature if started, and wrap
// everything up.
func (score *tabScore) close() {
	if score.tab_started {
		// Add ending lines to last measure.
		end_y := MEASURE_THICKNESS / 2
		score.canvas.Line(0, end_y, TAB_WIDTH, end_y, MEASURE_STYLE)
		end_y += SYMBOL_HEIGHT / 2
		score.canvas.Line(0, end_y, TAB_WIDTH, end_y, THIN_STYLE)

		score.canvas.Gend()
		score.tab_started = false
	}

	score.canvas.End()
}

// Draw the space for a note across the entire tablature.
func (score *tabScore) drawNoteSpace(tab_height, note, offset_y int, marked bool) {
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

func (score *tabScore) newTablature() {
	if score.tab_started {
		score.canvas.Gend()
	}

	last_measure := false

	score.tab_measures_left = score.total_measures - GlobalConfig.MeasuresPerTab*score.cur_tab
	if score.tab_measures_left > GlobalConfig.MeasuresPerTab {
		score.tab_measures_left = GlobalConfig.MeasuresPerTab
	} else {
		last_measure = true
	}

	score.cur_tab++

	tab_height := tabHeight(score.tab_measures_left, last_measure)
	offset_x := score.cur_tab*TAB_MARGIN_X + (score.cur_tab-1)*TAB_WIDTH
	offset_y := TAB_MARGIN_Y + maxTabHeight() - tab_height

	score.canvas.Translate(offset_x, offset_y)

	// Draw tablature spaces
	for i := 0; i < NUM_NOTES; i++ {
		if i < HALF_NOTES {
			// Draw notes going down
			score.drawNoteSpace(tab_height, i, i, (i+1)%3 == 0)
		} else {
			// Draw notes going up
			score.drawNoteSpace(tab_height, i, NUM_NOTES-i-1, (i+1)%3 == 0)
		}
	}

	// Draw central line
	line_height := tab_height + HALF_NOTES*TABNOTE_OFFSET_Y
	score.canvas.Line(TAB_CENTER, 0, TAB_CENTER, line_height, MEASURE_STYLE)

	score.current_y = tab_height
	score.tab_started = true
}

// Count the measures needed to fit in all musical symbols in the
// given slice.
func countMeasures(symbols []Symbol) int {
	eighth_beats := 0

	for _, symb := range symbols {
		eighth_beats += SymbolLength(symb)
	}

	result := int(math.Ceil(float64(eighth_beats) / float64(GlobalConfig.BeatsPerMeasure)))
	return result
}

// Add a new measure to the tablature score to start adding symbols
// to.
func (score *tabScore) addMeasure() {
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
func (tab *tabScore) drawPitch(note Note) (int, error) {
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

// Draw the taper to a lonely eighth note.
func (score *tabScore) drawLonelyTaper() {
	score.has_lonely_eighth = false
	score.canvas.Line(-20, score.eighth_pos-NOTE_RADIUS,
		-15, score.eighth_pos-NOTE_RADIUS-5, THIN_STYLE)
}

// Draw a note's stem and taper if appropriate.
func (score *tabScore) drawStem(note_x int, length byte) {
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

// Add a note to the current tablature with a stem and taper when
// appropriate.
func (score *tabScore) addNote(note Note) error {
	note_x, err := score.drawPitch(note)
	if err != nil {
		return err
	}

	score.drawStem(note_x, note.length)
	return nil
}

// Add a chord to the current tablature.
func (score *tabScore) addChord(chord Chord) error {
	rightmost_x := 0

	for _, pitch := range chord.pitches {
		note_x, err := score.drawPitch(Note{
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

	score.drawStem(rightmost_x, chord.length)
	return nil
}

// Add a rest to the current tablature.
func (score *tabScore) addRest(rest Rest) {
	y := score.current_y

	switch rest.length {
	case WHOLE_NOTE:
		// Draw a rectangle from the center right to signify a whole rest.
		score.canvas.Rect(TAB_CENTER, y-NOTE_RADIUS/2,
			NOTE_RADIUS/2, NOTE_RADIUS, "fill:black")
	case HALF_NOTE:
		// Draw a rectangle from the center left to signify a half rest.
		score.canvas.Rect(TAB_CENTER-NOTE_RADIUS/2, y-NOTE_RADIUS/2,
			NOTE_RADIUS/2, NOTE_RADIUS, "fill:black")
	case QUARTER_NOTE:
		// Draw a squiggly line from the center to the right to
		// signify a quarter rest.
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
		// Draw a diagonal line with a small tick to signify an eighth rest.
		score.canvas.Line(TAB_CENTER+3*TABNOTE_WIDTH/2, y-NOTE_RADIUS,
			TAB_CENTER+5*TABNOTE_WIDTH/2, y+NOTE_RADIUS,
			"stroke-width:2;stroke:black;fill:none")
		score.canvas.Circle(TAB_CENTER+7*TABNOTE_WIDTH/4, y+NOTE_RADIUS/2,
			NOTE_RADIUS*.75, "fill:black")
	}

}

// Add a symbol to the tablature score, adding measures and tablatures
// when necessary.
func (score *tabScore) addSymbol(sym Symbol) (err error) {

	if score.measure_beats%GlobalConfig.BeatsPerMeasure == 0 {
		if score.has_lonely_eighth {
			score.drawLonelyTaper()
		}

		if score.tab_measures_left == 0 {
			score.newTablature()
		}

		score.addMeasure()
		score.measure_beats = 0
	}

	switch sym.(type) {
	case Note:
		err = score.addNote(sym.(Note))
	case Chord:
		err = score.addChord(sym.(Chord))
	case Rest:
		score.addRest(sym.(Rest))
	default:
		err = errors.New(fmt.Sprintf("Unrecognized symbol %s.", sym))
	}

	// Draw lonely taper when necessary
	if score.has_lonely_eighth && score.eighth_pos != score.current_y {
		score.drawLonelyTaper()
	}

	// Add spacing to separate the previously drawn musical symbol with
	// any future symbols.
	length := SymbolLength(sym)
	score.current_y -= length * SYMBOL_HEIGHT

	score.measure_beats += length
	if score.measure_beats > GlobalConfig.BeatsPerMeasure {
		return errors.New(fmt.Sprintf(
			"Expected %d beats in measure, received %d",
			GlobalConfig.BeatsPerMeasure, score.measure_beats))
	}

	return err
}

// Write a complete tablature score to the writer in SVG format.
func DrawScore(w io.Writer, symbols []Symbol) error {
	score := newScore(w, countMeasures(symbols))
	defer score.close()

	for _, sym := range symbols {
		err := score.addSymbol(sym)
		if err != nil {
			return err
		}
	}

	return nil
}
