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
			// A trailing option has no value. Leave it in remains so that the caller can
			// report it, rather than reading past the end of args.
			case a == opt && i+1 < len(args):
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
