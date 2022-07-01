// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as sc from "./index";

export function funcMintSupply(ctx: wasmlib.ScFuncContext, f: sc.MintSupplyContext): void {
    let minted = ctx.minted();
    let mintedTokens = minted.tokenIDs();
    ctx.require(mintedTokens.length == 1, "need single minted color");
    let mintedColor = mintedTokens[0];
    let currentToken = f.state.registry().getToken(mintedColor);
    if (currentToken.exists()) {
        // should never happen, because transaction id is unique
        ctx.panic("TokenRegistry: registry for color already exists");
    }
    let token = new sc.Token();
    token.supply = minted.balance(mintedColor).uint64();
    token.mintedBy = ctx.caller();
    token.owner = ctx.caller();
    token.created = ctx.timestamp();
    token.updated = ctx.timestamp();
    token.description = f.params.description().value();
    token.userDefined = f.params.userDefined().value();
    if (token.description == "") {
        token.description = "no dscr";
    }
    currentToken.setValue(token);
    let colorList = f.state.tokenList();
    colorList.appendTokenID().setValue(mintedColor);
}

export function funcTransferOwnership(ctx: wasmlib.ScFuncContext, f: sc.TransferOwnershipContext): void {
    // TODO
}

export function funcUpdateMetadata(ctx: wasmlib.ScFuncContext, f: sc.UpdateMetadataContext): void {
    // TODO
}

export function viewGetInfo(ctx: wasmlib.ScViewContext, f: sc.GetInfoContext): void {
    // TODO
}

export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
	if (f.params.owner().exists()) {
		f.state.owner().setValue(f.params.owner().value());
		return;
	}
	f.state.owner().setValue(ctx.requestSender());
}
