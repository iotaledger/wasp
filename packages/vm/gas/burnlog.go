package gas

import (
	"fmt"
	"strings"
)

type BurnRecord struct {
	Code      BurnCode `json:"code" swagger:"required"`
	GasBurned uint64   `json:"gasBurned" swagger:"required"`
}

type BurnLog struct {
	Records []BurnRecord `json:"records" swagger:"required"`
}

func NewGasBurnLog() *BurnLog {
	return &BurnLog{Records: make([]BurnRecord, 0)}
}

func (h *BurnLog) Record(code BurnCode, gas uint64) {
	if h != nil {
		h.Records = append(h.Records, BurnRecord{code, gas})
	}
}

func (h *BurnLog) String() string {
	if h == nil {
		return "(no burn history)"
	}
	ret := make([]string, 0, len(h.Records)+2)
	var total uint64
	for i := range h.Records {
		ret = append(ret, fmt.Sprintf("%10s: %d", h.Records[i].Code.Name(), h.Records[i].GasBurned))
		total += h.Records[i].GasBurned
	}
	ret = append(ret, "---------------", fmt.Sprintf("%10s: %d", "TOTAL", total))
	return strings.Join(ret, "\n")
}
