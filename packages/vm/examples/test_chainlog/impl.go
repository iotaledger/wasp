package test_chainlog

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

//Function used called times in log_test.go
func example_TestGeneric(ctx vmtypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	inc, ok, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, err
	}
	if !ok {
		inc = 1
	}
	ctx.ChainLog([]byte("Counter Number " + strconv.Itoa(int(inc))))

	return nil, nil

}
