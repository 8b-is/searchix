package components

import (
	"regexp"
)

var firstSentenceRegexp = regexp.MustCompile(`^.+?(\.[[:space:]]|:\n)`)

func firstSentence[T ~string](text T) T {
	if fs := firstSentenceRegexp.FindString(string(text)); fs != "" {
		return T(fs)
	}

	return text
}
