package test_sandbox

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

// testChainLogGenericData is called several times in log_test.go
func testChainLogGenericData(ctx vmtypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	ctx.ChainLog([]byte(fmt.Sprintf("[TRGenericData] Counter Number: %d", inc)))
	return nil, nil
}

func testChainLogEventData(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Event("[TREvent] - Testing Event...")
	return nil, nil
}

func testChainLogEventDataFormatted(ctx vmtypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	ctx.Eventf("[TREvent] - (%d) - Testing Event...", inc)

	return nil, nil
}
