package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

func TestOffLedgerRequestCodec(t *testing.T) {
	offLedgerRequest := isc.NewOffLedgerRequest(isc.NewMessage(3, 14), 123, 200).Sign(cryptolib.NewKeyPair())
	bcs.TestCodec(t, offLedgerRequest.(*isc.OffLedgerRequestData))

	offLedgerRequest = isc.NewOffLedgerRequest(isc.NewMessage(3, 14), 123, 200).Sign(cryptolib.TestKeyPair)
	bcs.TestCodecAndHash(t, offLedgerRequest.(*isc.OffLedgerRequestData), "0b76ea31b34a")
}
