// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract, "EVM contract")

var (
	// EVM state
	FuncSendTransaction = coreutil.Func(evmnames.FuncSendTransaction)
	FuncCallContract    = coreutil.Func(evmnames.FuncCallContract)
	FuncGetChainID      = coreutil.ViewFunc(evmnames.FuncGetChainID)

	FuncRegisterERC20NativeToken              = coreutil.Func(evmnames.FuncRegisterERC20NativeToken)
	FuncRegisterERC20NativeTokenOnRemoteChain = coreutil.Func(evmnames.FuncRegisterERC20NativeTokenOnRemoteChain)
	FuncRegisterERC20ExternalNativeToken      = coreutil.Func(evmnames.FuncRegisterERC20ExternalNativeToken)
	FuncGetERC20ExternalNativeTokenAddress    = coreutil.ViewFunc(evmnames.FuncGetERC20ExternalNativeTokenAddress)
	FuncRegisterERC721NFTCollection           = coreutil.Func(evmnames.FuncRegisterERC721NFTCollection)

	// block context
	FuncOpenBlockContext  = coreutil.Func(evmnames.FuncOpenBlockContext)
	FuncCloseBlockContext = coreutil.Func(evmnames.FuncCloseBlockContext)
)

const (
	FieldTransaction      = evmnames.FieldTransaction
	FieldCallMsg          = evmnames.FieldCallMsg
	FieldChainID          = evmnames.FieldChainID
	FieldAddress          = evmnames.FieldAddress
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
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = uint16(1074) // IOTA -- get it?

	BlockKeepAll           = -1
	BlockKeepAmountDefault = int32(BlockKeepAll)
)

// Gas is charged in isc VM (L1 currencies), not ETH
var GasPrice = big.NewInt(0)

const (
	// KeyEVMState is the subrealm prefix for the EVM state, used by the emulator
	KeyEVMState = "s"

	// KeyISCMagic is the subrealm prefix for the ISC magic contract
	KeyISCMagic = "m"
)
