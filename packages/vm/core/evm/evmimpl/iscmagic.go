// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

func keyAllowance(from, to common.Address) kv.Key {
	return kv.Key("a") + kv.Key(from.Bytes()) + kv.Key(to.Bytes())
}

// deployMagicContractOnGenesis sets up the initial state of the ISC EVM contract
// which will go into the EVM genesis block
func deployMagicContractOnGenesis(genesisAlloc core.GenesisAlloc) {
	genesisAlloc[vm.ISCAddress] = core.GenesisAccount{
		// dummy code, because some contracts check the code size before calling
		// the contract; the code itself will never get executed
		Code:    common.Hex2Bytes("600180808053f3"),
		Storage: map[common.Hash]common.Hash{},
		Balance: &big.Int{},
	}
}

var iscABI abi.ABI

func init() {
	var err error
	iscABI, err = abi.JSON(strings.NewReader(iscmagic.ABI))
	if err != nil {
		panic(err)
	}
}

func parseCall(input []byte) (*abi.Method, []interface{}) {
	method, err := iscABI.MethodById(input[:4])
	if err != nil {
		panic(err)
	}
	if method == nil {
		panic(fmt.Sprintf("iscmagic: method not found: %x", input[:4]))
	}
	args, err := method.Inputs.Unpack(input[4:])
	if err != nil {
		panic(err)
	}
	return method, args
}

type magicContract struct {
	ctx isc.Sandbox
}

func newMagicContract(ctx isc.Sandbox) vm.ISCContract {
	return &magicContract{ctx}
}

func adjustStorageDeposit(ctx isc.Sandbox, req isc.RequestParameters) {
	sd := ctx.EstimateRequiredStorageDeposit(req)
	if req.FungibleTokens.BaseTokens < sd {
		if !req.AdjustToMinimumStorageDeposit {
			panic(fmt.Sprintf(
				"base tokens (%d) not enough to cover storage deposit (%d)",
				req.FungibleTokens.BaseTokens,
				sd,
			))
		}
		req.FungibleTokens.BaseTokens = sd
	}
}

// moveAssetsToCommonAccount moves the assets from the caller's L2 account to the common
// account before sending to L1
func moveAssetsToCommonAccount(ctx isc.Sandbox, caller vm.ContractRef, fungibleTokens *isc.FungibleTokens, nftIDs []iotago.NFTID) {
	ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(caller.Address()),
		ctx.AccountID(),
		fungibleTokens,
		nftIDs,
	)
}

type RunFunc func(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64)

// catchISCPanics executes a `Run` function (either from a call or view), and catches ISC exceptions, if any ISC exception happens, ErrExecutionReverted is issued
func catchISCPanics(run RunFunc, evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool, log isc.LogInterface) (ret []byte, remainingGas uint64, err error) {
	err = panicutil.CatchAllExcept(
		func() {
			ret, remainingGas = run(evm, caller, input, gas, readOnly)
		},
		vmexceptions.AllProtocolLimits...,
	)
	if err != nil {
		remainingGas = gas
		log.Infof("EVM request failed with ISC panic, caller: %s, input: %s,err: %v", caller.Address(), hex.EncodeToString(input), err)
		// the ISC error is lost inside the EVM, a possible solution would be to wrap the ErrExecutionReverted error, but the ISC information still gets deleted at some point
		// err = errors.Wrap(vm.ErrExecutionReverted, err.Error())
		err = vm.ErrExecutionReverted
	}
	return ret, remainingGas, err
}

func (c *magicContract) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	return catchISCPanics(c.doRun, evm, caller, input, gas, readOnly, c.ctx.Log())
}

