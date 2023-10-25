// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract)

var (
	// FuncSendTransaction is the main entry point, called by an
	// evmOffLedgerTxRequest in order to process an Ethereum tx (e.g.
	// eth_sendRawTransaction).
	FuncSendTransaction = coreutil.Func(evmnames.FuncSendTransaction)

	// FuncCallContract is the entry point called by an evmOffLedgerCallRequest
	// in order to process a view call or gas estimation (e.g. eth_call, eth_estimateGas).
	FuncCallContract = coreutil.Func(evmnames.FuncCallContract)

	FuncGetChainID = coreutil.ViewFunc(evmnames.FuncGetChainID)

	FuncRegisterERC20NativeToken              = coreutil.Func(evmnames.FuncRegisterERC20NativeToken)
	FuncRegisterERC20NativeTokenOnRemoteChain = coreutil.Func(evmnames.FuncRegisterERC20NativeTokenOnRemoteChain)
	FuncRegisterERC20ExternalNativeToken      = coreutil.Func(evmnames.FuncRegisterERC20ExternalNativeToken)
	FuncGetERC20ExternalNativeTokenAddress    = coreutil.ViewFunc(evmnames.FuncGetERC20ExternalNativeTokenAddress)
	FuncRegisterERC721NFTCollection           = coreutil.Func(evmnames.FuncRegisterERC721NFTCollection)

	FuncNewL1Deposit = coreutil.Func(evmnames.FuncNewL1Deposit)
)

const (
	FieldTransaction      = evmnames.FieldTransaction
	FieldCallMsg          = evmnames.FieldCallMsg
	FieldChainID          = evmnames.FieldChainID
	FieldAddress          = evmnames.FieldAddress
	FieldAssets           = evmnames.FieldAssets
	FieldKey              = evmnames.FieldKey
	FieldAgentID          = evmnames.FieldAgentID
	FieldTransactionIndex = evmnames.FieldTransactionIndex
	FieldTransactionHash  = evmnames.FieldTransactionHash
	FieldResult           = evmnames.FieldResult
	FieldBlockNumber      = evmnames.FieldBlockNumber
	FieldBlockHash        = evmnames.FieldBlockHash
	FieldFilterQuery      = evmnames.FieldFilterQuery
	FieldBlockKeepAmount  = evmnames.FieldBlockKeepAmount // int32

	FieldNativeTokenID      = evmnames.FieldNativeTokenID
	FieldFoundrySN          = evmnames.FieldFoundrySN         // uint32
	FieldTokenName          = evmnames.FieldTokenName         // string
	FieldTokenTickerSymbol  = evmnames.FieldTokenTickerSymbol // string
	FieldTokenDecimals      = evmnames.FieldTokenDecimals     // uint8
	FieldNFTCollectionID    = evmnames.FieldNFTCollectionID   // NFTID
	FieldFoundryTokenScheme = evmnames.FieldFoundryTokenScheme
	FieldTargetAddress      = evmnames.FieldTargetAddress

	FieldAgentIDDepositOriginator = evmnames.FieldAgentIDDepositOriginator
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = uint16(1074) // IOTA -- get it?
)
