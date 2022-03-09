// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// WasmConvertor converts ISCP data types to WasmLib data types
type WasmConvertor struct{}

func (cvt WasmConvertor) IscpAddress(address *wasmtypes.ScAddress) iotago.Address {
	buf := wasmtypes.AddressToBytes(*address)
	switch buf[0] {
	case wasmtypes.ScAddressAlias:
		iscpAliasAddress := new(iotago.AliasAddress)
		copy((*iscpAliasAddress)[:], buf[1:])
		return iscpAliasAddress
	case wasmtypes.ScAddressEd25519:
		iscpEd25519Address := new(iotago.Ed25519Address)
		copy((*iscpEd25519Address)[:], buf[1:])
		return iscpEd25519Address
	case wasmtypes.ScAddressNFT:
		iscpNFTAddress := new(iotago.NFTAddress)
		copy((*iscpNFTAddress)[:], buf[1:])
		return iscpNFTAddress
	default:
		panic("invalid ScAddress type")
	}
}

func (cvt WasmConvertor) IscpAgentID(agentID *wasmtypes.ScAgentID) *iscp.AgentID {
	address := agentID.Address()
	hname := agentID.Hname()
	return iscp.NewAgentID(cvt.IscpAddress(&address), cvt.IscpHname(hname))
}

func (cvt WasmConvertor) IscpAssets(assets wasmlib.ScAssets) *iscp.Allowance {
	iscpAllowance := iscp.NewEmptyAllowance()
	iscpAssets := iscpAllowance.Assets
	for color, amount := range assets {
		if color == wasmtypes.IOTA {
			iscpAssets.Iotas = amount
			continue
		}
		token := new(iotago.NativeToken)
		copy(token.ID[:], color.Bytes())
		token.Amount = new(big.Int)
		token.Amount.SetUint64(amount)
		iscpAssets.Tokens = append(iscpAssets.Tokens, token)
	}
	return iscpAllowance
}

func (cvt WasmConvertor) IscpChainID(chainID *wasmtypes.ScChainID) *iscp.ChainID {
	buf := wasmtypes.ChainIDToBytes(*chainID)
	iscpChainID := new(iscp.ChainID)
	copy(iscpChainID[:], buf)
	return iscpChainID
}

// TODO switch WasmLib from Color to Token
func (cvt WasmConvertor) IscpColor(color *wasmtypes.ScColor) *iotago.NativeTokenID {
	buf := wasmtypes.ColorToBytes(*color)
	iscpTokenID := new(iotago.NativeTokenID)
	copy(iscpTokenID[:], buf)
	return iscpTokenID
}

func (cvt WasmConvertor) IscpHash(hash *wasmtypes.ScHash) *hashing.HashValue {
	buf := wasmtypes.HashToBytes(*hash)
	iscpHashValue := new(hashing.HashValue)
	copy(iscpHashValue[:], buf)
	return iscpHashValue
}

func (cvt WasmConvertor) IscpHname(hname wasmtypes.ScHname) iscp.Hname {
	return iscp.Hname(hname)
}

func (cvt WasmConvertor) IscpRequestID(requestID *wasmtypes.ScRequestID) *iscp.RequestID {
	buf := wasmtypes.RequestIDToBytes(*requestID)
	iscpRequestID := new(iscp.RequestID)
	copy(iscpRequestID.TransactionID[:], buf)
	iscpRequestID.TransactionOutputIndex = wasmtypes.Uint16FromBytes(buf[wasmtypes.ScHashLength:])
	return iscpRequestID
}

func (cvt WasmConvertor) ScAddress(address iotago.Address) wasmtypes.ScAddress {
	buf := iscp.BytesFromAddress(address)
	return wasmtypes.AddressFromBytes(buf)
}

func (cvt WasmConvertor) ScAgentID(agentID *iscp.AgentID) wasmtypes.ScAgentID {
	return wasmtypes.NewScAgentID(cvt.ScAddress(agentID.Address()), cvt.ScHname(agentID.Hname()))
}

func (cvt WasmConvertor) ScBalances(assets *iscp.Assets) wasmlib.ScAssets {
	scAssets := make(wasmlib.ScAssets)
	if assets.Iotas != 0 {
		scAssets[wasmtypes.IOTA] = assets.Iotas
	}
	for _, token := range assets.Tokens {
		color := cvt.ScColor(&token.ID)
		// TODO handle big.Int
		scAssets[color] = token.Amount.Uint64()
	}
	return scAssets
}

func (cvt WasmConvertor) ScChainID(chainID *iscp.ChainID) wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(chainID.Bytes())
}

// TODO switch WasmLib from Color to Token
func (cvt WasmConvertor) ScColor(color *iotago.NativeTokenID) wasmtypes.ScColor {
	return wasmtypes.ColorFromBytes(color[:])
}

func (cvt WasmConvertor) ScHash(hash hashing.HashValue) wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(hash.Bytes())
}

func (cvt WasmConvertor) ScHname(hname iscp.Hname) wasmtypes.ScHname {
	return wasmtypes.ScHname(hname)
}

func (cvt WasmConvertor) ScRequestID(requestID iscp.RequestID) wasmtypes.ScRequestID {
	return wasmtypes.RequestIDFromBytes(requestID.Bytes())
}
