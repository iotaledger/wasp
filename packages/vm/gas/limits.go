package gas

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

type Limits struct {
	MaxGasPerBlock         uint64 `json:"maxGasPerBlock" swagger:"desc(The maximum gas per block),required"`
	MinGasPerRequest       uint64 `json:"minGasPerRequest" swagger:"desc(The minimum gas per request),required"`
	MaxGasPerRequest       uint64 `json:"maxGasPerRequest" swagger:"desc(The maximum gas per request),required"`
	MaxGasExternalViewCall uint64 `json:"maxGasExternalViewCall" swagger:"desc(The maximum gas per external view call),required"`
}

var LimitsDefault = &Limits{
	MaxGasPerBlock:         1_000_000_000,
	MinGasPerRequest:       10_000,
	MaxGasPerRequest:       50_000_000, // 20 requests per block max
	MaxGasExternalViewCall: 50_000_000,
}

func MustLimitsFromBytes(data []byte) *Limits {
	ret, err := LimitsFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret
}

var ErrInvalidLimits = errors.New("invalid gas limits")

func LimitsFromBytes(data []byte) (*Limits, error) {
	return LimitsFromMarshalUtil(marshalutil.New(data))
}

func LimitsFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Limits, error) {
	ret := &Limits{}
	var err error
	if ret.MaxGasPerBlock, err = mu.ReadUint64(); err != nil {
		return nil, fmt.Errorf("unable to parse MaxGasPerBlock: %w", err)
	}
	if ret.MinGasPerRequest, err = mu.ReadUint64(); err != nil {
		return nil, fmt.Errorf("unable to parse MinGasPerRequest: %w", err)
	}
	if ret.MaxGasPerRequest, err = mu.ReadUint64(); err != nil {
		return nil, fmt.Errorf("unable to parse MaxGasPerRequest: %w", err)
	}
	if ret.MaxGasExternalViewCall, err = mu.ReadUint64(); err != nil {
		return nil, fmt.Errorf("unable to parse MaxGasExternalViewCall: %w", err)
	}
	if !ret.IsValid() {
		return nil, ErrInvalidLimits
	}
	return ret, nil
}

func (gl *Limits) IsValid() bool {
	if gl.MaxGasPerBlock == 0 {
		return false
	}
	if gl.MinGasPerRequest == 0 || gl.MinGasPerRequest > gl.MaxGasPerBlock {
		return false
	}
	if gl.MaxGasPerRequest < gl.MinGasPerRequest {
		return false
	}
	if gl.MaxGasExternalViewCall == 0 {
		return false
	}
	return true
}

func (gl *Limits) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint64(gl.MaxGasPerBlock)
	mu.WriteUint64(gl.MinGasPerRequest)
	mu.WriteUint64(gl.MaxGasPerRequest)
	mu.WriteUint64(gl.MaxGasExternalViewCall)
	return mu.Bytes()
}

func (gl *Limits) String() string {
	return fmt.Sprintf(
		"GasLimits(max/block: %d, min/req: %d, max/req: %d, max/view: %d",
		gl.MaxGasPerBlock,
		gl.MaxGasPerBlock,
		gl.MinGasPerRequest,
		gl.MaxGasExternalViewCall,
	)
}
