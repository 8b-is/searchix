package components

import (
	"regexp"
)

var firstSentenceRegexp = regexp.MustCompile(`^((?s).+?)(?:\.[[:space:]]*[[:upper:]]|$)`)

func firstSentence[T ~string](text T) T {
	if fs := firstSentenceRegexp.FindStringSubmatch(string(text)); len(fs) > 0 && fs[1] != "" {
		return T(fs[1])
	}

	return text
}
