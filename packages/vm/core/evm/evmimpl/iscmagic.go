// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

var allMethods = make(map[string]*magicMethod)

type magicMethod struct {
	isPrivileged bool
	abi          *abi.Method
}

func init() {
	for _, iface := range []struct {
		abi          string
		isPrivileged bool
	}{
		{abi: iscmagic.SandboxABI, isPrivileged: false},
		{abi: iscmagic.AccountsABI, isPrivileged: false},
		{abi: iscmagic.UtilABI, isPrivileged: false},
		{abi: iscmagic.PrivilegedABI, isPrivileged: true},
	} {
		parsedABI, err := abi.JSON(strings.NewReader(iface.abi))
		if err != nil {
			panic(err)
		}
		for k := range parsedABI.Methods {
			method := parsedABI.Methods[k]
			allMethods[string(method.ID)] = &magicMethod{
				isPrivileged: iface.isPrivileged,
				abi:          &method,
			}
		}
	}
}

func parseCall(input []byte, privileged bool) (*abi.Method, []any) {
	magicMethod := allMethods[string(input[:4])]
	if magicMethod == nil {
		panic("method not found")
	}
	if !privileged && magicMethod.isPrivileged {
		panic("unauthorized")
	}
	if len(input) == 4 {
		return magicMethod.abi, nil
	}
	args, err := magicMethod.abi.Inputs.Unpack(input[4:])
	if err != nil {
		panic(err)
	}
	return magicMethod.abi, args
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
		log.Debugf("EVM request failed with ISC panic, caller: %s, input: %s,err: %v", caller.Address(), iotago.EncodeHex(input), executionErr)
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

type magicContract struct {
	ctx isc.Sandbox
}

func newMagicContract(ctx isc.Sandbox) map[common.Address]vm.ISCMagicContract {
	return map[common.Address]vm.ISCMagicContract{
		iscmagic.Address: &magicContract{ctx},
	}
}

func (c *magicContract) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	return catchISCPanics(c.doRun, evm, caller, input, gas, readOnly, c.ctx.Log())
}

func (c *magicContract) doRun(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) []byte {
	privileged := isCallerPrivileged(c.ctx, caller.Address())
	method, args := parseCall(input, privileged)
	if method.IsConstant() {
		return callViewHandler(c.ctx, caller, method, args)
	}
	if readOnly {
		panic("attempt to call non-view method in read-only context")
	}
	return callHandler(c.ctx, caller, method, args)
}

type magicContractView struct {
	ctx isc.SandboxBase
}

func newMagicContractView(ctx isc.SandboxBase) map[common.Address]vm.ISCMagicContract {
	return map[common.Address]vm.ISCMagicContract{
		iscmagic.Address: &magicContractView{ctx},
	}
}

func (c *magicContractView) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	return catchISCPanics(c.doRun, evm, caller, input, gas, readOnly, c.ctx.Log())
}

func (c *magicContractView) doRun(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) []byte {
	privileged := isCallerPrivileged(c.ctx, caller.Address())
	method, args := parseCall(input, privileged)
	return callViewHandler(c.ctx, caller, method, args)
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
