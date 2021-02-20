package main

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/gdamore/tcell"
)

var styleMain = tcell.StyleDefault.
	Background(tcell.NewRGBColor(50, 50, 50)).
	Foreground(tcell.ColorWhite)

var styleStatusBar = tcell.StyleDefault.
	Background(tcell.ColorWhite).
	Foreground(tcell.NewRGBColor(50, 50, 50))

var styleLineNumbers = tcell.StyleDefault.
	Background(tcell.NewRGBColor(40, 40, 40)).
	Foreground(tcell.ColorWhite)

type renderer struct {
	offset point
}

func (r *renderer) init(screen tcell.Screen) {
	screen.SetStyle(styleMain)
}

func (r *renderer) render(screen tcell.Screen, e *editor) {
	screen.Clear()

	w, h := screen.Size()
	sizeLineNumbers := utf8.RuneCountInString(strconv.Itoa(len(e.text))) + 1

	// Adjust offset
	if e.cursor.x-r.offset.x < 0 {
		r.offset.x = e.cursor.x
	}
	if e.cursor.x-r.offset.x >= w-sizeLineNumbers {
		r.offset.x = e.cursor.x - (w - sizeLineNumbers) + 1
	}
	if e.cursor.y-r.offset.y < 0 {
		r.offset.y = e.cursor.y
	}
	if e.cursor.y-r.offset.y >= h-2 {
		r.offset.y = e.cursor.y - (h - 2) + 1
	}

	// Line numbers
	for y := 0; y < h-2; y++ {
		for x := 0; x < sizeLineNumbers; x++ {
			screen.SetContent(x, y+1, ' ', nil, styleLineNumbers)
		}
	}
	for y := range e.text[r.offset.y:] {
		if y >= h-2 {
			break
		}
		for x, r := range fmt.Sprintf(fmt.Sprintf("%%%dd", sizeLineNumbers-1), y+r.offset.y+1) {
			screen.SetContent(x, y+1, r, nil, styleLineNumbers)
		}
	}

	// Top status bar
	for x := 0; x < w; x++ {
		screen.SetContent(x, 0, ' ', nil, styleStatusBar)
	}
	fileText := e.filePath
	if e.unsaved {
		fileText += "*"
	}
	for idx, r := range fileText {
		screen.SetContent(idx+1, 0, r, nil, styleStatusBar)
	}

	// Bottom status bar
	for x := 0; x < w; x++ {
		screen.SetContent(x, h-1, ' ', nil, styleStatusBar)
	}
	for idx, r := range fmt.Sprintf("(%d,%d)", e.cursor.x+1, e.cursor.y+1) {
		screen.SetContent(idx+1, h-1, r, nil, styleStatusBar)
	}
	if e.mode == modeInsert {
		keysText := fmt.Sprintf("^F Find, ^Z Undo, ^S Save, ^Q Quit")
		for idx, r := range keysText {
			screen.SetContent(idx+w-len(keysText)-1, h-1, r, nil, styleStatusBar)
		}
	} else if e.mode == modeExiting {
		message := "EXIT: Save changes? [y/n] | Esc Back"
		for idx, r := range message {
			screen.SetContent(idx+w-len(message)-1, h-1, r, nil, styleStatusBar)
		}
	} else if e.mode == modeFind {
		message := "FIND: \"" + string(e.query) + "\" | Esc Back"
		for idx, r := range message {
			screen.SetContent(idx+w-len(message)-1, h-1, r, nil, styleStatusBar)
		}
	}

	// Text
	for y, line := range e.text[r.offset.y:] {
		if y >= h-2 {
			break
		}
		if len(line) < r.offset.x {
			continue
		}
		for x, r := range line[r.offset.x:] {
			if x >= w-sizeLineNumbers {
				break
			}
			screen.SetContent(x+sizeLineNumbers, y+1, r, nil, styleMain)
		}
	}

	// Cursor
	screen.ShowCursor(e.cursor.x-r.offset.x+sizeLineNumbers, e.cursor.y-r.offset.y+1)

	screen.Show()
}
