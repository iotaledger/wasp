// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governanceimpl

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// rotateStateController the entry point is called when committee is about to be rotated to the new address
// If it fails, nothing happens and the state has trace of the failure in the state
// If it is successful VM takes over and replaces resulting transaction with
// governance transition. The state of the chain remains unchanged
func rotateStateController(ctx isc.Sandbox, newStateControllerAddr *cryptolib.Address) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	// check is address is allowed
	state := governance.NewStateWriterFromSandbox(ctx)
	amap := state.AllowedStateControllerAddressesMap()
	if !amap.HasAt(newStateControllerAddr.Bytes()) {
		panic(vm.ErrUnauthorized)
	}

	if !newStateControllerAddr.Equals(ctx.StateAnchor().StateController) {
		// rotate request to another address has been issued. State update will be taken over by VM and will have no effect
		// By setting VarRotateToAddress we signal the VM this special situation
		// VarRotateToAddress value should never persist in the state
		ctx.Log().Infof("Governance::RotateStateController: newStateControllerAddress=%s", newStateControllerAddr.String())
		state.SetRotationAddress(newStateControllerAddr)
		return nil
	}
	// here the new state controller address from the request equals to the state controller address in the anchor output
	// Two situations possible:
	// - either there's no need to rotate
	// - or it just has been rotated. In case of the second situation we emit a 'rotate' event
	if !ctx.StateAnchor().StateController.Equals(newStateControllerAddr) {
		// state controller address recorded in the blocklog is different from the new one
		// It means rotation happened
		eventRotate(ctx, newStateControllerAddr, ctx.StateAnchor().StateController)
		return nil
	}
	// no need to rotate because address does not change
	return nil
}

func addAllowedStateControllerAddress(ctx isc.Sandbox, addr *cryptolib.Address) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	amap := state.AllowedStateControllerAddressesMap()
	amap.SetAt(codec.Address.Encode(addr), []byte{0x01})
	return nil
}

func removeAllowedStateControllerAddress(ctx isc.Sandbox, addr *cryptolib.Address) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	amap := state.AllowedStateControllerAddressesMap()
	amap.DelAt(codec.Address.Encode(addr))
	return nil
}

func getAllowedStateControllerAddresses(ctx isc.SandboxView) []*cryptolib.Address {
	state := governance.NewStateReaderFromSandbox(ctx)
	amap := state.AllowedStateControllerAddressesMap()
	ret := make([]*cryptolib.Address, 0)
	amap.IterateKeys(func(elemKey []byte) bool {
		ret = append(ret, lo.Must(codec.Address.Decode(elemKey)))
		return true
	})
	return ret
}
