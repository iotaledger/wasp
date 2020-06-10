package state

import (
	"testing"
	"time"
)

func TestTs(t *testing.T) {
	var timestamp time.Time
	ts := timestamp.UnixNano()

	t.Logf("%v    %d", timestamp, ts)
	t.Logf("%v", time.Unix(0, 0))
	//assert.EqualValues(t, ts, int64(0))

}
