package util

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
