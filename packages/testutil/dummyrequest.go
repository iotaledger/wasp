package testutil

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
)

func DummyOffledgerRequest(chainID isc.ChainID) isc.OffLedgerRequest {
	contract := isc.Hn("somecontract")
	entrypoint := isc.Hn("someentrypoint")
	args := dict.Dict{}
	req := isc.NewOffLedgerRequest(chainID, contract, entrypoint, args, 0)
	keys, _ := testkey.GenKeyAddr()
	return req.Sign(keys)
}
