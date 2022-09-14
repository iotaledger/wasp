package gas

import (
	"fmt"
	"strings"
)

type BurnRecord struct {
	Code      BurnCode
	GasBurned uint64
}

type BurnLog struct {
	records []BurnRecord
}

func NewGasBurnLog() *BurnLog {
	return &BurnLog{records: make([]BurnRecord, 0)}
}

func (h *BurnLog) Record(code BurnCode, gas uint64) {
	if h != nil {
		h.records = append(h.records, BurnRecord{code, gas})
	}
}

func (h *BurnLog) String() string {
	if h == nil {
		return "(no burn history)"
	}
	ret := make([]string, 0, len(h.records)+2)
	var total uint64
	for i := range h.records {
		ret = append(ret, fmt.Sprintf("%10s: %d", h.records[i].Code.Name(), h.records[i].GasBurned))
		total += h.records[i].GasBurned
	}
	ret = append(ret, "---------------", fmt.Sprintf("%10s: %d", "TOTAL", total))
	return strings.Join(ret, "\n")
}
