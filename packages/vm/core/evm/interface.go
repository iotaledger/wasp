// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap/buffer"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
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

func (e ERC20NativeTokenParams) ToDict() dict.Dict {
	return dict.Dict{
		FieldFoundrySN:         codec.Uint32.Encode(e.FoundrySN),
		FieldTokenName:         codec.String.Encode(e.Name),
		FieldTokenTickerSymbol: codec.String.Encode(e.TickerSymbol),
		FieldTokenDecimals:     codec.Uint8.Encode(e.Decimals),
	}
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

type RegisterERC20NativeTokenOnRemoteChainRequest struct {
	TargetChain *cryptolib.Address
	Token       ERC20NativeTokenParams
}

type RegisterERC20ExternalNativeTokenRequest struct {
	SourceChain        *cryptolib.Address
	FoundryTokenScheme iotago.TokenScheme
	Token              ERC20NativeTokenParams
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
