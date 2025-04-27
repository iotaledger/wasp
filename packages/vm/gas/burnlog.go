package gas

import (
	"fmt"
	"strings"

	bcs "github.com/iotaledger/bcs-go"
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

func (l *BurnLog) UnmarshalBCS(d *bcs.Decoder) error {
	recordLen := d.ReadLen()
	l.Records = make([]BurnRecord, recordLen)

	for i := 0; i < recordLen; i++ {
		name := d.ReadString()
		if err := d.Err(); err != nil {
			return err
		}

		l.Records[i] = BurnRecord{
			Code:      BurnCodeFromName(name),
			GasBurned: d.ReadUint64(),
		}
	}

	return nil
}

func (l *BurnLog) MarshalBCS(e *bcs.Encoder) error {
	e.WriteLen(len(l.Records))

	for _, record := range l.Records {
		e.WriteString(record.Code.Name())
		e.WriteUint64(record.GasBurned)
	}

	return nil
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
