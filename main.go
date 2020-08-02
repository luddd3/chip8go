package main

import (
	"fmt"
	"os"

	"github.com/luddd3/chip8/chip"
	"github.com/gdamore/tcell"
)

func main() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	screen.Clear()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	screen.Clear()

	chip := chip.New(&screen)

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()
}
