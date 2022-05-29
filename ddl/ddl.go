package ddl

import (
	"bufio"
	"strings"
	"unicode/utf8"
)

// ParseReferencedTables parse DDL of view table and list tables referenced by view table.
func ParseReferencedTables(src string) []string {
	scanner := bufio.NewScanner(strings.NewReader(src))
	// original: bufio.ScanWords()
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		sq := false // '***'
		dq := false // "***"
		bq := false // `***`

		// Skip leading spaces.
		start := 0
		for width := 0; start < len(data); start += width {
			var r rune
			r, width = utf8.DecodeRune(data[start:])
			if !isSpace(r) && !isSkipSymbol(r) {
				break
			}
		}
		// Scan until space, marking end of word.
		for width, i := 0, start; i < len(data); i += width {
			if data[i] == '\'' {
				sq = !sq
			}
			if data[i] == '"' {
				dq = !dq
			}
			if data[i] == '`' {
				bq = !bq
			}

			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if isSpace(r) || isSkipSymbol(r) {
				if !sq && !dq && !bq {
					return i + width, data[start:i], nil
				}
			}
		}
		// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
		if atEOF && len(data) > start {
			return len(data), data[start:], nil
		}
		// Request more data.
		return start, nil, nil
	})

	tables := []string{}
	with := []string{}
	tFrom := false
	tJoin := false
	tWith := false
	for scanner.Scan() {
		token := scanner.Text()
		switch strings.ToUpper(token) {
		case "FROM":
			tFrom = true
		case "JOIN":
			tJoin = true
		case "WITH":
			tWith = true
		case "SELECT":
			tFrom = false
			tJoin = false
			tWith = false
		default:
			if tFrom {
				tables = append(tables, strings.Replace(token, "`", "", -1))
			}
			if tJoin {
				tables = append(tables, strings.Replace(token, "`", "", -1))
			}
			if tWith {
				with = append(with, strings.Replace(token, "`", "", -1))
			}
			tFrom = false
			tJoin = false
			tWith = false
		}
	}

	result := []string{}
	for _, t := range tables {
		if contains(with, t) {
			continue
		}
		result = append(result, t)
	}
	return unique(result)
}

func isSkipSymbol(r rune) bool {
	switch r {
	case ',':
		return true
	case '+', '-', '*', '/', '%':
		return true
	case '=', '<', '>':
		return true
	case '(', ')':
		return true
	case '&', '|':
		return true
	}
	return false
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func unique(in []string) []string {
	u := []string{}
	m := map[string]struct{}{}
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		u = append(u, s)
		m[s] = struct{}{}
	}
	return u
}
