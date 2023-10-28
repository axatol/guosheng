package util

import "strings"

func Obscure(input string, start int, end ...int) string {
	var b strings.Builder
	for i, r := range input {
		if i < start || (len(end) == 1 && i > end[0]) {
			b.WriteRune(r)
		} else {
			b.WriteRune('*')
		}
	}

	return b.String()
}
