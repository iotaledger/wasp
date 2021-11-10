// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package coreroot

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

type ImmutableDeployContractParams struct {
	id int32
}

func (s ImmutableDeployContractParams) Description() wasmlib.ScImmutableString {
	return wasmlib.NewScImmutableString(s.id, wasmlib.KeyID(ParamDescription))
}

func (s ImmutableDeployContractParams) Name() wasmlib.ScImmutableString {
	return wasmlib.NewScImmutableString(s.id, wasmlib.KeyID(ParamName))
}

func (s ImmutableDeployContractParams) ProgramHash() wasmlib.ScImmutableHash {
	return wasmlib.NewScImmutableHash(s.id, wasmlib.KeyID(ParamProgramHash))
}

type MutableDeployContractParams struct {
	id int32
}

func (s MutableDeployContractParams) Description() wasmlib.ScMutableString {
	return wasmlib.NewScMutableString(s.id, wasmlib.KeyID(ParamDescription))
}

func (s MutableDeployContractParams) Name() wasmlib.ScMutableString {
	return wasmlib.NewScMutableString(s.id, wasmlib.KeyID(ParamName))
}

func (s MutableDeployContractParams) ProgramHash() wasmlib.ScMutableHash {
	return wasmlib.NewScMutableHash(s.id, wasmlib.KeyID(ParamProgramHash))
}

type ImmutableGrantDeployPermissionParams struct {
	id int32
}

func (s ImmutableGrantDeployPermissionParams) Deployer() wasmlib.ScImmutableAgentID {
	return wasmlib.NewScImmutableAgentID(s.id, wasmlib.KeyID(ParamDeployer))
}

type MutableGrantDeployPermissionParams struct {
	id int32
}

func (s MutableGrantDeployPermissionParams) Deployer() wasmlib.ScMutableAgentID {
	return wasmlib.NewScMutableAgentID(s.id, wasmlib.KeyID(ParamDeployer))
}

type ImmutableRevokeDeployPermissionParams struct {
	id int32
}

func (s ImmutableRevokeDeployPermissionParams) Deployer() wasmlib.ScImmutableAgentID {
	return wasmlib.NewScImmutableAgentID(s.id, wasmlib.KeyID(ParamDeployer))
}

type MutableRevokeDeployPermissionParams struct {
	id int32
}

func (s MutableRevokeDeployPermissionParams) Deployer() wasmlib.ScMutableAgentID {
	return wasmlib.NewScMutableAgentID(s.id, wasmlib.KeyID(ParamDeployer))
}

type ImmutableFindContractParams struct {
	id int32
}

func (s ImmutableFindContractParams) Hname() wasmlib.ScImmutableHname {
	return wasmlib.NewScImmutableHname(s.id, wasmlib.KeyID(ParamHname))
}

type MutableFindContractParams struct {
	id int32
}

func (s MutableFindContractParams) Hname() wasmlib.ScMutableHname {
	return wasmlib.NewScMutableHname(s.id, wasmlib.KeyID(ParamHname))
}
