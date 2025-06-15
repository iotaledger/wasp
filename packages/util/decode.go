package util

import "fortio.org/safecast"

import "strconv"

func DecodeUint64(numberAsString string) (uint64, error) {
	val, err := strconv.ParseInt(numberAsString, 10, 64)
	if err != nil {
		return 0, err
	}

	result, err := safecast.Convert[uint64](val)
	if err != nil {
		return 0, err
	}
	return result, nil
}
