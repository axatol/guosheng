package util

import "fmt"

func MDLink(text, href string) string {
	return fmt.Sprintf("[%s](%s)", text, href)
}
