// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gotemplates

var thunksGo = map[string]string{
	// *******************************
	"thunks.go": `
//nolint:dupl
package $package$+impl

import (
$#if funcs importPackage
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

var exportMap = wasmlib.ScExportMap{
	Names: []string{
$#each func libExportName
	},
	Funcs: []wasmlib.ScFuncContextFunction{
$#each func libExportFunc
	},
	Views: []wasmlib.ScViewContextFunction{
$#each func libExportView
	},
}

func OnDispatch(index int32) {
	exportMap.Dispatch(index)
}
$#each func libThunk
`,
	// *******************************
	"importPackage": `
	"$module/go/$package"
`,
	// *******************************
	"libExportName": `
		$package.$Kind$FuncName,
`,
	// *******************************
	"libExportFunc": `
$#if func libExportFuncThunk
`,
	// *******************************
	"libExportFuncThunk": `
		$kind$FuncName$+Thunk,
`,
	// *******************************
	"libExportView": `
$#if view libExportViewThunk
`,
	// *******************************
	"libExportViewThunk": `
		$kind$FuncName$+Thunk,
`,
	// *******************************
	"libThunk": `
$#emit alignCalculate

type $FuncName$+Context struct {
$#if func PackageEvents
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if state PackageState
}

func $kind$FuncName$+Thunk(ctx wasmlib.Sc$Kind$+Context) {
	ctx.Log("$package.$kind$FuncName")
$#if result initResultDict
	f := &$FuncName$+Context{
$#if param ImmutableFuncNameParamsInit
$#if result MutableFuncNameResultsInit
$#if state PackageStateInit
	}
$#emit accessCheck
$#each mandatory requireMandatory
	$kind$FuncName(ctx, f)
$#if result returnResultDict
	ctx.Log("$package.$kind$FuncName ok")
}
`,
	// *******************************
	"initResultDict": `
	results := wasmlib.NewScDict()
`,
	// *******************************
	"PackageEvents": `
$#if events PackageEventsExist
`,
	// *******************************
	"PackageEventsExist": `
	Events$align $package.$Package$+Events
`,
	// *******************************
	"ImmutableFuncNameParams": `
	Params$align $package.Immutable$FuncName$+Params
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
		Params:$align $package.NewImmutable$FuncName$+Params(),
`,
	// *******************************
	"MutableFuncNameResults": `
	Results $package.Mutable$FuncName$+Results
`,
	// *******************************
	"MutableFuncNameResultsInit": `
		Results: $package.NewMutable$FuncName$+Results(results),
`,
	// *******************************
	"PackageState": `
$#if func MutablePackageState
$#if view ImmutablePackageState
`,
	// *******************************
	"MutablePackageState": `
	State$salign $package.Mutable$Package$+State
`,
	// *******************************
	"ImmutablePackageState": `
	State$salign $package.Immutable$Package$+State
`,
	// *******************************
	"PackageStateInit": `
$#if func MutablePackageStateInit
$#if view ImmutablePackageStateInit
`,
	// *******************************
	"MutablePackageStateInit": `
		State:$salign $package.NewMutable$Package$+State(),
`,
	// *******************************
	"ImmutablePackageStateInit": `
		State:$salign $package.NewImmutable$Package$+State(),
`,
	// *******************************
	"returnResultDict": `
	ctx.Results(results)
`,
	// *******************************
	"requireMandatory": `
	ctx.Require(f.Params.$FldName().Exists(), "missing mandatory param: $fldName")
`,
	// *******************************
	"accessCheck": `
$#set accessFinalize accessOther
$#emit caseAccess$funcAccess
$#emit $accessFinalize
`,
	// *******************************
	"caseAccess": `
$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccessself": `

$#each funcAccessComment _funcAccessComment
	ctx.Require(ctx.Caller() == ctx.AccountID(), "no permission")

$#set accessFinalize accessDone
`,
	// *******************************
	"caseAccesschain": `

$#each funcAccessComment _funcAccessComment
	ctx.Require(ctx.Caller() == ctx.ChainOwnerID(), "no permission")

$#set accessFinalize accessDone
`,
	// *******************************
	"accessOther": `

$#each funcAccessComment _funcAccessComment
	access := f.State.$FuncAccess()
	ctx.Require(access.Exists(), "access not set: $funcAccess")
	ctx.Require(ctx.Caller() == access.Value(), "no permission")

`,
	// *******************************
	"accessDone": `
`,
}
