package util

import "strconv"

func DecodeUint64(numberAsString string) (uint64, error) {
	val, err := strconv.ParseInt(numberAsString, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint64(val), nil
}
