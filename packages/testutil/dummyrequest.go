package testutil

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
)

func DummyOffledgerRequest(chainID *iscp.ChainID) *request.OffLedger {
	contract := iscp.Hn("somecontract")
	entrypoint := iscp.Hn("someentrypoint")
	args := requestargs.New(dict.Dict{})
	req := request.NewOffLedger(chainID, contract, entrypoint, args)
	keys, _ := testkey.GenKeyAddr()
	req.Sign(keys)
	return req
}
