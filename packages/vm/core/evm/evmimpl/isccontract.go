// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm/isccontract"
)

// deployISCContractOnGenesis sets up the initial state of the ISC EVM contract
// which will go into the EVM genesis block
func deployISCContractOnGenesis(genesisAlloc core.GenesisAlloc) {
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
	iscABI, err = abi.JSON(strings.NewReader(isccontract.ABI))
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
		panic(fmt.Sprintf("ISCContract: method not found: %x", input[:4]))
	}
	args, err := method.Inputs.Unpack(input[4:])
	if err != nil {
		panic(err)
	}
	return method, args
}

type iscContract struct {
	ctx iscp.Sandbox
}

func newISCContract(ctx iscp.Sandbox) vm.ISCContract {
	return &iscContract{ctx}
}

//nolint:funlen
func (c *iscContract) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64) {
	ret, remainingGas, _, ok := tryBaseCall(c.ctx, evm, caller, input, gas, readOnly)
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
		outs = []interface{}{isccontract.WrapISCAgentID(c.ctx.Request().SenderAccount())}

	case "getAllowanceIotas":
		outs = []interface{}{c.ctx.Request().Allowance().Assets.Iotas}

	case "getAllowanceNativeTokensLen":
		outs = []interface{}{uint16(len(c.ctx.Request().Allowance().Assets.Tokens))}

	case "getAllowanceNativeToken":
		i := args[0].(uint16)
		outs = []interface{}{isccontract.WrapIotaNativeToken(c.ctx.Request().Allowance().Assets.Tokens[i])}

	case "getAllowanceNFTsLen":
		outs = []interface{}{uint16(len(c.ctx.Request().Allowance().NFTs))}

	case "getAllowanceNFTID":
		i := args[0].(uint16)
		outs = []interface{}{isccontract.WrapIotaNFTID(c.ctx.Request().Allowance().NFTs[i])}

	case "getAllowanceNFT":
		i := args[0].(uint16)
		nftID := isccontract.WrapIotaNFTID(c.ctx.Request().Allowance().NFTs[i])
		nft := c.ctx.GetNFTData(nftID.Unwrap())
		outs = []interface{}{isccontract.WrapISCNFT(&nft)}

	case "getCaller":
		outs = []interface{}{isccontract.WrapISCAgentID(c.ctx.Caller())}

	case "registerError":
		errorMessage := args[0].(string)
		outs = []interface{}{c.ctx.RegisterError(errorMessage).Create().Code().ID}

	case "send":
		params := isccontract.ISCRequestParameters{}
		err := method.Inputs.Copy(&params, args)
		c.ctx.RequireNoError(err)
		c.ctx.Send(params.Unwrap())

	case "sendAsNFT":
		var callArgs struct {
			isccontract.ISCRequestParameters
			ID isccontract.IotaNFTID
		}
		err := method.Inputs.Copy(&callArgs, args)
		c.ctx.RequireNoError(err)
		c.ctx.TransferAllowedFunds(c.ctx.AccountID())
		c.ctx.SendAsNFT(callArgs.Unwrap(), callArgs.ID.Unwrap())

	case "call":
		var callArgs struct {
			ContractHname uint32
			EntryPoint    uint32
			Params        isccontract.ISCDict
			Allowance     isccontract.ISCAllowance
		}
		err := method.Inputs.Copy(&callArgs, args)
		c.ctx.RequireNoError(err)
		callRet := c.ctx.Call(
			iscp.Hname(callArgs.ContractHname),
			iscp.Hname(callArgs.EntryPoint),
			callArgs.Params.Unwrap(),
			callArgs.Allowance.Unwrap(),
		)
		outs = []interface{}{isccontract.WrapISCDict(callRet)}

	case "getAllowanceAvailableIotas":
		outs = []interface{}{c.ctx.AllowanceAvailable().Assets.Iotas}

	case "getAllowanceAvailableNativeToken":
		i := args[0].(uint16)
		outs = []interface{}{isccontract.WrapIotaNativeToken(c.ctx.AllowanceAvailable().Assets.Tokens[i])}

	case "getAllowanceAvailableNativeTokensLen":
		outs = []interface{}{uint16(len(c.ctx.AllowanceAvailable().Assets.Tokens))}

	case "getAllowanceAvailableNFTsLen":
		outs = []interface{}{uint16(len(c.ctx.AllowanceAvailable().NFTs))}

	case "getAllowanceAvailableNFT":
		i := args[0].(uint16)
		nftID := isccontract.WrapIotaNFTID(c.ctx.AllowanceAvailable().NFTs[i])
		nft := c.ctx.GetNFTData(nftID.Unwrap())
		outs = []interface{}{isccontract.WrapISCNFT(&nft)}

	default:
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	c.ctx.RequireNoError(err)
	return ret, remainingGas
}

type iscContractView struct {
	ctx iscp.SandboxView
}

func newISCContractView(ctx iscp.SandboxView) vm.ISCContract {
	return &iscContractView{ctx}
}

var _ vm.ISCContract = &iscContractView{}

func (c *iscContractView) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64) {
	ret, remainingGas, _, ok := tryBaseCall(c.ctx, evm, caller, input, gas, readOnly)
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
			Params        isccontract.ISCDict
		}
		err := method.Inputs.Copy(&callViewArgs, args)
		c.ctx.RequireNoError(err)
		callRet := c.ctx.CallView(
			iscp.Hname(callViewArgs.ContractHname),
			iscp.Hname(callViewArgs.EntryPoint),
			callViewArgs.Params.Unwrap(),
		)
		outs = []interface{}{isccontract.WrapISCDict(callRet)}

	default:
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	c.ctx.RequireNoError(err)
	return ret, remainingGas
}

// TODO evm param is not used, can it be removed?

func tryBaseCall(ctx iscp.SandboxBase, evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, method *abi.Method, ok bool) {
	remainingGas = gas
	method, args := parseCall(input)
	var outs []interface{}

	switch method.Name {
	case "hn":
		outs = []interface{}{iscp.Hn(args[0].(string))}

	case "hasParam":
		outs = []interface{}{ctx.Params().MustHas(kv.Key(args[0].(string)))}

	case "getParam":
		outs = []interface{}{ctx.Params().MustGet(kv.Key(args[0].(string)))}

	case "getChainID":
		outs = []interface{}{isccontract.WrapISCChainID(ctx.ChainID())}

	case "getChainOwnerID":
		outs = []interface{}{isccontract.WrapISCAgentID(ctx.ChainOwnerID())}

	case "getNFTData":
		var nftID isccontract.IotaNFTID
		err := method.Inputs.Copy(&nftID, args)
		ctx.RequireNoError(err)
		nft := ctx.GetNFTData(nftID.Unwrap())
		outs = []interface{}{isccontract.WrapISCNFT(&nft)}

	case "getTimestampUnixSeconds":
		outs = []interface{}{ctx.Timestamp().Unix()}

	case "logInfo":
		ctx.Log().Infof("%s", args[0].(string))

	case "logDebug":
		ctx.Log().Debugf("%s", args[0].(string))

	case "logPanic":
		ctx.Log().Panicf("%s", args[0].(string))

	default:
		return
	}

	ok = true
	ret, err := method.Outputs.Pack(outs...)
	ctx.RequireNoError(err)
	return ret, remainingGas, method, ok
}
