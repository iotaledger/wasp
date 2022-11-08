// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/wasp/packages/isc"
)

type magicContractHandler struct {
	ctx    isc.Sandbox
	caller vm.ContractRef
}

func callHandler(ctx isc.Sandbox, caller vm.ContractRef, method *abi.Method, args []any) []byte {
	return reflectCall(&magicContractHandler{
		ctx:    ctx,
		caller: caller,
	}, method, args)
}

type magicContractViewHandler struct {
	ctx    isc.SandboxBase
	caller vm.ContractRef
}

func callViewHandler(ctx isc.SandboxBase, caller vm.ContractRef, method *abi.Method, args []any) []byte {
	return reflectCall(&magicContractViewHandler{
		ctx:    ctx,
		caller: caller,
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
		for i := 0; i < len(args); i++ {
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
		for i := 0; i < len(args); i++ {
			callArgs[i] = reflect.ValueOf(v).Elem().Field(i)
		}
	}
	results := handlerMethod.Call(callArgs)

	if len(results) == 0 {
		return nil
	}
	ret, err := method.Outputs.Pack(lo.Map(results, func(v reflect.Value) any {
		return v.Interface()
	})...)
	if err != nil {
		panic(err)
	}
	return ret
}
