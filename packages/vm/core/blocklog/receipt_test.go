package blocklog_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestReceiptCodec(t *testing.T) {
	bcs.TestCodec(t, blocklog.RequestReceipt{
		Request: isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(isc.Hn("0"),
			isc.Hn("0")), 123, gas.LimitsDefault.MaxGasPerRequest).Sign(cryptolib.NewKeyPair()),
		Error: &isc.UnresolvedVMError{
			ErrorCode: blocklog.ErrBlockNotFound.Code(),
			Params:    []isc.VMParam{1, 2, "string"},
		},
	})
}
