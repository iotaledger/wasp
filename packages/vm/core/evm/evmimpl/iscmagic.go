// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	iscvm "github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

const (
	prefixPrivileged = "p"
	prefixAllowance  = "a"
)

// directory of EVM contracts that have access to the privileged methods of ISC magic
func keyPrivileged(addr common.Address) kv.Key {
	return kv.Key(prefixPrivileged) + kv.Key(addr.Bytes())
}

func isPrivileged(ctx isc.SandboxBase, addr common.Address) bool {
	state := iscMagicSubrealmR(ctx.StateR())
	return state.MustHas(keyPrivileged(addr))
}

func addToPrivileged(ctx isc.Sandbox, addr common.Address) {
	state := iscMagicSubrealm(ctx.State())
	state.Set(keyPrivileged(addr), []byte{1})
}

// allowance between two EVM accounts
func keyAllowance(from, to common.Address) kv.Key {
	return kv.Key(prefixAllowance) + kv.Key(from.Bytes()) + kv.Key(to.Bytes())
}

func getAllowance(ctx isc.SandboxBase, from, to common.Address) *isc.Allowance {
	state := iscMagicSubrealmR(ctx.StateR())
	key := keyAllowance(from, to)
	return codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
}

func addToAllowance(ctx isc.Sandbox, from, to common.Address, add *isc.Allowance) {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)
	allowance := codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
	allowance.Add(add)
	state.Set(key, allowance.Bytes())
}

func subtractFromAllowance(ctx isc.Sandbox, from, to common.Address, taken *isc.Allowance) *isc.Allowance {
	state := iscMagicSubrealm(ctx.State())
	key := keyAllowance(from, to)

	remaining := codec.MustDecodeAllowance(state.MustGet(key), isc.NewEmptyAllowance())
	if taken.IsEmpty() {
		taken = remaining.Clone()
	}

	ok := remaining.SpendFromBudget(taken)
	ctx.Requiref(ok, "takeAllowedFunds: not previously allowed")
	if remaining.IsEmpty() {
		state.Del(key)
	} else {
		state.Set(key, remaining.Bytes())
	}

	return taken
}

// deployMagicContractOnGenesis sets up the initial state of the ISC EVM contract
// which will go into the EVM genesis block
func deployMagicContractOnGenesis(genesisAlloc core.GenesisAlloc) {
	genesisAlloc[iscmagic.Address] = core.GenesisAccount{
		// dummy code, because some contracts check the code size before calling
		// the contract; the code itself will never get executed
		Code:    common.Hex2Bytes("600180808053f3"),
		Storage: map[common.Hash]common.Hash{},
		Balance: &big.Int{},
	}
}

var (
	iscABI           abi.ABI
	iscPrivilegedABI abi.ABI
)

func init() {
	var err error
	iscABI, err = abi.JSON(strings.NewReader(iscmagic.ABI))
	if err != nil {
		panic(err)
	}
	iscPrivilegedABI, err = abi.JSON(strings.NewReader(iscmagic.PrivilegedABI))
	if err != nil {
		panic(err)
	}
}

