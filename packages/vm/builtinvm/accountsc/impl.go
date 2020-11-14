package accountsc

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize the init call
func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("accountsc.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("accountsc.initialize.fail: already_initialized")
	}
	ctx.Eventf("accountsc.initialize.success hname = %s", Hname.String())
	return nil, nil
}

// getBalance returns colored balances of the account belonging to the AgentID
func getBalance(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	aid, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamsAgentIDNotFound
	}
	// each account has own map with name of AgentID
	account := ctx.State().GetMap(kv.Key(aid[:]))
	ret := codec.NewCodec(dict.NewDict())
	// expected balance.Color: int64 but no type control here
	account.Iterate(func(key []byte, value []byte) bool {
		ret.Set(kv.Key(key), value)
		return true
	})
	return ret, nil
}

// TODO inter chain transfer
// transfer moves tokens from smart contract account to another account inside same chain
func transfer(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	params := ctx.Params()
	targetAgentID, ok, err := params.GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamsAgentIDNotFound
	}
	sourceAgentID := ctx.MyAgentID()

	state := ctx.AccessState()
	sourceAccount := state.GetMap(kv.Key(sourceAgentID[:]))
	targetAccount := state.GetMap(kv.Key(targetAgentID[:]))

	// first check balances
	tran := ctx.Accounts().Transfer()
	balancesOk := true
	tran.Iterate(func(col balance.Color, bal int64) bool {
		var sourceBalance int64
		v := sourceAccount.GetAt(col[:])
		if v != nil {
			sourceBalance = int64(util.Uint64From8Bytes(v))
		}
		if sourceBalance < bal {
			balancesOk = false
			return false
		}
		return true
	})
	if !balancesOk {
		return nil, ErrNotEnoughBalance
	}
	tran.Iterate(func(col balance.Color, bal int64) bool {
		var sourceBalance, targetBalance int64
		v := sourceAccount.GetAt(col[:])
		if v != nil {
			sourceBalance = int64(util.Uint64From8Bytes(v))
		}
		v = targetAccount.GetAt(col[:])
		if v != nil {
			targetBalance = int64(util.Uint64From8Bytes(v))
		}
		targetAccount.SetAt(col[:], util.Uint64To8Bytes(uint64(targetBalance+bal)))
		if sourceBalance == bal {
			sourceAccount.DelAt(col[:])
		} else {
			targetAccount.SetAt(col[:], util.Uint64To8Bytes(uint64(sourceBalance-bal)))
		}
		return true
	})
	return nil, nil
}
