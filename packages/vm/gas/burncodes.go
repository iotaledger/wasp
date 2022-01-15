package gas

import (
	"fmt"
	"strings"
)

// used for testing

type BurnCode uint8

const (
	Storage = BurnCode(iota)
	CallTargetNotFound
	SandboxUtils
)

var codeNames = map[BurnCode]string{
	Storage: "storage",
}

type GasBurnRecord struct {
	Code      BurnCode
	GasBurned uint64
}

type GasBurnLog struct {
	records []GasBurnRecord
}

func NewGasBurnLog() *GasBurnLog {
	return &GasBurnLog{records: make([]GasBurnRecord, 0)}
}

func (h *GasBurnLog) Record(code BurnCode, gas uint64) {
	if h != nil {
		h.records = append(h.records, GasBurnRecord{code, gas})
	}
}

func (h *GasBurnLog) String() string {
	if h == nil {
		return "(no burn history)"
	}
	ret := make([]string, len(h.records))
	for i := range h.records {
		s, ok := codeNames[h.records[i].Code]
		if !ok {
			s = "(unkown)"
		}
		ret[i] = fmt.Sprintf("%10s: %d\n", s, h.records[i].GasBurned)
	}
	return strings.Join(ret, "\n")
}
