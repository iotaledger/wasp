package gas

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/util/bcs"
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

func LimitsFromBytes(data []byte) (*Limits, error) {
	v, err := bcs.Unmarshal[*Limits](data)

	if err != nil {
		return nil, err
	}

	if !v.IsValid() {
		return nil, fmt.Errorf("invalid gas limits")
	}

	return v, nil
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
	return bcs.MustMarshal(gl)
}

func (gl *Limits) String() string {
	return fmt.Sprintf(
		"GasLimits(max/block: %d, min/req: %d, max/req: %d, max/view: %d",
		gl.MaxGasPerBlock,
		gl.MinGasPerRequest,
		gl.MaxGasPerRequest,
		gl.MaxGasExternalViewCall,
	)
}
