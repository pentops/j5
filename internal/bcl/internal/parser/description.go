package parser

import (
	"regexp"
	"strings"
)

var bulletRegex = regexp.MustCompile(`-|[0-9]\.`)

func reformatDescription(input string, maxWidth int) []string {
	lines := strings.Split(input, "\n")
	linesOut := []string{}

	pend := ""
	lastWasEmpty := false
	for idx, line := range lines {
		// Check for paragraph breaks
		if idx > 0 && strings.TrimSpace(line) == "" {
			if pend != "" {
				// This is a line break between paragraphs, flush the pending line
				// TODO: Allows empty lines at the end of the description
				linesOut = append(linesOut, pend)
				pend = ""
			}

			// Add a line break, unless the previous line was also empty
			if !lastWasEmpty {
				linesOut = append(linesOut, "")
			}

			lastWasEmpty = true
			continue
		}
		lastWasEmpty = false

		words := strings.Split(line, " ")
		for wordIdx, word := range words {
			if wordIdx == 0 && bulletRegex.Match([]byte(word)) {
				// Bullet point, flush the pending line
				if pend != "" {
					linesOut = append(linesOut, pend)
					pend = ""
				}
			}

			if pend == "" {
				pend = word
				continue
			}

			if len(pend)+len(word) > maxWidth {
				linesOut = append(linesOut, pend)
				pend = word
				continue
			}

			pend += " " + word
		}
	}

	if pend != "" {
		linesOut = append(linesOut, pend)
	}

	return linesOut
}
