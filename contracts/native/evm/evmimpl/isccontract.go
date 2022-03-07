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
	"github.com/iotaledger/wasp/contracts/native/evm/isccontract"
	"github.com/iotaledger/wasp/packages/iscp"
)

func init() {
	if iscp.ChainIDLength != 20 {
		panic("static check: ChainID length does not match bytes20 in ISC.sol")
	}
}

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

func (c *iscContract) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64) {
	remainingGas = gas
	method, args := parseCall(input)
	var outs []interface{}

	switch method.Name {
	case "getChainID":
		outs = getChainID(c.ctx, method)

	case "getEntropy":
		outs = []interface{}{c.ctx.GetEntropy()}

	case "triggerEvent":
		c.ctx.Event(args[0].(string))

	default:
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	if err != nil {
		panic(err)
	}
	return
}

type iscContractView struct {
	ctx iscp.SandboxView
}

func newISCContractView(ctx iscp.SandboxView) vm.ISCContract {
	return &iscContractView{ctx}
}

var _ vm.ISCContract = &iscContractView{}

func (c *iscContractView) Run(evm *vm.EVM, caller vm.ContractRef, input []byte, gas uint64, readOnly bool) (ret []byte, remainingGas uint64) {
	remainingGas = gas
	method, _ := parseCall(input)
	var outs []interface{}

	switch method.Name {
	case "getChainID":
		outs = getChainID(c.ctx, method)

	default:
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}

	ret, err := method.Outputs.Pack(outs...)
	if err != nil {
		panic(err)
	}
	return
}

func getChainID(ctx iscp.SandboxBase, method *abi.Method) (ret []interface{}) {
	return []interface{}{ctx.ChainID()}
}