func parseCall(input []byte, privileged bool) (*abi.Method, []interface{}) {
	var method *abi.Method
	if privileged {
		method, _ = iscPrivilegedABI.MethodById(input[:4])
	}
	if method == nil {
		var err error
		method, err = iscABI.MethodById(input[:4])
		if err != nil {
			panic(err)
		}
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

func newMagicContract(ctx isc.Sandbox) map[common.Address]vm.ISCMagicContract {
	return map[common.Address]vm.ISCMagicContract{
		iscmagic.Address: &magicContract{ctx},
	}
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

type RunFunc func(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) []byte

// see UnpackRevert in go-ethereum/accounts/abi/abi.go
var revertSelector = crypto.Keccak256([]byte("Error(string)"))[:4]

// catchISCPanics executes a `Run` function (either from a call or view), and catches ISC exceptions, if any ISC exception happens, ErrExecutionReverted is issued
func catchISCPanics(run RunFunc, evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool, log isc.LogInterface) (ret []byte, remainingGas uint64, executionErr error) {
	executionErr = panicutil.CatchAllExcept(
		func() {
			ret = run(evm, caller, input, gas, readOnly)
		},
		vmexceptions.AllProtocolLimits...,
	)
	if executionErr != nil {
		log.Infof("EVM request failed with ISC panic, caller: %s, input: %s,err: %v", caller.Address(), hexutil.Encode(input), executionErr)
		// TODO this works, but is there a better way to encode the error in the required abi format?

		// include the ISC error as the revert reason by encoding it into the returnData
		ret = revertSelector
		abiString, err := abi.NewType("string", "", nil)
		if err != nil {
			panic(err)
		}
		encodedErr, err := abi.Arguments{{Type: abiString}}.Pack(executionErr.Error())
		if err != nil {
			panic(err)
		}
		ret = append(ret, encodedErr...)
		// override the error to be returned (must be "execution reverted")
		executionErr = vm.ErrExecutionReverted
	}
	return ret, gas, executionErr
}

func (c *magicContract) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	return catchISCPanics(c.doRun, evm, caller, input, gas, readOnly, c.ctx.Log())
}

func (c *magicContract) doRun(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) []byte {
	privileged := isPrivileged(c.ctx, caller.Address())
	method, args := parseCall(input, privileged)

	var outs []interface{}
	var ok bool
	if privileged {
		outs, ok = tryPrivilegedCall(c.ctx, caller, method, args)
	}
	if !ok {
		outs, ok = tryUnprivilegedCall(c.ctx, caller, method, args)
	}
	if !ok {
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	c.ctx.RequireNoError(err)
	return ret
}

func tryUnprivilegedCall(ctx isc.Sandbox, caller vm.ContractRef, method *abi.Method, args []interface{}) ([]interface{}, bool) {
	if outs, ok := tryViewCall(ctx, caller, method, args); ok {
		return outs, ok
	}
	if outs, ok := tryCall(ctx, caller, method, args); ok {
		return outs, ok
	}
	return nil, false
}

//nolint:funlen
func tryCall(ctx isc.Sandbox, caller vm.ContractRef, method *abi.Method, args []interface{}) ([]interface{}, bool) {
	switch method.Name {
	case "getEntropy":
		return []interface{}{ctx.GetEntropy()}, true

	case "triggerEvent":
		ctx.Event(args[0].(string))
		return nil, true

	case "getRequestID":
		return []interface{}{ctx.Request().ID()}, true

	case "getSenderAccount":
		return []interface{}{iscmagic.WrapISCAgentID(ctx.Request().SenderAccount())}, true

	case "allow":
		params := struct {
			Target    common.Address
			Allowance iscmagic.ISCAllowance
		}{}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)
		addToAllowance(ctx, caller.Address(), params.Target, params.Allowance.Unwrap())
		return nil, true

	case "takeAllowedFunds":
		params := struct {
			Addr      common.Address
			Allowance iscmagic.ISCAllowance
		}{}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)

		taken := subtractFromAllowance(ctx, params.Addr, caller.Address(), params.Allowance.Unwrap())
		ctx.Privileged().MustMoveBetweenAccounts(
			isc.NewEthereumAddressAgentID(params.Addr),
			isc.NewEthereumAddressAgentID(caller.Address()),
			taken.Assets,
			taken.NFTs,
		)
		return nil, true

	case "send":
		params := struct {
			TargetAddress               iscmagic.L1Address
			FungibleTokens              iscmagic.ISCFungibleTokens
			AdjustMinimumStorageDeposit bool
			Metadata                    iscmagic.ISCSendMetadata
			SendOptions                 iscmagic.ISCSendOptions
		}{}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)
		req := isc.RequestParameters{
			TargetAddress:                 params.TargetAddress.MustUnwrap(),
			FungibleTokens:                params.FungibleTokens.Unwrap(),
			AdjustToMinimumStorageDeposit: params.AdjustMinimumStorageDeposit,
			Metadata:                      params.Metadata.Unwrap(),
			Options:                       params.SendOptions.Unwrap(),
		}
		adjustStorageDeposit(ctx, req)

		moveAssetsToCommonAccount(ctx, caller, req.FungibleTokens, nil)

		// assert that remaining tokens in the sender's account are enough to pay for the gas budget
		if !ctx.HasInAccount(
			ctx.Request().SenderAccount(),
			ctx.Privileged().TotalGasTokens(),
		) {
			panic(iscvm.ErrNotEnoughTokensLeftForGas)
		}
		ctx.Send(req)
		return nil, true

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
		ctx.RequireNoError(err)
		req := isc.RequestParameters{
			TargetAddress:                 params.TargetAddress.MustUnwrap(),
			FungibleTokens:                params.FungibleTokens.Unwrap(),
			AdjustToMinimumStorageDeposit: params.AdjustMinimumStorageDeposit,
			Metadata:                      params.Metadata.Unwrap(),
			Options:                       params.SendOptions.Unwrap(),
		}
		nftID := params.NFTID.Unwrap()
		adjustStorageDeposit(ctx, req)

		// make sure that allowance <= sent tokens, so that the target contract does not
		// spend from the common account
		ctx.Requiref(
			isc.NewAllowanceFungibleTokens(req.FungibleTokens).AddNFTs(nftID).SpendFromBudget(req.Metadata.Allowance),
			"allowance must not be greater than sent tokens",
		)

		moveAssetsToCommonAccount(ctx, caller, req.FungibleTokens, []iotago.NFTID{nftID})

		// assert that remaining tokens in the sender's account are enough to pay for the gas budget
		if !ctx.HasInAccount(
			ctx.Request().SenderAccount(),
			ctx.Privileged().TotalGasTokens(),
		) {
			panic(iscvm.ErrNotEnoughTokensLeftForGas)
		}
		ctx.SendAsNFT(req, nftID)
		return nil, true

	case "call":
		var callArgs struct {
			ContractHname uint32
			EntryPoint    uint32
			Params        iscmagic.ISCDict
			Allowance     iscmagic.ISCAllowance
		}
		err := method.Inputs.Copy(&callArgs, args)
		ctx.RequireNoError(err)
		allowance := callArgs.Allowance.Unwrap()
		moveAssetsToCommonAccount(ctx, caller, allowance.Assets, allowance.NFTs)
		callRet := ctx.Call(
			isc.Hname(callArgs.ContractHname),
			isc.Hname(callArgs.EntryPoint),
			callArgs.Params.Unwrap(),
			allowance,
		)
		return []interface{}{iscmagic.WrapISCDict(callRet)}, true
	}
	return nil, false
}

