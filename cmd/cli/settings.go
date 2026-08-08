package cli

import (
	"bufio"
	"os"
	"strings"
)

const (
	altScreenOn  = "\033[?1049h"
	altScreenOff = "\033[?1049l"
	cursorHide   = "\033[?25l"
	cursorShow   = "\033[?25h"
	cursorHome   = "\033[H"
	clearBelow   = "\033[0J"
	clearAll     = "\033[2J\033[H"

	wrapOff = "\033[?7l"
	wrapOn  = "\033[?7h"
)

type Screen struct {
	out  *bufio.Writer
	live bool
}

func NewScreen(live bool) *Screen {
	s := &Screen{out: bufio.NewWriterSize(os.Stdout, 64<<10), live: live}
	if live {
		s.out.WriteString(altScreenOn + cursorHide + wrapOff)
		s.out.Flush()
	}
	return s
}

func (s *Screen) Frame(frame string) error {
	if s.live {
		s.out.WriteString(cursorHome)
		s.out.WriteString(strings.TrimSuffix(frame, "\n"))
		s.out.WriteString(clearBelow)
	} else {
		s.out.WriteString(clearAll)
		s.out.WriteString(frame)
	}
	return s.out.Flush()
}

func (s *Screen) Close() error {
	if s.live {
		s.out.WriteString(wrapOn + cursorShow + altScreenOff)
	}
	return s.out.Flush()
}
