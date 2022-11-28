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
	FuncSendTransaction                     = coreutil.Func(evmnames.FuncSendTransaction)
	FuncEstimateGas                         = coreutil.Func(evmnames.FuncEstimateGas)
	FuncGetBalance                          = coreutil.ViewFunc(evmnames.FuncGetBalance)
	FuncCallContract                        = coreutil.ViewFunc(evmnames.FuncCallContract)
	FuncGetNonce                            = coreutil.ViewFunc(evmnames.FuncGetNonce)
	FuncGetReceipt                          = coreutil.ViewFunc(evmnames.FuncGetReceipt)
	FuncGetCode                             = coreutil.ViewFunc(evmnames.FuncGetCode)
	FuncGetBlockNumber                      = coreutil.ViewFunc(evmnames.FuncGetBlockNumber)
	FuncGetBlockByNumber                    = coreutil.ViewFunc(evmnames.FuncGetBlockByNumber)
	FuncGetBlockByHash                      = coreutil.ViewFunc(evmnames.FuncGetBlockByHash)
	FuncGetTransactionByHash                = coreutil.ViewFunc(evmnames.FuncGetTransactionByHash)
	FuncGetTransactionByBlockHashAndIndex   = coreutil.ViewFunc(evmnames.FuncGetTransactionByBlockHashAndIndex)
	FuncGetTransactionByBlockNumberAndIndex = coreutil.ViewFunc(evmnames.FuncGetTransactionByBlockNumberAndIndex)
	FuncGetTransactionCountByBlockHash      = coreutil.ViewFunc(evmnames.FuncGetTransactionCountByBlockHash)
	FuncGetTransactionCountByBlockNumber    = coreutil.ViewFunc(evmnames.FuncGetTransactionCountByBlockNumber)
	FuncGetStorage                          = coreutil.ViewFunc(evmnames.FuncGetStorage)
	FuncGetLogs                             = coreutil.ViewFunc(evmnames.FuncGetLogs)
	FuncGetChainID                          = coreutil.ViewFunc(evmnames.FuncGetChainID)

	FuncRegisterERC20NativeToken = coreutil.Func(evmnames.FuncRegisterERC20NativeToken)

	// block context
	FuncOpenBlockContext  = coreutil.Func(evmnames.FuncOpenBlockContext)
	FuncCloseBlockContext = coreutil.Func(evmnames.FuncCloseBlockContext)
)

const (
	FieldTransaction      = evmnames.FieldTransaction
	FieldCallMsg          = evmnames.FieldCallMsg
	FieldChainID          = evmnames.FieldChainID
	FieldGenesisAlloc     = evmnames.FieldGenesisAlloc
	FieldAddress          = evmnames.FieldAddress
	FieldKey              = evmnames.FieldKey
	FieldAgentID          = evmnames.FieldAgentID
	FieldTransactionIndex = evmnames.FieldTransactionIndex
	FieldTransactionHash  = evmnames.FieldTransactionHash
	FieldResult           = evmnames.FieldResult
	FieldBlockNumber      = evmnames.FieldBlockNumber
	FieldBlockHash        = evmnames.FieldBlockHash
	FieldBlockGasLimit    = evmnames.FieldBlockGasLimit
	FieldFilterQuery      = evmnames.FieldFilterQuery
	FieldBlockKeepAmount  = evmnames.FieldBlockKeepAmount // int32

	FieldFoundrySN         = evmnames.FieldFoundrySN         // uint32
	FieldTokenName         = evmnames.FieldTokenName         // string
	FieldTokenTickerSymbol = evmnames.FieldTokenTickerSymbol // string
	FieldTokenDecimals     = evmnames.FieldTokenDecimals     // uint8
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = uint16(1074) // IOTA -- get it?

	BlockGasLimitDefault = uint64(15000000)

	BlockKeepAll           = -1
	BlockKeepAmountDefault = BlockKeepAll
)

// Gas is charged in isc VM (L1 currencies), not ETH
var GasPrice = big.NewInt(0)
