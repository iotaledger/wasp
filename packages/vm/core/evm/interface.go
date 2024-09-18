// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract)

var (
	// FuncSendTransaction is the main entry point, called by an
	// evmOffLedgerTxRequest in order to process an Ethereum tx (e.g.
	// eth_sendRawTransaction).
	FuncSendTransaction = coreutil.NewEP1(Contract, evmnames.FuncSendTransaction,
		coreutil.FieldWithCodec(codec.NewCodecFromBCS[*types.Transaction]()),
	)

	// FuncCallContract is the entry point called by an evmOffLedgerCallRequest
	// in order to process a view call or gas estimation (e.g. eth_call, eth_estimateGas).
	FuncCallContract = coreutil.NewEP11(Contract, evmnames.FuncCallContract,
		coreutil.FieldWithCodec(codec.NewCodecFromBCS[ethereum.CallMsg]()),
		coreutil.FieldWithCodec(codec.Bytes),
	)

	FuncRegisterERC20Coin = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20Coin,
		coreutil.FieldWithCodec(codec.CoinType),
	)
	FuncRegisterERC721NFTCollection = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC721NFTCollection,
		coreutil.FieldWithCodec(codec.ObjectID),
	)
	FuncNewL1Deposit = coreutil.NewEP3(Contract, evmnames.FuncNewL1Deposit,
		coreutil.FieldWithCodec(codec.AgentID),
		coreutil.FieldWithCodec(codec.EthereumAddress),
		coreutil.FieldWithCodec(codec.NewCodecEx(isc.AssetsFromBytes)),
	)
	ViewGetChainID = coreutil.NewViewEP01(Contract, evmnames.ViewGetChainID,
		coreutil.FieldWithCodec(codec.Uint16),
	)
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = uint16(1074) // IOTA -- get it?
)
