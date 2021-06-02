package util

func StringInList(s string, lst []string) bool {
	for _, l := range lst {
		if l == s {
			return true
		}
	}
	return false
}

func AllDifferentStrings(lst []string) bool {
	for i := range lst {
		for j := range lst {
			if i >= j {
				continue
			}
			if lst[i] == lst[j] {
				return false
			}
		}
	}
	return true
}
