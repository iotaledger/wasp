// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tokenregistryimpl

import (
	"github.com/iotaledger/wasp/contracts/wasm/tokenregistry/go/tokenregistry"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

func funcMintSupply(ctx wasmlib.ScFuncContext, f *MintSupplyContext) {
	minted := ctx.Minted()
	mintedTokens := minted.TokenIDs()
	ctx.Require(len(mintedTokens) == 1, "need single minted color")
	mintedColor := mintedTokens[0]
	currentToken := f.State.Registry().GetToken(*mintedColor)
	if currentToken.Exists() {
		// should never happen, because transaction id is unique
		ctx.Panic("TokenRegistry: registry for color already exists")
	}
	token := &tokenregistry.Token{
		Supply:      minted.Balance(mintedColor).Uint64(),
		MintedBy:    ctx.Caller(),
		Owner:       ctx.Caller(),
		Created:     ctx.Timestamp(),
		Updated:     ctx.Timestamp(),
		Description: f.Params.Description().Value(),
		UserDefined: f.Params.UserDefined().Value(),
	}
	if token.Description == "" {
		token.Description += "no dscr"
	}
	currentToken.SetValue(token)
	colorList := f.State.TokenList()
	colorList.AppendTokenID().SetValue(*mintedColor)
}

func funcTransferOwnership(_ wasmlib.ScFuncContext, _ *TransferOwnershipContext) {
	// TODO implement
}

func funcUpdateMetadata(_ wasmlib.ScFuncContext, _ *UpdateMetadataContext) {
	// TODO implement
}

func viewGetInfo(_ wasmlib.ScViewContext, _ *GetInfoContext) {
	// TODO implement
}

func funcInit(ctx wasmlib.ScFuncContext, f *InitContext) {
	if f.Params.Owner().Exists() {
		f.State.Owner().SetValue(f.Params.Owner().Value())
		return
	}
	f.State.Owner().SetValue(ctx.RequestSender())
}
