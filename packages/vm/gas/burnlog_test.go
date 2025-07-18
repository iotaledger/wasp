package gas_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestBurnLogSerialization(t *testing.T) {
	var burnLog gas.BurnLog
	burnLog.Records = []gas.BurnRecord{
		{
			Code:      gas.BurnCodeCallTargetNotFound,
			GasBurned: 10,
		},
		{
			Code:      gas.BurnCodeUtilsHashingSha3,
			GasBurned: 80,
		},
	}
	bcs.TestCodecAndHash(t, burnLog, "6a9acdb8be5b")
}
