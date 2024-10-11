package codec

import (
	"time"
)

var Time = NewCodecFromBCS[time.Time]()
