// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// NOTE: These strings correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFuncNames = []string{
	"nil",
	"FnAccountID",
	"#FnBalance",
	"FnBalances",
	"FnBlockContext",
	"FnCall",
	"FnCaller",
	"FnChainID",
	"FnChainOwnerID",
	"FnContract",
	"FnContractCreator",
	"#FnDeployContract",
	"FnEntropy",
	"$FnEvent",
	"FnIncomingTransfer",
	"$FnLog",
	"FnMinted",
	"$FnPanic",
	"FnParams",
	"#FnPost",
	"FnRequest",
	"FnRequestID",
	"#FnResults",
	"#FnSend",
	"#FnStateAnchor",
	"FnTimestamp",
	"$FnTrace",
	"$FnUtilsBase58Decode",
	"#FnUtilsBase58Encode",
	"#FnUtilsBlsAddress",
	"#FnUtilsBlsAggregate",
	"#FnUtilsBlsValid",
	"#FnUtilsEd25519Address",
	"#FnUtilsEd25519Valid",
	"#FnUtilsHashBlake2b",
	"$FnUtilsHashName",
	"#FnUtilsHashSha3",
}

func traceSandbox(funcNr int32, params []byte) string {
	name := sandboxFuncNames[-funcNr]
	if name[0] == '$' {
		return name[1:] + ", " + string(params)
	}
	if name[0] != '#' {
		return name
	}
	return name[1:] + ", " + wasmtypes.Hex(params)
}
