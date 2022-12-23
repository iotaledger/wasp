// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	iscvm "github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getEntropy
func (h *magicContractHandler) GetEntropy() hashing.HashValue {
	return h.ctx.GetEntropy()
}

// handler for ISCSandbox::triggerEvent
func (h *magicContractHandler) TriggerEvent(s string) {
	h.ctx.Event(s)
}

// handler for ISCSandbox::getRequestID
func (h *magicContractHandler) GetRequestID() iscmagic.ISCRequestID {
	return iscmagic.WrapISCRequestID(h.ctx.Request().ID())
}

// handler for ISCSandbox::getSenderAccount
func (h *magicContractHandler) GetSenderAccount() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.Request().SenderAccount())
}

// handler for ISCSandbox::allow
func (h *magicContractHandler) Allow(target common.Address, allowance iscmagic.ISCAllowance) {
	addToAllowance(h.ctx, h.caller.Address(), target, allowance.Unwrap())
}

// handler for ISCSandbox::takeAllowedFunds
func (h *magicContractHandler) TakeAllowedFunds(addr common.Address, allowance iscmagic.ISCAllowance) {
	taken := subtractFromAllowance(h.ctx, addr, h.caller.Address(), allowance.Unwrap())
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(addr),
		isc.NewEthereumAddressAgentID(h.caller.Address()),
		taken.Assets,
		taken.NFTs,
	)
}

// handler for ISCSandbox::send
func (h *magicContractHandler) Send(
	targetAddress iscmagic.L1Address,
	fungibleTokens iscmagic.ISCFungibleTokens,
	adjustMinimumStorageDeposit bool,
	metadata iscmagic.ISCSendMetadata,
	sendOptions iscmagic.ISCSendOptions,
) {
	req := isc.RequestParameters{
		TargetAddress:                 targetAddress.MustUnwrap(),
		FungibleTokens:                fungibleTokens.Unwrap(),
		AdjustToMinimumStorageDeposit: adjustMinimumStorageDeposit,
		Metadata:                      metadata.Unwrap(),
		Options:                       sendOptions.Unwrap(),
	}
	h.adjustStorageDeposit(req)

	h.moveAssetsToCommonAccount(req.FungibleTokens, nil)

	// assert that remaining tokens in the sender's account are enough to pay for the gas budget
	if !h.ctx.HasInAccount(
		h.ctx.Request().SenderAccount(),
		h.ctx.Privileged().TotalGasTokens(),
	) {
		panic(iscvm.ErrNotEnoughTokensLeftForGas)
	}
	h.ctx.Send(req)
}

// handler for ISCSandbox::sendAsNFT
func (h *magicContractHandler) SendAsNFT(
	targetAddress iscmagic.L1Address,
	fungibleTokens iscmagic.ISCFungibleTokens,
	nftID iscmagic.NFTID,
	adjustMinimumStorageDeposit bool,
	metadata iscmagic.ISCSendMetadata,
	sendOptions iscmagic.ISCSendOptions,
) {
	req := isc.RequestParameters{
		TargetAddress:                 targetAddress.MustUnwrap(),
		FungibleTokens:                fungibleTokens.Unwrap(),
		AdjustToMinimumStorageDeposit: adjustMinimumStorageDeposit,
		Metadata:                      metadata.Unwrap(),
		Options:                       sendOptions.Unwrap(),
	}
	id := nftID.Unwrap()
	h.adjustStorageDeposit(req)

	// make sure that allowance <= sent tokens, so that the target contract does not
	// spend from the common account
	h.ctx.Requiref(
		isc.NewAllowanceFungibleTokens(req.FungibleTokens).AddNFTs(id).SpendFromBudget(req.Metadata.Allowance),
		"allowance must not be greater than sent tokens",
	)

	h.moveAssetsToCommonAccount(req.FungibleTokens, []iotago.NFTID{id})

	// assert that remaining tokens in the sender's account are enough to pay for the gas budget
	if !h.ctx.HasInAccount(
		h.ctx.Request().SenderAccount(),
		h.ctx.Privileged().TotalGasTokens(),
	) {
		panic(iscvm.ErrNotEnoughTokensLeftForGas)
	}
	h.ctx.SendAsNFT(req, id)
}

// handler for ISCSandbox::call
func (h *magicContractHandler) Call(
	contractHname uint32,
	entryPoint uint32,
	params iscmagic.ISCDict,
	allowance iscmagic.ISCAllowance,
) iscmagic.ISCDict {
	a := allowance.Unwrap()
	h.moveAssetsToCommonAccount(a.Assets, a.NFTs)
	callRet := h.ctx.Call(
		isc.Hname(contractHname),
		isc.Hname(entryPoint),
		params.Unwrap(),
		a,
	)
	return iscmagic.WrapISCDict(callRet)
}

func (h *magicContractHandler) adjustStorageDeposit(req isc.RequestParameters) {
	sd := h.ctx.EstimateRequiredStorageDeposit(req)
	if req.FungibleTokens.BaseTokens < sd {
		if !req.AdjustToMinimumStorageDeposit {
			panic(fmt.Sprintf(
				"base tokens (%d) not enough to cover storage deposit (%d)",
				req.FungibleTokens.BaseTokens,
				sd,
			))
		}
		req.FungibleTokens.BaseTokens = sd
	}
}

// moveAssetsToCommonAccount moves the assets from the caller's L2 account to the common
// account before sending to L1
func (h *magicContractHandler) moveAssetsToCommonAccount(fungibleTokens *isc.FungibleTokens, nftIDs []iotago.NFTID) {
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(h.caller.Address()),
		h.ctx.AccountID(),
		fungibleTokens,
		nftIDs,
	)
}
