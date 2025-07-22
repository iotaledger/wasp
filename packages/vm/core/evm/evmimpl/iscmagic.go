// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	"github.com/iotaledger/wasp/v2/packages/isc"
	iscvm "github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/iscmagic"
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

var (
	errMethodNotFound        = coreerrors.Register("method not found").Create()
	errInvalidMethodArgs     = coreerrors.Register("invalid method arguments").Create()
	errReadOnlyContext       = coreerrors.Register("attempt to call non-view method in read-only context").Create()
	ErrPayingUnpayableMethod = coreerrors.Register("attempt to pay unpayable method %v")
)

// magicContract implements [vm.ISCMagicContract], which is an interface added
// to go-ethereum in ISC's fork. Whenever a call is made to the address
// [iscmagic.Address], [magicContract.Run] is called, allowing to process the call in
// native Go instead of solidity or EVM code.
type magicContract struct {
	ctx isc.Sandbox
}

func newMagicContract(ctx isc.Sandbox) map[common.Address]vm.ISCMagicContract {
	return map[common.Address]vm.ISCMagicContract{
		iscmagic.Address: &magicContract{ctx},
	}
}

func (c *magicContract) Run(evm *vm.EVM, caller common.Address, input []byte, value *uint256.Int, gas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	privileged := isCallerPrivileged(c.ctx, caller)
	method, args := parseCall(input, privileged)
	if readOnly && !method.IsConstant() {
		return nil, gas, errReadOnlyContext
	}

	c.ctx.Privileged().GasBurnEnable(true)
	defer c.ctx.Privileged().GasBurnEnable(false)

	// Reject value transactions calling non-payable methods.
	if value.BitLen() > 0 && !method.IsPayable() {
		return nil, gas, ErrPayingUnpayableMethod.Create(method.Name)
	}

	ret = callHandler(c.ctx, evm, caller, value, method, args)
	return ret, gas, nil
}

func parseCall(input []byte, privileged bool) (*abi.Method, []any) {
	magicMethod := allMethods[string(input[:4])]
	if magicMethod == nil {
		panic(errMethodNotFound)
	}
	if !privileged && magicMethod.isPrivileged {
		panic(iscvm.ErrUnauthorized)
	}
	if len(input) == 4 {
		return magicMethod.abi, nil
	}
	args, err := magicMethod.abi.Inputs.Unpack(input[4:])
	if err != nil {
		panic(errInvalidMethodArgs)
	}
	return magicMethod.abi, args
}
