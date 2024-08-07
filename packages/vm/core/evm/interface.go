// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/ethereum/go-ethereum/common"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract)

var (
	// FuncSendTransaction is the main entry point, called by an
	// evmOffLedgerTxRequest in order to process an Ethereum tx (e.g.
	// eth_sendRawTransaction).
	FuncSendTransaction = Contract.Func(evmnames.FuncSendTransaction)

	// FuncCallContract is the entry point called by an evmOffLedgerCallRequest
	// in order to process a view call or gas estimation (e.g. eth_call, eth_estimateGas).
	FuncCallContract = Contract.Func(evmnames.FuncCallContract)

	FuncRegisterERC20NativeToken = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20NativeToken,
		InputRegisterERC20NativeToken{},
	)
	FuncRegisterERC20NativeTokenOnRemoteChain = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20NativeTokenOnRemoteChain,
		InputRegisterERC20NativeTokenOnRemoteChain{},
	)
	FuncRegisterERC20ExternalNativeToken = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20ExternalNativeToken,
		InputRegisterERC20ExteralNativeToken{},
	)
	FuncRegisterERC721NFTCollection = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC721NFTCollection,
		coreutil.FieldWithCodec(FieldNFTCollectionID, codec.ObjectID),
	)
	FuncNewL1Deposit = coreutil.NewEP1(Contract, evmnames.FuncNewL1Deposit,
		InputNewL1Deposit{},
	)

	ViewGetChainID = coreutil.NewViewEP01(Contract, evmnames.ViewGetChainID,
		coreutil.FieldWithCodec(FieldResult, codec.Uint16),
	)

	ViewGetERC20ExternalNativeTokenAddress = coreutil.NewViewEP11(Contract, evmnames.ViewGetERC20ExternalNativeTokenAddress,
		coreutil.FieldWithCodec(FieldNativeTokenID, codec.CoinType),
		coreutil.FieldWithCodecOptional(FieldResult, codec.EthereumAddress),
	)

	ViewGetERC721CollectionAddress = coreutil.NewViewEP11(Contract, evmnames.ViewGetERC721CollectionAddress,
		coreutil.FieldWithCodec(FieldNativeTokenID, codec.ObjectID),
		coreutil.FieldWithCodecOptional(FieldResult, codec.EthereumAddress),
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
	FieldNFTCollectionID    = evmnames.FieldNFTCollectionID   // ObjectID
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

func (e ERC20NativeTokenParams) ToDict() dict.Dict {
	return dict.Dict{
		FieldFoundrySN:         codec.Uint32.Encode(e.FoundrySN),
		FieldTokenName:         codec.String.Encode(e.Name),
		FieldTokenTickerSymbol: codec.String.Encode(e.TickerSymbol),
		FieldTokenDecimals:     codec.Uint8.Encode(e.Decimals),
	}
}

func ERC20NativeTokenParamsFromDict(d dict.Dict) (ret ERC20NativeTokenParams, err error) {
	ret.FoundrySN, err = codec.Uint32.Decode(d.Get(FieldFoundrySN))
	if err != nil {
		return
	}
	ret.Name, err = codec.String.Decode(d.Get(FieldTokenName))
	if err != nil {
		return
	}
	ret.TickerSymbol, err = codec.String.Decode(d.Get(FieldTokenTickerSymbol))
	if err != nil {
		return
	}
	ret.Decimals, err = codec.Uint8.Decode(d.Get(FieldTokenDecimals))
	return
}

type InputRegisterERC20NativeToken struct{}

func (InputRegisterERC20NativeToken) Encode(token ERC20NativeTokenParams) dict.Dict {
	return token.ToDict()
}

func (InputRegisterERC20NativeToken) Decode(d dict.Dict) (ERC20NativeTokenParams, error) {
	return ERC20NativeTokenParamsFromDict(d)
}

type RegisterERC20NativeTokenOnRemoteChainRequest struct {
	TargetChain *cryptolib.Address
	Token       ERC20NativeTokenParams
}

type InputRegisterERC20NativeTokenOnRemoteChain struct{}

func (InputRegisterERC20NativeTokenOnRemoteChain) Encode(p RegisterERC20NativeTokenOnRemoteChainRequest) dict.Dict {
	d := p.Token.ToDict()
	d[FieldTargetAddress] = codec.Address.Encode(p.TargetChain)
	return d
}

func (InputRegisterERC20NativeTokenOnRemoteChain) Decode(d dict.Dict) (ret RegisterERC20NativeTokenOnRemoteChainRequest, err error) {
	ret.Token, err = ERC20NativeTokenParamsFromDict(d)
	if err != nil {
		return
	}
	ret.TargetChain, err = codec.Address.Decode(d[FieldTargetAddress])
	return
}

type RegisterERC20ExternalNativeTokenRequest struct {
	SourceChain        *cryptolib.Address
	FoundryTokenScheme iotago.TokenScheme
	Token              ERC20NativeTokenParams
}

type InputRegisterERC20ExteralNativeToken struct{}

func (InputRegisterERC20ExteralNativeToken) Encode(p RegisterERC20ExternalNativeTokenRequest) dict.Dict {
	panic("TODO")
	d := p.Token.ToDict()
	d[FieldTargetAddress] = codec.Address.Encode(p.SourceChain)
	// d[FieldFoundryTokenScheme] = codec.TokenScheme.Encode(p.FoundryTokenScheme)
	return d
}

func (InputRegisterERC20ExteralNativeToken) Decode(d dict.Dict) (ret RegisterERC20ExternalNativeTokenRequest, err error) {
	panic("TODO")
	ret.Token, err = ERC20NativeTokenParamsFromDict(d)
	if err != nil {
		return
	}
	ret.SourceChain, err = codec.Address.Decode(d[FieldTargetAddress])
	// ret.FoundryTokenScheme, err = codec.TokenScheme.Decode(d[FieldFoundryTokenScheme])
	if err != nil {
		return
	}
	return
}

type NewL1DepositRequest struct {
	DepositOriginator isc.AgentID
	Receiver          common.Address
	Assets            *isc.Assets
}

type InputNewL1Deposit struct{}

func (InputNewL1Deposit) Decode(d dict.Dict) (ret NewL1DepositRequest, err error) {
	ret.DepositOriginator, err = codec.AgentID.Decode(d[FieldAgentIDDepositOriginator])
	if err != nil {
		return
	}
	ret.Receiver, err = codec.EthereumAddress.Decode(d[FieldAddress])
	if err != nil {
		return
	}
	ret.Assets, err = codec.NewCodecEx(isc.AssetsFromBytes).Decode(d[FieldAssets])
	if err != nil {
		return
	}
	return
}

func (InputNewL1Deposit) Encode(r NewL1DepositRequest) dict.Dict {
	return dict.Dict{
		FieldAddress:                  codec.EthereumAddress.Encode(r.Receiver),
		FieldAssets:                   codec.NewCodecEx(isc.AssetsFromBytes).Encode(r.Assets),
		FieldAgentIDDepositOriginator: codec.AgentID.Encode(r.DepositOriginator),
	}
}
