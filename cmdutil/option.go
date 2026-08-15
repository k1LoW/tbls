package cmdutil

import (
	"fmt"
	"strings"
)

func PickOption(args []string, opts []string) (string, []string) {
	var (
		v        string
		skipNext bool
	)
	remains := []string{}

L:
	for i, a := range args {
		for _, opt := range opts {
			switch {
			case a == opt:
				if i+1 >= len(args) {
					// The option has no value. Keep it in remains so that the caller can report it.
					break
				}
				v = args[i+1]
				skipNext = true
				continue L
			case strings.HasPrefix(a, fmt.Sprintf("%s=", opt)):
				_, v, _ = strings.Cut(a, "=")
				continue L
			}
		}
		if skipNext {
			skipNext = false
			continue
		}
		remains = append(remains, a)
	}

	return v, remains
}
