package util

import (
	"regexp"
	"strings"
)

func FormatError(errmsg string) string {
	errmsg = strings.TrimSpace(errmsg)
	errmsg = "│ " + strings.ReplaceAll(errmsg, "\n", "\n│ ")
	return errmsg
}

func HeadContent(content string) string {
	r := regexp.MustCompile(`[\s\n]+`)
	content = strings.TrimSpace(r.ReplaceAllString(content, " "))

	if runes := []rune(content); len(runes) > 30 {
		content = string(runes[:30])
	}

	return content
}
