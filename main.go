package main

import (
	"errors"
	"log"
	"os"

	"github.com/gdamore/tcell"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return errors.New("Usage: txt <filepath>")
	}
	filePath := os.Args[1]

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	err = screen.Init()
	if err != nil {
		return err
	}
	defer screen.Fini()

	unsaved := false
	initialText, err := readFile(filePath)
	if err != nil {
		return err
	}
	if initialText == nil || len(initialText) == 0 {
		unsaved = true
		initialText = [][]rune{{}}
	}

	editor := &editor{
		text:     initialText,
		cursor:   point{0, 0},
		filePath: filePath,
		unsaved:  unsaved,
		mode:     modeInsert,
		history:  make([]editorHistory, 0),
		query:    make([]rune, 0),
	}

	renderer := &renderer{}
	renderer.init(screen)
	renderer.render(screen, editor)

	for {
		switch event := screen.PollEvent().(type) {
		case *tcell.EventKey:
			done, err := editor.handleEventKey(event)
			if done {
				return err
			}
			renderer.render(screen, editor)
		case *tcell.EventResize:
			renderer.render(screen, editor)
		}
	}

}
