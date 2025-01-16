// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract)

var (
	// FuncSendTransaction is the main entry point, called by an
	// evmOffLedgerTxRequest in order to process an Ethereum tx (e.g.
	// eth_sendRawTransaction).
	FuncSendTransaction = coreutil.NewEP1(Contract, evmnames.FuncSendTransaction,
		coreutil.Field[*types.Transaction](),
	)

	// FuncCallContract is the entry point called by an evmOffLedgerCallRequest
	// in order to process a view call or gas estimation (e.g. eth_call, eth_estimateGas).
	FuncCallContract = coreutil.NewEP11(Contract, evmnames.FuncCallContract,
		coreutil.Field[ethereum.CallMsg](),
		coreutil.Field[[]byte](),
	)

	FuncRegisterERC20Coin = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20Coin,
		coreutil.Field[coin.Type](),
	)
	FuncRegisterERC721NFTCollection = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC721NFTCollection,
		coreutil.Field[iotago.ObjectID](),
	)
	FuncNewL1Deposit = coreutil.NewEP3(Contract, evmnames.FuncNewL1Deposit,
		coreutil.Field[isc.AgentID](),
		coreutil.Field[common.Address](),
		coreutil.Field[*isc.Assets](),
	)
	ViewGetChainID = coreutil.NewViewEP01(Contract, evmnames.ViewGetChainID,
		coreutil.Field[uint16](),
	)
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = uint16(1074) // IOTA -- get it?
)
