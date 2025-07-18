package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
)

func TestOffLedgerRequestCodec(t *testing.T) {
	offLedgerRequest := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(3, 14), 123, 200).Sign(cryptolib.NewKeyPair())
	bcs.TestCodec(t, offLedgerRequest.(*isc.OffLedgerRequestData))

	offLedgerRequest = isc.NewOffLedgerRequest(isctest.TestChainID, isc.NewMessage(3, 14), 123, 200).Sign(cryptolib.TestKeyPair)
	bcs.TestCodecAndHash(t, offLedgerRequest.(*isc.OffLedgerRequestData), "0b76ea31b34a")
}
