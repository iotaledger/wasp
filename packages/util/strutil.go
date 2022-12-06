package util

import "time"

const (
	ending = "[..]"
)

func GentleTruncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= len(ending) {
		return ending
	}
	return s[:length-len(ending)] + ending
}

func TimeOrNever2(t time.Time, never string) string {
	timestampNever := time.Time{}
	if t == timestampNever {
		return never
	}
	return t.UTC().Format(time.RFC3339)
}

func TimeOrNever(t time.Time) string {
	return TimeOrNever2(t, "never")
}