//nolint:unparam
func tryPrivilegedCall(ctx isc.Sandbox, caller vm.ContractRef, method *abi.Method, args []interface{}) ([]interface{}, bool) {
	if !isPrivileged(ctx, caller.Address()) {
		return nil, false
	}
	switch method.Name {
	case "moveBetweenAccounts":
		var params struct {
			Sender    common.Address
			Receiver  common.Address
			Allowance iscmagic.ISCAllowance
		}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)
		allowance := params.Allowance.Unwrap()
		ctx.Privileged().MustMoveBetweenAccounts(
			isc.NewEthereumAddressAgentID(params.Sender),
			isc.NewEthereumAddressAgentID(params.Receiver),
			allowance.Assets,
			allowance.NFTs,
		)
		return nil, true

	case "addToAllowance":
		var params struct {
			From      common.Address
			To        common.Address
			Allowance iscmagic.ISCAllowance
		}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)
		addToAllowance(ctx, params.From, params.To, params.Allowance.Unwrap())
		return nil, true

	case "moveAllowedFunds":
		var params struct {
			From      common.Address
			To        common.Address
			Allowance iscmagic.ISCAllowance
		}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)
		taken := subtractFromAllowance(ctx, params.From, params.To, params.Allowance.Unwrap())
		ctx.Privileged().MustMoveBetweenAccounts(
			isc.NewEthereumAddressAgentID(params.From),
			isc.NewEthereumAddressAgentID(params.To),
			taken.Assets,
			taken.NFTs,
		)
		return nil, true
	}
	return nil, false
}