//nolint:funlen
func (c *magicContract) doRun(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64) {
	c.ctx.Privileged().GasBurnEnable(true)
	defer c.ctx.Privileged().GasBurnEnable(false)

	ret, remainingGas, _, ok := tryBaseCall(c.ctx, input, gas)
	if ok {
		return ret, remainingGas
	}

	remainingGas = gas
	method, args := parseCall(input)
	var outs []interface{}

	switch method.Name {
	case "getEntropy":
		outs = []interface{}{c.ctx.GetEntropy()}

	case "triggerEvent":
		c.ctx.Event(args[0].(string))

	case "getRequestID":
		outs = []interface{}{c.ctx.Request().ID()}

	case "getSenderAccount":
		outs = []interface{}{iscmagic.WrapISCAgentID(c.ctx.Request().SenderAccount())}

	case "registerError":
		errorMessage := args[0].(string)
		outs = []interface{}{c.ctx.RegisterError(errorMessage).Create().Code().ID}

	case "allow":
		params := struct {
			Target    common.Address
			Allowance iscmagic.ISCAllowance
		}{}
		err := method.Inputs.Copy(&params, args)
		c.ctx.RequireNoError(err)

		state := iscMagicSubrealm(c.ctx.State())
		key := keyAllowance(caller.Address(), params.Target)
		allowance := codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
		allowance.Add(params.Allowance.Unwrap())
		state.Set(key, allowance.Bytes())

	case "takeAllowedFunds":
		params := struct {
			Addr      common.Address
			Allowance iscmagic.ISCAllowance
		}{}
		err := method.Inputs.Copy(&params, args)
		c.ctx.RequireNoError(err)

		state := iscMagicSubrealm(c.ctx.State())
		key := keyAllowance(params.Addr, caller.Address())
		taken := params.Allowance.Unwrap()
		allowance := codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
		ok := allowance.SpendFromBudget(taken)
		c.ctx.Requiref(ok, "takeAllowedFunds: not previously allowed")
		if allowance.IsEmpty() {
			state.Del(key)
		} else {
			state.Set(key, allowance.Bytes())
		}

		c.ctx.Privileged().MustMoveBetweenAccounts(
			isc.NewEthereumAddressAgentID(params.Addr),
			isc.NewEthereumAddressAgentID(caller.Address()),
			taken.Assets,
			taken.NFTs,
		)

	case "send":
		params := struct {
			TargetAddress               iscmagic.L1Address
			FungibleTokens              iscmagic.ISCFungibleTokens
			AdjustMinimumStorageDeposit bool
			Metadata                    iscmagic.ISCSendMetadata
			SendOptions                 iscmagic.ISCSendOptions
		}{}
		err := method.Inputs.Copy(&params, args)
		c.ctx.RequireNoError(err)
		req := isc.RequestParameters{
			TargetAddress:                 params.TargetAddress.MustUnwrap(),
			FungibleTokens:                params.FungibleTokens.Unwrap(),
			AdjustToMinimumStorageDeposit: params.AdjustMinimumStorageDeposit,
			Metadata:                      params.Metadata.Unwrap(),
			Options:                       params.SendOptions.Unwrap(),
		}
		adjustStorageDeposit(c.ctx, req)

		// make sure that allowance <= sent tokens, so that the target contract does not
		// spend from the common account
		c.ctx.Requiref(
			isc.NewAllowanceFungibleTokens(req.FungibleTokens).SpendFromBudget(req.Metadata.Allowance),
			"allowance must not be greater than sent tokens",
		)

		moveAssetsToCommonAccount(c.ctx, caller, req.FungibleTokens, nil)
		c.ctx.Send(req)

	case "sendAsNFT":
		params := struct {
			TargetAddress               iscmagic.L1Address
			FungibleTokens              iscmagic.ISCFungibleTokens
			NFTID                       iscmagic.NFTID
			AdjustMinimumStorageDeposit bool
			Metadata                    iscmagic.ISCSendMetadata
			SendOptions                 iscmagic.ISCSendOptions
		}{}
		err := method.Inputs.Copy(&params, args)
		c.ctx.RequireNoError(err)
		req := isc.RequestParameters{
			TargetAddress:                 params.TargetAddress.MustUnwrap(),
			FungibleTokens:                params.FungibleTokens.Unwrap(),
			AdjustToMinimumStorageDeposit: params.AdjustMinimumStorageDeposit,
			Metadata:                      params.Metadata.Unwrap(),
			Options:                       params.SendOptions.Unwrap(),
		}
		nftID := params.NFTID.Unwrap()
		adjustStorageDeposit(c.ctx, req)

		// make sure that allowance <= sent tokens, so that the target contract does not
		// spend from the common account
		c.ctx.Requiref(
			isc.NewAllowanceFungibleTokens(req.FungibleTokens).AddNFTs(nftID).SpendFromBudget(req.Metadata.Allowance),
			"allowance must not be greater than sent tokens",
		)

		moveAssetsToCommonAccount(c.ctx, caller, req.FungibleTokens, []iotago.NFTID{nftID})
		c.ctx.SendAsNFT(req, nftID)

	case "call":
		var callArgs struct {
			ContractHname uint32
			EntryPoint    uint32
			Params        iscmagic.ISCDict
			Allowance     iscmagic.ISCAllowance
		}
		err := method.Inputs.Copy(&callArgs, args)
		c.ctx.RequireNoError(err)
		allowance := callArgs.Allowance.Unwrap()
		moveAssetsToCommonAccount(c.ctx, caller, allowance.Assets, allowance.NFTs)
		callRet := c.ctx.Call(
			isc.Hname(callArgs.ContractHname),
			isc.Hname(callArgs.EntryPoint),
			callArgs.Params.Unwrap(),
			allowance,
		)
		outs = []interface{}{iscmagic.WrapISCDict(callRet)}

	default:
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	c.ctx.RequireNoError(err)
	return ret, remainingGas
}

type magicContractView struct {
	ctx isc.SandboxView
}

func newMagicContractView(ctx isc.SandboxView) vm.ISCContract {
	return &magicContractView{ctx}
}

func (c *magicContractView) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	return catchISCPanics(c.doRun, evm, caller, input, gas, readOnly, c.ctx.Log())
}

