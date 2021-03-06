// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
//////// DO NOT CHANGE THIS FILE! ////////
// Change the json schema instead

package test

import "github.com/iotaledger/wasp/packages/iscp"

const (
	ScName        = "dividend"
	ScDescription = "Simple dividend smart contract"
	HScName       = iscp.Hname(0xcce2e239)
)

const (
	ParamAddress = "address"
	ParamFactor  = "factor"
	ParamOwner   = "owner"
)

const ResultFactor = "factor"

const (
	StateFactor      = "factor"
	StateMemberList  = "memberList"
	StateMembers     = "members"
	StateOwner       = "owner"
	StateTotalFactor = "totalFactor"
)

const (
	FuncDivide    = "divide"
	FuncInit      = "init"
	FuncMember    = "member"
	FuncSetOwner  = "setOwner"
	ViewGetFactor = "getFactor"
)

const (
	HFuncDivide    = iscp.Hname(0xc7878107)
	HFuncInit      = iscp.Hname(0x1f44d644)
	HFuncMember    = iscp.Hname(0xc07da2cb)
	HFuncSetOwner  = iscp.Hname(0x2a15fe7b)
	HViewGetFactor = iscp.Hname(0x0ee668fe)
)
