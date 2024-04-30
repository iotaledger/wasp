package codec

import (
	"time"
)

var Time = NewCodec(decodeTime, encodeTime)

func decodeTime(b []byte) (ret time.Time, err error) {
	nanos, err := Int64.Decode(b)
	if err != nil || nanos == 0 {
		return ret, err
	}
	return time.Unix(0, nanos), nil
}

func encodeTime(value time.Time) []byte {
	if value.IsZero() {
		return make([]byte, 8)
	}
	return Int64.Encode(value.UnixNano())
}
