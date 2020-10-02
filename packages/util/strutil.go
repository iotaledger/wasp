package util

const (
	ending = "[..]"
)

func GentleCut(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= len(ending) {
		return ending
	}
	return s[:length-len(ending)] + ending
}