type magicContractView struct {
	ctx isc.SandboxView
}

func newMagicContractView(ctx isc.SandboxView) map[common.Address]vm.ISCMagicContract {
	return map[common.Address]vm.ISCMagicContract{
		iscmagic.Address: &magicContractView{ctx},
	}
}

func (c *magicContractView) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	return catchISCPanics(c.doRun, evm, caller, input, gas, readOnly, c.ctx.Log())
}

func (c *magicContractView) doRun(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) []byte {
	method, args := parseCall(input, isPrivileged(c.ctx, caller.Address()))

	outs, ok := tryViewCall(c.ctx, caller, method, args)
	if !ok {
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	c.ctx.RequireNoError(err)
	return ret
}

func tryViewCall(ctx isc.SandboxBase, caller vm.ContractRef, method *abi.Method, args []interface{}) (outs []interface{}, ok bool) {
	switch method.Name {
	case "hn":
		return []interface{}{isc.Hn(args[0].(string))}, true

	case "getChainID":
		return []interface{}{iscmagic.WrapISCChainID(ctx.ChainID())}, true

	case "getChainOwnerID":
		return []interface{}{iscmagic.WrapISCAgentID(ctx.ChainOwnerID())}, true

	case "getNFTData":
		var nftID iscmagic.NFTID
		err := method.Inputs.Copy(&nftID, args)
		ctx.RequireNoError(err)
		nft := ctx.GetNFTData(nftID.Unwrap())
		return []interface{}{iscmagic.WrapISCNFT(&nft)}, true

	case "getTimestampUnixSeconds":
		return []interface{}{ctx.Timestamp().Unix()}, true

	case "callView":
		var callViewArgs struct {
			ContractHname uint32
			EntryPoint    uint32
			Params        iscmagic.ISCDict
		}
		err := method.Inputs.Copy(&callViewArgs, args)
		ctx.RequireNoError(err)
		callRet := ctx.CallView(
			isc.Hname(callViewArgs.ContractHname),
			isc.Hname(callViewArgs.EntryPoint),
			callViewArgs.Params.Unwrap(),
		)
		return []interface{}{iscmagic.WrapISCDict(callRet)}, true

	case "getAllowanceFrom":
		var addr common.Address
		err := method.Inputs.Copy(&addr, args)
		ctx.RequireNoError(err)
		return []interface{}{iscmagic.WrapISCAllowance(getAllowance(ctx, addr, caller.Address()))}, true

	case "getAllowanceTo":
		var target common.Address
		err := method.Inputs.Copy(&target, args)
		ctx.RequireNoError(err)
		return []interface{}{iscmagic.WrapISCAllowance(getAllowance(ctx, caller.Address(), target))}, true

	case "getAllowance":
		var params struct {
			From common.Address
			To   common.Address
		}
		err := method.Inputs.Copy(&params, args)
		ctx.RequireNoError(err)
		return []interface{}{iscmagic.WrapISCAllowance(getAllowance(ctx, params.From, params.To))}, true

	case "getBaseTokenProperties":
		l1 := parameters.L1()
		return []interface{}{iscmagic.ISCTokenProperties{
			Name:         l1.BaseToken.Name,
			TickerSymbol: l1.BaseToken.TickerSymbol,
			Decimals:     uint8(l1.BaseToken.Decimals),
			TotalSupply:  big.NewInt(int64(l1.Protocol.TokenSupply)),
		}}, true

	case "print":
		var s string
		err := method.Inputs.Copy(&s, args)
		ctx.RequireNoError(err)
		ctx.Log().Debugf("isc.print() -> %q", s)
		return nil, true
	}
	return nil, false
}
