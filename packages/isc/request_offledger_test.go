package isc_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestOffLedgerRequestCodec(t *testing.T) {
	offLedgerRequest := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.NewMessage(3, 14), 123, 200).Sign(cryptolib.NewKeyPair())
	bcs.TestCodec(t, offLedgerRequest.(*isc.OffLedgerRequestData))
}
