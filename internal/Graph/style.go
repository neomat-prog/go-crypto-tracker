package Graph

import (
	"os"
	"strings"
)

type ColorMode int

const (
	ColorNone ColorMode = iota
	Color16
	Color256
	ColorTrue
)

type Style struct {
	Color ColorMode
	ASCII bool
}

func DetectStyle() Style {
	var st Style

	term := os.Getenv("TERM")
	switch {
	case term == "" || term == "dumb":
		return Style{Color: ColorNone, ASCII: true}
	case isTrueColor(os.Getenv("COLORTERM")):
		st.Color = ColorTrue
	case strings.Contains(term, "256color"):
		st.Color = Color256
	default:
		st.Color = Color16
	}

	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		st.Color = ColorNone
	}

	st.ASCII = !localeIsUTF8()
	return st
}

func isTrueColor(colorterm string) bool {
	switch strings.ToLower(strings.TrimSpace(colorterm)) {
	case "truecolor", "24bit":
		return true
	}
	return false
}

func localeIsUTF8() bool {
	set := false
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(v) == "" {
			continue
		}
		set = true

		v = strings.ToLower(v)
		if strings.Contains(v, "utf-8") || strings.Contains(v, "utf8") {
			return true
		}
	}
	return !set
}

const (
	sgrReset   = "\x1b[0m"
	sgrReverse = "\x1b[7m"
)

func (s Style) up() string {
	switch s.Color {
	case ColorTrue:
		return "\x1b[38;2;38;166;154m"
	case Color256:
		return "\x1b[38;5;36m"
	case Color16:
		return "\x1b[92m"
	}
	return ""
}

func (s Style) down() string {
	switch s.Color {
	case ColorTrue:
		return "\x1b[38;2;239;83;80m"
	case Color256:
		return "\x1b[38;5;203m"
	case Color16:
		return "\x1b[91m"
	}
	return ""
}

func (s Style) axis() string {
	switch s.Color {
	case ColorTrue:
		return "\x1b[38;2;120;123;134m"
	case Color256:
		return "\x1b[38;5;244m"
	case Color16:
		return "\x1b[90m"
	}
	return ""
}

func (s Style) grid() string {
	switch s.Color {
	case ColorTrue:
		return "\x1b[38;2;60;63;74m"
	case Color256:
		return "\x1b[38;5;238m"
	case Color16:
		return "\x1b[90m"
	}
	return ""
}

func (s Style) reset() string {
	if s.Color == ColorNone {
		return ""
	}
	return sgrReset
}

func (s Style) dirColor(up bool) string {
	if up {
		return s.up()
	}
	return s.down()
}

type glyphSet struct {
	bodyFull  rune
	bodyUpper rune
	bodyLower rune

	wickFull  rune
	wickUpper rune
	wickLower rune

	lastPrice rune
	gridDot   rune
	axisTick  rune
}

var (
	unicodeGlyphs = glyphSet{
		bodyFull:  '█',
		bodyUpper: '▀',
		bodyLower: '▄',
		wickFull:  '│',
		wickUpper: '╵',
		wickLower: '╷',
		lastPrice: '┈',
		gridDot:   '·',
		axisTick:  '┤',
	}

	asciiGlyphs = glyphSet{
		bodyFull:  '#',
		bodyUpper: '#',
		bodyLower: '#',
		wickFull:  '|',
		wickUpper: '|',
		wickLower: '|',
		lastPrice: '-',
		gridDot:   '.',
		axisTick:  '|',
	}
)

func (s Style) glyphs() glyphSet {
	if s.ASCII {
		return asciiGlyphs
	}
	return unicodeGlyphs
}

func (s Style) subRowsPerRow() int {
	if s.ASCII {
		return 1
	}
	return 2
}
