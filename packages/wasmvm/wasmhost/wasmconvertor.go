// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// WasmConvertor converts ISCP data types to WasmLib data types
type WasmConvertor struct{}

func (cvt WasmConvertor) IscpAddress(address *wasmtypes.ScAddress) iotago.Address {
	buf := wasmtypes.AddressToBytes(*address)
	addr, _, err := isc.AddressFromBytes(buf)
	if err != nil {
		panic(err)
	}
	return addr
}

func (cvt WasmConvertor) IscpAgentID(agentID *wasmtypes.ScAgentID) isc.AgentID {
	scAddress := agentID.Address()
	address := cvt.IscpAddress(&scAddress)
	if agentID.IsAddress() {
		return isc.NewAgentID(address)
	}
	if agentID.IsContract() {
		chainID := isc.ChainIDFromAddress(address.(*iotago.AliasAddress))
		return isc.NewContractAgentID(&chainID, cvt.IscpHname(agentID.Hname()))
	}

	// TODO implement missing agent id types
	panic("WasmConvertor.IscpAgentID kind")
}

func (cvt WasmConvertor) IscpAllowance(assets *wasmlib.ScAssets) *isc.Allowance {
	iscpAllowance := isc.NewAllowanceBaseTokens(assets.BaseTokens)
	iscpAssets := iscpAllowance.Assets
	for tokenID, amount := range assets.Tokens {
		token := &iotago.NativeToken{
			ID:     *cvt.IscpTokenID(&tokenID),
			Amount: cvt.IscpBigInt(amount),
		}
		iscpAssets.Tokens = append(iscpAssets.Tokens, token)
	}
	for _, nftID := range assets.NftIDs {
		nft := cvt.IscpNFTID(nftID)
		iscpAllowance.NFTs = append(iscpAllowance.NFTs, *nft)
	}
	return iscpAllowance
}

func (cvt WasmConvertor) IscpBigInt(amount wasmtypes.ScBigInt) *big.Int {
	res := new(big.Int)
	res.SetBytes(wasmtypes.BigIntToBytes(amount))
	return res
}

func (cvt WasmConvertor) IscpChainID(chainID *wasmtypes.ScChainID) *isc.ChainID {
	buf := wasmtypes.ChainIDToBytes(*chainID)
	iscpChainID := new(isc.ChainID)
	copy(iscpChainID[:], buf)
	return iscpChainID
}

func (cvt WasmConvertor) IscpHash(hash *wasmtypes.ScHash) *hashing.HashValue {
	buf := wasmtypes.HashToBytes(*hash)
	iscpHashValue := new(hashing.HashValue)
	copy(iscpHashValue[:], buf)
	return iscpHashValue
}

func (cvt WasmConvertor) IscpHname(hname wasmtypes.ScHname) isc.Hname {
	return isc.Hname(hname)
}

func (cvt WasmConvertor) IscpNFTID(nftID *wasmtypes.ScNftID) *iotago.NFTID {
	buf := wasmtypes.NftIDToBytes(*nftID)
	iscpNFTID := new(iotago.NFTID)
	copy(iscpNFTID[:], buf)
	return iscpNFTID
}

func (cvt WasmConvertor) IscpRequestID(requestID *wasmtypes.ScRequestID) *isc.RequestID {
	buf := wasmtypes.RequestIDToBytes(*requestID)
	iscpRequestID := new(isc.RequestID)
	copy(iscpRequestID.TransactionID[:], buf)
	iscpRequestID.TransactionOutputIndex = wasmtypes.Uint16FromBytes(buf[wasmtypes.ScHashLength:])
	return iscpRequestID
}

func (cvt WasmConvertor) IscpTokenID(tokenID *wasmtypes.ScTokenID) *iotago.NativeTokenID {
	buf := wasmtypes.TokenIDToBytes(*tokenID)
	iscpTokenID := new(iotago.NativeTokenID)
	copy(iscpTokenID[:], buf)
	return iscpTokenID
}

func (cvt WasmConvertor) ScAddress(address iotago.Address) wasmtypes.ScAddress {
	buf := isc.BytesFromAddress(address)
	return wasmtypes.AddressFromBytes(buf)
}

func (cvt WasmConvertor) ScAgentID(agentID isc.AgentID) wasmtypes.ScAgentID {
	switch agentID.Kind() {
	case isc.AgentIDKindAddress:
		addr, _ := isc.AddressFromAgentID(agentID)
		return wasmtypes.NewScAgentIDFromAddress(cvt.ScAddress(addr))
	case isc.AgentIDKindContract:
		chainID, _ := isc.AddressFromAgentID(agentID)
		hname, _ := isc.HnameFromAgentID(agentID)
		return wasmtypes.NewScAgentID(cvt.ScAddress(chainID), cvt.ScHname(hname))
	default:
		// TODO implement missing agent id types
		panic("WasmConvertor.ScAgentID kind")
	}
}

func (cvt WasmConvertor) ScBalances(allowance *isc.Allowance) *wasmlib.ScBalances {
	transfer := wasmlib.NewScTransferBaseTokens(allowance.Assets.BaseTokens)
	for _, token := range allowance.Assets.Tokens {
		tokenID := cvt.ScTokenID(&token.ID)
		transfer.Set(&tokenID, cvt.ScBigInt(token.Amount))
	}
	for _, nft := range allowance.NFTs {
		nftID := cvt.ScNftID(&nft) //nolint:gosec
		transfer.AddNFT(&nftID)
	}
	return &transfer.ScBalances
}

func (cvt WasmConvertor) ScBigInt(bigInt *big.Int) wasmtypes.ScBigInt {
	return wasmtypes.BigIntFromBytes(bigInt.Bytes())
}

func (cvt WasmConvertor) ScChainID(chainID *isc.ChainID) wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(chainID.Bytes())
}

func (cvt WasmConvertor) ScHash(hash hashing.HashValue) wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(hash.Bytes())
}

func (cvt WasmConvertor) ScHname(hname isc.Hname) wasmtypes.ScHname {
	return wasmtypes.ScHname(hname)
}

func (cvt WasmConvertor) ScNftID(nftID *iotago.NFTID) wasmtypes.ScNftID {
	return wasmtypes.NftIDFromBytes(nftID[:])
}

func (cvt WasmConvertor) ScRequestID(requestID isc.RequestID) wasmtypes.ScRequestID {
	return wasmtypes.RequestIDFromBytes(requestID.Bytes())
}

func (cvt WasmConvertor) ScTokenID(tokenID *iotago.NativeTokenID) wasmtypes.ScTokenID {
	return wasmtypes.TokenIDFromBytes(tokenID[:])
}

func (cvt WasmConvertor) ToBigInt(amount interface{}) *big.Int {
	switch it := amount.(type) {
	case wasmtypes.ScBigInt:
		return cvt.IscpBigInt(it)
	default:
		return util.ToBigInt(amount)
	}
}
