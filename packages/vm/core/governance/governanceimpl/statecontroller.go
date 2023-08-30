// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// rotateStateController the entry point is called when committee is about to be rotated to the new address
// If it fails, nothing happens and the state has trace of the failure in the state
// If it is successful VM takes over and replaces resulting transaction with
// governance transition. The state of the chain remains unchanged
func rotateStateController(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	newStateControllerAddr := ctx.Params().MustGetAddress(governance.ParamStateControllerAddress)
	// check is address is allowed
	state := ctx.State()
	amap := collections.NewMapReadOnly(state, governance.VarAllowedStateControllerAddresses)
	if !amap.HasAt(isc.AddressToBytes(newStateControllerAddr)) {
		panic(vm.ErrUnauthorized)
	}

	if !newStateControllerAddr.Equal(ctx.StateAnchor().StateController) {
		// rotate request to another address has been issued. State update will be taken over by VM and will have no effect
		// By setting VarRotateToAddress we signal the VM this special situation
		// VarRotateToAddress value should never persist in the state
		ctx.Log().Infof("Governance::RotateStateController: newStateControllerAddress=%s", newStateControllerAddr.String())
		state.Set(governance.VarRotateToAddress, isc.AddressToBytes(newStateControllerAddr))
		return nil
	}
	// here the new state controller address from the request equals to the state controller address in the anchor output
	// Two situations possible:
	// - either there's no need to rotate
	// - or it just has been rotated. In case of the second situation we emit a 'rotate' event
	if !ctx.StateAnchor().StateController.Equal(newStateControllerAddr) {
		// state controller address recorded in the blocklog is different from the new one
		// It means rotation happened
		eventRotate(ctx, newStateControllerAddr, ctx.StateAnchor().StateController)
		return nil
	}
	// no need to rotate because address does not change
	return nil
}

func addAllowedStateControllerAddress(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	addr := ctx.Params().MustGetAddress(governance.ParamStateControllerAddress)
	amap := collections.NewMap(ctx.State(), governance.VarAllowedStateControllerAddresses)
	amap.SetAt(isc.AddressToBytes(addr), []byte{0x01})
	return nil
}

func removeAllowedStateControllerAddress(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	addr := ctx.Params().MustGetAddress(governance.ParamStateControllerAddress)
	amap := collections.NewMap(ctx.State(), governance.VarAllowedStateControllerAddresses)
	amap.DelAt(isc.AddressToBytes(addr))
	return nil
}

func getAllowedStateControllerAddresses(ctx isc.SandboxView) dict.Dict {
	amap := collections.NewMapReadOnly(ctx.StateR(), governance.VarAllowedStateControllerAddresses)
	if amap.Len() == 0 {
		return nil
	}
	ret := dict.New()
	retArr := collections.NewArray(ret, governance.ParamAllowedStateControllerAddresses)
	amap.IterateKeys(func(elemKey []byte) bool {
		retArr.Push(elemKey)
		return true
	})
	return ret
}
