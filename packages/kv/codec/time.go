package codec

import (
	"errors"
	"time"
)

func DecodeTime(b []byte, def ...time.Time) (ret time.Time, err error) {
	if b == nil {
		if len(def) == 0 {
			return ret, errors.New("cannot decode nil time")
		}
		return def[0], nil
	}
	nanos, err := DecodeInt64(b)
	if err != nil || nanos == 0 {
		return ret, err
	}
	return time.Unix(0, nanos), nil
}

func MustDecodeTime(b []byte, def ...time.Time) time.Time {
	t, err := DecodeTime(b, def...)
	if err != nil {
		panic(err)
	}
	return t
}

func EncodeTime(value time.Time) []byte {
	if value.IsZero() {
		return make([]byte, 8)
	}
	return EncodeInt64(value.UnixNano())
}
