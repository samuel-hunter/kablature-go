package main

import (
	"github.com/ajstarks/svgo"
	"io"
	"strconv"
)

const (
	TAB_WIDTH    = TABNOTE_WIDTH * NUM_TABNOTES
	TAB_MARGIN_X = 50
	TAB_MARGIN_Y = 10
	TAB_NOTES    = "BGECAFDCEGBDFAC"

	NUM_TABNOTES     = 15
	TABNOTE_OFFSET_Y = 3
	TABNOTE_WIDTH    = 15
	TABNOTE_COLOR    = "white"
	TABNOTE_MARKED   = "#ee7c80"

	MEASURE_THICKNESS = 3
	FONTSIZE          = 10

	NOTE_DIAMETER = 8
	NUM_NOTES     = len(TAB_NOTES)
	HALF_NOTES    = NUM_NOTES / 2
)

const (
	WHOLE_NOTE = iota
	HALF_NOTE
	QUARTER_NOTE
	EIGHTH_NOTE

	WHOLE_REST
	HALF_REST
	QUARTER_REST
	EIGHTH_REST
)

func drawTabNote(canvas *svg.SVG, tab_height, note, offset_y int, marked bool) {
	x := TAB_MARGIN_X + note*TABNOTE_WIDTH
	rect_height := tab_height + offset_y*TABNOTE_OFFSET_Y

	line_style := "stroke-width:1;stroke:black"

	if marked {
		line_style += ";fill:" + TABNOTE_MARKED
	} else {
		line_style += ";fill:" + TABNOTE_COLOR
	}

	canvas.Rect(x, TAB_MARGIN_Y, TABNOTE_WIDTH, rect_height, line_style)

	text_style := "text-anchor:middle;font-size:" + strconv.Itoa(FONTSIZE) + ";fill:black"
	canvas.Text(x+TABNOTE_WIDTH/2, TAB_MARGIN_Y+rect_height+FONTSIZE,
		string(TAB_NOTES[note]), text_style)
}

func makeTemplate(w io.Writer) *svg.SVG {
	tab_height := 500

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
		"stroke-width:"+strconv.Itoa(MEASURE_THICKNESS)+";stroke:black")

	// canvas.Circle(width/2, height/2, 100)

	return canvas
}

func Hello(w io.Writer) {
	canvas := makeTemplate(w)
	canvas.End()
}
