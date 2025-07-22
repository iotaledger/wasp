// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/samber/lo"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

// magicContractHandler has one public receiver for each ISC magic method, with
// the same name capitalized.
// For example, if ISC.getChainID() is called from solidity, this will
// correspond to a call to [GetChainID].
type magicContractHandler struct {
	ctx       isc.Sandbox
	evm       *vm.EVM
	caller    common.Address
	callValue *uint256.Int
}

// callHandler finds the requested ISC magic method by reflection, and executes
// it.
func callHandler(ctx isc.Sandbox, evm *vm.EVM, caller common.Address, callValue *uint256.Int, method *abi.Method, args []any) []byte {
	return reflectCall(&magicContractHandler{
		ctx:       ctx,
		evm:       evm,
		caller:    caller,
		callValue: callValue,
	}, method, args)
}

func titleCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func reflectCall(handler any, method *abi.Method, args []any) []byte {
	handlerMethod := reflect.ValueOf(handler).MethodByName(titleCase(method.Name))
	if !handlerMethod.IsValid() {
		panic(fmt.Sprintf("no handler for method %s", method.Name))
	}
	handlerMethodType := handlerMethod.Type()
	if handlerMethodType.NumIn() != len(args) {
		panic(fmt.Sprintf("%s: arguments length mismatch", method.Name))
	}

	callArgs := make([]reflect.Value, len(args))
	if len(args) > 0 {
		fields := make([]reflect.StructField, len(args))
		for i := range args {
			field := reflect.StructField{
				Name: titleCase(method.Inputs[i].Name),
				Type: handlerMethodType.In(i),
			}
			fields[i] = field
		}
		v := reflect.New(reflect.StructOf(fields)).Interface()
		err := method.Inputs.Copy(v, args)
		if err != nil {
			panic(err)
		}
		for i := range args {
			callArgs[i] = reflect.ValueOf(v).Elem().Field(i)
		}
	}
	results := handlerMethod.Call(callArgs)

	if len(results) == 0 {
		return nil
	}
	ret, err := method.Outputs.Pack(lo.Map(results, func(v reflect.Value, _ int) any {
		return v.Interface()
	})...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (h *magicContractHandler) call(msg isc.Message, allowance *isc.Assets) isc.CallArguments {
	return h.ctx.Privileged().CallOnBehalfOf(
		isc.NewEthereumAddressAgentID(h.caller),
		msg, allowance,
	)
}

func (h *magicContractHandler) callView(msg isc.Message) isc.CallArguments {
	return h.ctx.Privileged().CallOnBehalfOf(
		isc.NewEthereumAddressAgentID(h.caller),
		msg,
		isc.NewEmptyAssets(),
	)
}
