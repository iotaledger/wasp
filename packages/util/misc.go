package util

func Short(s string) string {
	if len(s) <= 6 {
		return s
	}
	return s[:6] + ".."
}
