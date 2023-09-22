package gas

import (
	"fmt"
	"io"
	"strings"

	"github.com/iotaledger/wasp/packages/util/rwutil"
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

func (l *BurnLog) Record(code BurnCode, gas uint64) {
	if l != nil {
		l.Records = append(l.Records, BurnRecord{code, gas})
	}
}

func (l *BurnLog) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	recordLen := rr.ReadUint32()
	l.Records = make([]BurnRecord, recordLen)
	for i := 0; i < int(recordLen); i++ {
		name := rr.ReadString()
		burnCode := BurnCodeFromName(name)
		gasBurned := rr.ReadUint64()

		l.Records[i] = BurnRecord{
			Code:      burnCode,
			GasBurned: gasBurned,
		}
	}
	return rr.Err
}

func (l *BurnLog) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	recordLen := len(l.Records)
	ww.WriteUint32(uint32(recordLen))
	for _, record := range l.Records {
		ww.WriteString(record.Code.Name())
		ww.WriteUint64(record.GasBurned)
	}
	return ww.Err
}

func (l *BurnLog) String() string {
	if l == nil {
		return "(no burn history)"
	}
	ret := make([]string, 0, len(l.Records)+2)
	var total uint64
	for i := range l.Records {
		ret = append(ret, fmt.Sprintf("%10s: %d", l.Records[i].Code.Name(), l.Records[i].GasBurned))
		total += l.Records[i].GasBurned
	}
	ret = append(ret, "---------------", fmt.Sprintf("%10s: %d", "TOTAL", total))
	return strings.Join(ret, "\n")
}
