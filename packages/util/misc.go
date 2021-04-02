package util

import "time"

func Short(s string) string {
	if len(s) <= 6 {
		return s
	}
	return s[:6] + ".."
}

func ContainsDuplicates(lst []string) bool {
	for i := range lst {
		for j := i + 1; j < len(lst); j++ {
			if lst[i] == lst[j] {
				return true
			}
		}
	}
	return false
}

func NanoSecToUnixSec(ts int64) uint32 {
	return uint32(ts / int64(time.Second))
}

func UnixAfterSec(sec int) uint32 {
	return TimeNowUnix() + uint32(sec)
}

func TimeNowUnix() uint32 {
	return uint32(time.Now().Unix())
}