func (c *magicContractView) doRun(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64) {
	c.ctx.Privileged().GasBurnEnable(true)
	defer c.ctx.Privileged().GasBurnEnable(false)

	ret, remainingGas, _, ok := tryBaseCall(c.ctx, input, gas)
	if ok {
		return ret, remainingGas
	}

	remainingGas = gas
	method, args := parseCall(input)
	var outs []interface{}

	switch method.Name {
	case "callView":
		var callViewArgs struct {
			ContractHname uint32
			EntryPoint    uint32
			Params        iscmagic.ISCDict
		}
		err := method.Inputs.Copy(&callViewArgs, args)
		c.ctx.RequireNoError(err)
		callRet := c.ctx.CallView(
			isc.Hname(callViewArgs.ContractHname),
			isc.Hname(callViewArgs.EntryPoint),
			callViewArgs.Params.Unwrap(),
		)
		outs = []interface{}{iscmagic.WrapISCDict(callRet)}

	default:
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	c.ctx.RequireNoError(err)
	return ret, remainingGas
}

func tryBaseCall(ctx isc.SandboxBase, input []byte, gas uint64) (ret []byte, remainingGas uint64, method *abi.Method, ok bool) {
	remainingGas = gas
	method, args := parseCall(input)
	var outs []interface{}

	switch method.Name {
	case "hn":
		outs = []interface{}{isc.Hn(args[0].(string))}

	case "getChainID":
		outs = []interface{}{iscmagic.WrapISCChainID(ctx.ChainID())}

	case "getChainOwnerID":
		outs = []interface{}{iscmagic.WrapISCAgentID(ctx.ChainOwnerID())}

	case "getNFTData":
		var nftID iscmagic.NFTID
		err := method.Inputs.Copy(&nftID, args)
		ctx.RequireNoError(err)
		nft := ctx.GetNFTData(nftID.Unwrap())
		outs = []interface{}{iscmagic.WrapISCNFT(&nft)}

	case "getTimestampUnixSeconds":
		outs = []interface{}{ctx.Timestamp().Unix()}

	default:
		return
	}

	ok = true
	ret, err := method.Outputs.Pack(outs...)
	ctx.RequireNoError(err)
	return ret, remainingGas, method, ok
}
