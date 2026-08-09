package Graph

import (
	"os"
	"testing"
)

var styleEnvKeys = []string{"TERM", "COLORTERM", "NO_COLOR", "LC_ALL", "LC_CTYPE", "LANG"}

func TestDetectStyle(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want Style
	}{
		{
			name: "modern terminal",
			env:  map[string]string{"TERM": "xterm-256color", "COLORTERM": "truecolor", "LANG": "en_US.UTF-8"},
			want: Style{Color: ColorTrue},
		},
		{
			name: "256 color without COLORTERM",
			env:  map[string]string{"TERM": "screen-256color", "LANG": "en_US.UTF-8"},
			want: Style{Color: Color256},
		},
		{
			name: "plain xterm",
			env:  map[string]string{"TERM": "xterm", "LANG": "en_US.UTF-8"},
			want: Style{Color: Color16},
		},
		{
			name: "NO_COLOR wins over COLORTERM",
			env:  map[string]string{"TERM": "xterm-256color", "COLORTERM": "truecolor", "NO_COLOR": "1", "LANG": "en_US.UTF-8"},
			want: Style{Color: ColorNone},
		},
		{
			name: "empty NO_COLOR still counts",
			env:  map[string]string{"TERM": "xterm-256color", "NO_COLOR": "", "LANG": "en_US.UTF-8"},
			want: Style{Color: ColorNone},
		},
		{
			name: "non UTF-8 locale drops to ASCII",
			env:  map[string]string{"TERM": "xterm-256color", "COLORTERM": "truecolor", "LANG": "C"},
			want: Style{Color: ColorTrue, ASCII: true},
		},
		{
			name: "LC_ALL overrides a UTF-8 LANG",
			env:  map[string]string{"TERM": "xterm", "LC_ALL": "POSIX", "LANG": "C"},
			want: Style{Color: Color16, ASCII: true},
		},
		{
			name: "unset locale stays Unicode",
			env:  map[string]string{"TERM": "xterm-256color"},
			want: Style{Color: Color256},
		},
		{
			name: "dumb terminal gets nothing",
			env:  map[string]string{"TERM": "dumb", "COLORTERM": "truecolor", "LANG": "en_US.UTF-8"},
			want: Style{Color: ColorNone, ASCII: true},
		},
		{
			name: "no TERM at all",
			env:  map[string]string{},
			want: Style{Color: ColorNone, ASCII: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range styleEnvKeys {
				t.Setenv(key, "")
				if value, ok := tc.env[key]; ok {
					os.Setenv(key, value)
				} else {
					os.Unsetenv(key)
				}
			}

			if got := DetectStyle(); got != tc.want {
				t.Errorf("DetectStyle() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
