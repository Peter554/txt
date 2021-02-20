package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
)

type editorMode uint8

const (
	modeInsert editorMode = iota
	modeExiting
	modeFind
)

type editorHistory struct {
	text [][]rune
}

type editor struct {
	text     [][]rune
	cursor   point
	filePath string
	unsaved  bool
	mode     editorMode
	history  []editorHistory
	query    []rune
}

func (e *editor) handleEventKey(event *tcell.EventKey) (bool, error) {
	if e.mode == modeInsert {
		return e.handleEventKeyModeInsert(event)
	}
	if e.mode == modeExiting {
		return e.handleEventKeyModeExiting(event)
	}
	if e.mode == modeFind {
		return e.handleEventKeyModeFind(event)
	}
	return true, fmt.Errorf("Unexpected editor mode %d", e.mode)
}

func (e *editor) handleEventKeyModeInsert(event *tcell.EventKey) (bool, error) {
	x, y := e.cursor.x, e.cursor.y
	switch event.Key() {
	case tcell.KeyRune:
		e.pushHistory()
		r := event.Rune()
		t := e.text[y]
		e.text[y] = append(copyRunes(t[:x]), r)
		e.text[y] = append(e.text[y], t[x:]...)
		e.cursor.x++
		e.unsaved = true
	case tcell.KeyLeft:
		e.cursor.x = max(x-1, 0)
	case tcell.KeyRight:
		e.cursor.x = min(x+1, len(e.text[y]))
	case tcell.KeyUp:
		e.cursor.y = max(y-1, 0)
		e.cursor.x = min(x, len(e.text[e.cursor.y]))
	case tcell.KeyDown:
		e.cursor.y = min(y+1, len(e.text)-1)
		e.cursor.x = min(x, len(e.text[e.cursor.y]))
	case tcell.KeyHome:
		e.cursor.x = 0
	case tcell.KeyEnd:
		e.cursor.x = len(e.text[y])
	case tcell.KeyPgUp:
		e.cursor.y = 0
	case tcell.KeyPgDn:
		e.cursor.y = len(e.text) - 1
	case tcell.KeyEnter:
		e.pushHistory()
		t := e.text
		e.text = append(copyRunes2D(t[:y]), t[y][:x])
		e.text = append(e.text, t[y][x:])
		e.text = append(e.text, t[y+1:]...)
		e.cursor.x = 0
		e.cursor.y++
		e.unsaved = true
	case tcell.KeyDelete:
		if x == len(e.text[y]) && y == len(e.text)-1 {
			break
		}
		e.pushHistory()
		if x == len(e.text[y]) {
			t := e.text
			e.text = copyRunes2D(t[:y+1])
			e.text[y] = append(e.text[y], t[y+1]...)
			e.text = append(e.text, t[y+2:]...)
		} else {
			e.text[y] = append(copyRunes(e.text[y][:x]), e.text[y][x+1:]...)
			e.cursor.x = min(x, len(e.text[y]))
		}
		e.unsaved = true
	case tcell.KeyBackspace2:
		if x == 0 && y == 0 {
			break
		}
		e.pushHistory()
		if x == 0 {
			t := e.text
			e.text = copyRunes2D(t[:y])
			e.text[y-1] = append(e.text[y-1], t[y]...)
			e.text = append(e.text, t[y+1:]...)
			e.cursor.x = len(t[y-1])
			e.cursor.y--
		} else {
			e.text[y] = append(copyRunes(e.text[y][:x-1]), e.text[y][x:]...)
			e.cursor.x = max(x-1, 0)
		}
		e.unsaved = true
	case tcell.KeyCtrlF:
		e.mode = modeFind
	case tcell.KeyCtrlZ:
		e.popHistory()
	case tcell.KeyCtrlS:
		err := writeFile(e.filePath, e.text)
		if err != nil {
			return true, err
		}
		e.unsaved = false
	case tcell.KeyCtrlQ:
		if e.unsaved {
			e.mode = modeExiting
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (e *editor) handleEventKeyModeExiting(event *tcell.EventKey) (bool, error) {
	switch event.Key() {
	case tcell.KeyRune:
		r := event.Rune()
		if r == 'y' {
			err := writeFile(e.filePath, e.text)
			return true, err
		}
		if r == 'n' {
			return true, nil
		}
	case tcell.KeyEscape:
		e.mode = modeInsert
	}
	return false, nil
}

func (e *editor) handleEventKeyModeFind(event *tcell.EventKey) (bool, error) {
	switch event.Key() {
	case tcell.KeyRune:
		r := event.Rune()
		e.query = append(e.query, r)
	case tcell.KeyBackspace2:
		if len(e.query) == 0 {
			break
		}
		e.query = e.query[:len(e.query)-1]
	case tcell.KeyEnter:
		if len(e.query) == 0 {
			break
		}
		if e.cursor.x < len(e.text[e.cursor.y]) {
			if idx := strings.Index(string(e.text[e.cursor.y][e.cursor.x+1:]), string(e.query)); idx >= 0 {
				e.cursor.x += idx + 1
				break
			}
		}
		for y := e.cursor.y + 1; y < len(e.text); y++ {
			line := e.text[y]
			if idx := strings.Index(string(line), string(e.query)); idx >= 0 {
				e.cursor.y = y
				e.cursor.x = idx
				return false, nil
			}
		}
		for y := 0; y <= e.cursor.y; y++ {
			line := e.text[y]
			if idx := strings.Index(string(line), string(e.query)); idx >= 0 {
				e.cursor.y = y
				e.cursor.x = idx
			}
		}
	case tcell.KeyEscape:
		e.mode = modeInsert
		e.query = make([]rune, 0)
	}
	return false, nil
}

func (e *editor) pushHistory() {
	h := editorHistory{
		text: copyRunes2D(e.text),
	}
	e.history = append(e.history, h)
	historyDepth := 30
	if len(e.history) > historyDepth {
		e.history = e.history[len(e.history)-historyDepth:]
	}
}

func (e *editor) popHistory() {
	if len(e.history) == 0 {
		return
	}
	h := e.history[len(e.history)-1]
	e.history = e.history[:len(e.history)-1]
	e.text = h.text
	e.cursor.x = min(e.cursor.x, len(e.text[e.cursor.y]))
	e.cursor.y = min(e.cursor.y, len(e.text)-1)
}
