package utils

import (
	"sort"
	"strings"
)

func SortLines(s string) string {
	lines := strings.Split(s, "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func MultilinePreview(text string) string {
	// TODO: This action might take huge amount of time for big texts. Maybe just dont do it.
	return MultilinePreviewWithOpts(text, 6, 5, "\t...")
}

func MultilinePreviewWithOpts(text string, linesAtStart, linesAtEnd int, sep string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= (linesAtStart + linesAtEnd) {
		return text
	}

	startLines := lines[:linesAtStart]
	endLines := lines[len(lines)-linesAtEnd:]

	startLines = append(startLines, sep)

	return strings.Join(append(startLines, endLines...), "\n")
}
