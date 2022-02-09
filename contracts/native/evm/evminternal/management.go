// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evminternal

import (
	"math"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	keyGasRatio     = "g"
	keyEVMOwner     = "o"
	keyNextEVMOwner = "n"
	keyBlockTime    = "b"

	// keyEVMState is the subrealm prefix for the EVM state
	keyEVMState = "s"
)

var ManagementHandlers = []coreutil.ProcessorEntryPoint{
	evm.FuncSetNextOwner.WithHandler(setNextOwner),
	evm.FuncClaimOwnership.WithHandler(claimOwnership),
	evm.FuncSetGasRatio.WithHandler(setGasRatio),
	evm.FuncGetOwner.WithHandler(getOwner),
	evm.FuncGetGasRatio.WithHandler(getGasRatio),
	evm.FuncSetBlockTime.WithHandler(setBlockTime),
}

func EVMStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyEVMState)
}

func InitializeManagement(ctx iscp.Sandbox) {
	ctx.State().Set(keyGasRatio, evm.DefaultGasRatio.Bytes())
	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(ctx.ContractCreator()))
}

func setBlockTime(ctx iscp.Sandbox) dict.Dict {
	requireOwner(ctx)

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	blockTime := params.MustGetUint32(evm.FieldBlockTime)
	a.Requiref(blockTime > 0, "blockTime must be > 0")

	mustSchedule := !ctx.State().MustHas(keyBlockTime)

	ctx.State().Set(keyBlockTime, codec.EncodeUint32(blockTime))
	if mustSchedule {
		ScheduleNextBlock(ctx)
	}
	return nil
}

func getBlockTime(state kv.KVStoreReader) uint32 {
	bt, _ := codec.DecodeUint32(state.MustGet(keyBlockTime), 0)
	return bt
}

func ScheduleNextBlock(ctx iscp.Sandbox) {
	requireOwner(ctx, true)

	blockTime := getBlockTime(ctx.State())
	if blockTime == 0 {
		return
	}

	ctx.Send(iscp.RequestParameters{
		TargetAddress:              ctx.ChainID().AsAddress(),
		Assets:                     iscp.NewAssets(1, nil),
		AdjustToMinimumDustDeposit: true,
		Metadata: &iscp.SendMetadata{
			TargetContract: ctx.Contract(),
			EntryPoint:     evm.FuncMintBlock.Hname(),
			GasBudget:      math.MaxUint64, // TODO: ?
		},
		Options: iscp.SendOptions{Timelock: &iscp.TimeData{
			Time: time.Unix(0, ctx.Timestamp()).
				Add(time.Duration(blockTime) * time.Second),
		}},
	})
}

func requireOwner(ctx iscp.Sandbox, allowSelf ...bool) {
	contractOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyEVMOwner))
	ctx.RequireNoError(err)

	allowed := []*iscp.AgentID{contractOwner}
	if len(allowSelf) > 0 && allowSelf[0] {
		allowed = append(allowed, iscp.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract()))
	}

	ctx.RequireCallerAnyOf(allowed)
}

func setNextOwner(ctx iscp.Sandbox) dict.Dict {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	ctx.State().Set(keyNextEVMOwner, codec.EncodeAgentID(par.MustGetAgentID(evm.FieldNextEVMOwner)))
	return nil
}

func claimOwnership(ctx iscp.Sandbox) dict.Dict {
	nextOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyNextEVMOwner))
	ctx.RequireNoError(err)
	ctx.RequireCaller(nextOwner)

	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(nextOwner))
	ScheduleNextBlock(ctx)
	return nil
}

func getOwner(ctx iscp.SandboxView) dict.Dict {
	return Result(ctx.State().MustGet(keyEVMOwner))
}

func setGasRatio(ctx iscp.Sandbox) dict.Dict {
	requireOwner(ctx)
	ctx.State().Set(keyGasRatio, codec.MustDecodeRatio32(ctx.Params().MustGet(evm.FieldGasRatio)).Bytes())
	return nil
}

func getGasRatio(ctx iscp.SandboxView) dict.Dict {
	return Result(ctx.State().MustGet(keyGasRatio))
}

func ApplyTransaction(ctx iscp.Sandbox, apply func(tx *types.Transaction, blockTime uint32, gasBudget uint64) (uint64, error)) dict.Dict {
	a := assert.NewAssert(ctx.Log())

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(evm.FieldTransactionData))
	a.RequireNoError(err)

	blockTime := getBlockTime(ctx.State())

	gasRatio := codec.MustDecodeRatio32(ctx.State().MustGet(keyGasRatio), evm.DefaultGasRatio)

	// ignore the evm gas budget set in the evm tx, use the remaining ISC gas budget instead
	gasBudget := ISCGasBudgetToEVM(ctx.Gas().Budget(), gasRatio)

	gasUsed, err := apply(tx, blockTime, gasBudget)

	// burn gas even on error
	ctx.Gas().Burn(gas.BurnCodeEVM1P, EVMGasToISC(gasUsed, gasRatio))

	ctx.RequireNoError(err)

	return nil
}

func ISCGasBudgetToEVM(iscGasBudget uint64, gasRatio util.Ratio32) uint64 {
	// EVM gas budget = floor(ISC gas budget * B / A)
	return gasRatio.YFloor64(iscGasBudget)
}

func EVMGasToISC(evmGas uint64, gasRatio util.Ratio32) uint64 {
	// ISC gas burned = ceil(EVM gas * A / B)
	return gasRatio.XCeil64(evmGas)
}
