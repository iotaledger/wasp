// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"go.uber.org/zap/buffer"
	
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract)

var (
	// FuncSendTransaction is the main entry point, called by an
	// evmOffLedgerTxRequest in order to process an Ethereum tx (e.g.
	// eth_sendRawTransaction).
	FuncSendTransaction = coreutil.NewEP0(Contract, evmnames.FuncSendTransaction)

	// FuncCallContract is the entry point called by an evmOffLedgerCallRequest
	// in order to process a view call or gas estimation (e.g. eth_call, eth_estimateGas).
	FuncCallContract = coreutil.NewEP01(Contract, evmnames.FuncCallContract, coreutil.FieldWithCodec(codec.Bytes))

	FuncRegisterERC20NativeToken = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20NativeToken,
		ERC20NativeTokenParams{},
	)
	FuncRegisterERC20NativeTokenOnRemoteChain = coreutil.NewEP2(Contract, evmnames.FuncRegisterERC20NativeTokenOnRemoteChain,
		ERC20NativeTokenParams{},
		coreutil.FieldWithCodec(codec.Address),
	)
	FuncRegisterERC20ExternalNativeToken = coreutil.NewEP21(Contract, evmnames.FuncRegisterERC20ExternalNativeToken,
		ERC20NativeTokenParams{},
		coreutil.FieldWithCodec(codec.TokenScheme),
		coreutil.FieldWithCodec(codec.EthereumAddress),
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
	ViewGetERC20ExternalNativeTokenAddress = coreutil.NewViewEP11(Contract, evmnames.ViewGetERC20ExternalNativeTokenAddress,
		coreutil.FieldWithCodec(codec.CoinType),
		coreutil.FieldWithCodecOptional(codec.EthereumAddress),
	)
	ViewGetERC721CollectionAddress = coreutil.NewViewEP12(Contract, evmnames.ViewGetERC721CollectionAddress,
		coreutil.FieldWithCodec(codec.ObjectID),
		coreutil.FieldWithCodec(codec.Bool),
		coreutil.FieldWithCodec(codec.EthereumAddress),
	)
)

const (
	FieldTransaction      = evmnames.FieldTransaction
	FieldCallMsg          = evmnames.FieldCallMsg
	FieldChainID          = evmnames.FieldChainID
	FieldAddress          = evmnames.FieldAddress
	FieldAssets           = evmnames.FieldAssets
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
	FieldAgentIDWithdrawalTarget  = evmnames.FieldAgentIDWithdrawalTarget
	FieldFromAddress              = evmnames.FieldFromAddress
	FieldToAddress                = evmnames.FieldToAddress
)

const (
	// TODO shouldn't this be different between chain, to prevent replay attacks? (maybe derived from ISC ChainID)
	DefaultChainID = uint16(1074) // IOTA -- get it?
)

type ERC20NativeTokenParams struct {
	FoundrySN    uint32
	Name         string
	TickerSymbol string
	Decimals     uint8
}

func ERC20NativeTokenParamsFromBytes(d []byte) (ret ERC20NativeTokenParams, err error) {
	r := rwutil.NewBytesReader(d)
	ret.FoundrySN = r.ReadUint32()
	ret.Name = r.ReadString()
	ret.TickerSymbol = r.ReadString()
	ret.Decimals = r.ReadUint8()

	return ret, nil
}

func (ERC20NativeTokenParams) Encode(token ERC20NativeTokenParams) []byte {
	var buf buffer.Buffer
	w := rwutil.NewWriter(&buf)
	w.WriteUint32(token.FoundrySN)
	w.WriteString(token.Name)
	w.WriteString(token.TickerSymbol)
	w.WriteUint8(token.Decimals)
	return buf.Bytes()
}

func (ERC20NativeTokenParams) Decode(d []byte) (ERC20NativeTokenParams, error) {
	return ERC20NativeTokenParamsFromBytes(d)
}
