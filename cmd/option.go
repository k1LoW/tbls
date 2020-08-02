package cmd

import (
	"fmt"
	"strings"
)

func pickOption(args []string, opts []string) (string, []string) {
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
				v = args[i+1]
				skipNext = true
				continue L
			case strings.HasPrefix(a, fmt.Sprintf("%s=", opt)):
				splited := strings.Split(a, "=")
				v = splited[1]
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
