package bcs_test

import (
	"testing"
	"time"
)

func TestTimeCodec(t *testing.T) {
	testCodecNoRef(t, time.Unix(12345, 6789), []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})
}
