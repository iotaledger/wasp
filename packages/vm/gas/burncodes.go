package gas

import (
	"fmt"
	"strings"
)

// used for testing

type BurnCode uint8

const (
	BurnStorage = BurnCode(iota)
	BurnCallTargetNotFound
	BurnSandboxUtils
	BurnGetRequest
	BurnGetContractContext
	BurnGetCallerData
	BurnGetAllowance
	BurnGetStateAnchorInfo
	BurnGetBalance
	BurnCallContract
	BurnDeployContract
	BurnEmitEventFixed
	BurnTransferAllowance
	BurnSendL1Request
)

var codeNames = map[BurnCode]string{
	BurnStorage:            "storage",
	BurnCallTargetNotFound: "target not found",
	BurnGetRequest:         "req data",
	BurnGetContractContext: "sc context",
	BurnGetCallerData:      "caller data",
	BurnGetAllowance:       "allowance",
	BurnGetStateAnchorInfo: "anchor info",
	BurnGetBalance:         "balance",
	BurnCallContract:       "call",
	BurnDeployContract:     "deploy",
	BurnEmitEventFixed:     "event",
	BurnTransferAllowance:  "transfer",
	BurnSendL1Request:      "post req",
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
	ret := make([]string, 0, len(h.records)+2)
	var total uint64
	for i := range h.records {
		s, ok := codeNames[h.records[i].Code]
		if !ok {
			s = "(unkown)"
		}
		ret = append(ret, fmt.Sprintf("%10s: %d", s, h.records[i].GasBurned))
		total += h.records[i].GasBurned
	}
	ret = append(ret, fmt.Sprintf("---------------"), fmt.Sprintf("%10s: %d", "TOTAL", total))
	return strings.Join(ret, "\n")
}
