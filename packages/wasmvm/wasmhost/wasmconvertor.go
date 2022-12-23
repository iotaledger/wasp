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

// WasmConvertor converts ISC data types to WasmLib data types
type WasmConvertor struct{}

func (cvt WasmConvertor) IscAddress(scAddress *wasmtypes.ScAddress) iotago.Address {
	buf := wasmtypes.AddressToBytes(*scAddress)
	address, _, err := isc.AddressFromBytes(buf)
	if err != nil {
		panic(err)
	}
	return address
}

func (cvt WasmConvertor) IscAgentID(scAgentID *wasmtypes.ScAgentID) isc.AgentID {
	buf := scAgentID.Bytes()
	agentID, err := isc.AgentIDFromBytes(buf)
	if err != nil {
		panic(err)
	}
	return agentID
}

func (cvt WasmConvertor) IscAllowance(assets *wasmlib.ScAssets) *isc.Allowance {
	iscAllowance := isc.NewAllowanceBaseTokens(assets.BaseTokens)
	iscAssets := iscAllowance.Assets
	for tokenID, amount := range assets.Tokens {
		token := &iotago.NativeToken{
			ID:     *cvt.IscTokenID(&tokenID),
			Amount: cvt.IscBigInt(amount),
		}
		iscAssets.Tokens = append(iscAssets.Tokens, token)
	}
	for nftID := range assets.NftIDs {
		nft := cvt.IscNFTID(&nftID)
		iscAllowance.NFTs = append(iscAllowance.NFTs, *nft)
	}
	return iscAllowance
}

func (cvt WasmConvertor) IscBigInt(amount wasmtypes.ScBigInt) *big.Int {
	res := new(big.Int)
	res.SetBytes(wasmtypes.BigIntToBytes(amount))
	return res
}

func (cvt WasmConvertor) IscChainID(chainID *wasmtypes.ScChainID) isc.ChainID {
	buf := wasmtypes.ChainIDToBytes(*chainID)
	iscChainID := isc.ChainID{}
	copy(iscChainID[:], buf)
	return iscChainID
}

func (cvt WasmConvertor) IscHash(hash *wasmtypes.ScHash) *hashing.HashValue {
	buf := wasmtypes.HashToBytes(*hash)
	iscHashValue := new(hashing.HashValue)
	copy(iscHashValue[:], buf)
	return iscHashValue
}

func (cvt WasmConvertor) IscHname(hname wasmtypes.ScHname) isc.Hname {
	return isc.Hname(hname)
}

func (cvt WasmConvertor) IscNFTID(nftID *wasmtypes.ScNftID) *iotago.NFTID {
	buf := wasmtypes.NftIDToBytes(*nftID)
	iscNFTID := new(iotago.NFTID)
	copy(iscNFTID[:], buf)
	return iscNFTID
}

func (cvt WasmConvertor) IscRequestID(requestID *wasmtypes.ScRequestID) *isc.RequestID {
	buf := wasmtypes.RequestIDToBytes(*requestID)
	iscRequestID := new(isc.RequestID)
	copy(iscRequestID[:], buf)
	return iscRequestID
}

func (cvt WasmConvertor) IscTokenID(tokenID *wasmtypes.ScTokenID) *iotago.NativeTokenID {
	buf := wasmtypes.TokenIDToBytes(*tokenID)
	iscTokenID := new(iotago.NativeTokenID)
	copy(iscTokenID[:], buf)
	return iscTokenID
}

func (cvt WasmConvertor) ScAddress(address iotago.Address) wasmtypes.ScAddress {
	buf := isc.BytesFromAddress(address)
	return wasmtypes.AddressFromBytes(buf)
}

func (cvt WasmConvertor) ScAgentID(agentID isc.AgentID) wasmtypes.ScAgentID {
	buf := agentID.Bytes()
	return wasmtypes.AgentIDFromBytes(buf)
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

func (cvt WasmConvertor) ScChainID(chainID isc.ChainID) wasmtypes.ScChainID {
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
		return cvt.IscBigInt(it)
	default:
		return util.ToBigInt(amount)
	}
}
