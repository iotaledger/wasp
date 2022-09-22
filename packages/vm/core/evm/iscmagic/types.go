// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// ISCChainID matches the type definition in ISC.sol
type ISCChainID [isc.ChainIDLength]byte

func init() {
	if isc.ChainIDLength != 32 {
		panic("static check: ChainID length does not match bytes32 in ISC.sol")
	}
}

func WrapISCChainID(c *isc.ChainID) (ret ISCChainID) {
	copy(ret[:], c.Bytes())
	return
}

func (c ISCChainID) Unwrap() (*isc.ChainID, error) {
	return isc.ChainIDFromBytes(c[:])
}

func (c ISCChainID) MustUnwrap() *isc.ChainID {
	ret, err := c.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// NativeTokenID matches the struct definition in ISC.sol
type NativeTokenID struct {
	Data []byte
}

func WrapNativeTokenID(id *iotago.NativeTokenID) NativeTokenID {
	return NativeTokenID{Data: id[:]}
}

func (a NativeTokenID) Unwrap() (ret iotago.NativeTokenID) {
	copy(ret[:], a.Data)
	return
}

// NativeToken matches the struct definition in ISC.sol
type NativeToken struct {
	ID     NativeTokenID
	Amount *big.Int
}

func WrapNativeToken(nt *iotago.NativeToken) NativeToken {
	return NativeToken{
		ID:     WrapNativeTokenID(&nt.ID),
		Amount: nt.Amount,
	}
}

func (nt NativeToken) Unwrap() *iotago.NativeToken {
	return &iotago.NativeToken{
		ID:     nt.ID.Unwrap(),
		Amount: nt.Amount,
	}
}

// L1Address matches the struct definition in ISC.sol
type L1Address struct {
	Data []byte
}

func WrapL1Address(a iotago.Address) L1Address {
	if a == nil {
		return L1Address{Data: []byte{}}
	}
	return L1Address{Data: isc.BytesFromAddress(a)}
}

func (a L1Address) Unwrap() (iotago.Address, error) {
	ret, _, err := isc.AddressFromBytes(a.Data)
	return ret, err
}

func (a L1Address) MustUnwrap() iotago.Address {
	ret, err := a.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// ISCAgentID matches the struct definition in ISC.sol
type ISCAgentID struct {
	Data []byte
}

func WrapISCAgentID(a isc.AgentID) ISCAgentID {
	return ISCAgentID{Data: a.Bytes()}
}

func (a ISCAgentID) Unwrap() (isc.AgentID, error) {
	return isc.AgentIDFromBytes(a.Data)
}

func (a ISCAgentID) MustUnwrap() isc.AgentID {
	ret, err := a.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// NFTID matches the type definition in ISC.sol
type NFTID [iotago.NFTIDLength]byte

func init() {
	if iotago.NFTIDLength != 32 {
		panic("static check: NFTID length does not match bytes32 in ISC.sol")
	}
}

func WrapNFTID(c iotago.NFTID) (ret NFTID) {
	copy(ret[:], c[:])
	return
}

func (c NFTID) Unwrap() (ret iotago.NFTID) {
	copy(ret[:], c[:])
	return
}

// ISCNFT matches the struct definition in ISC.sol
type ISCNFT struct {
	ID       NFTID
	Issuer   L1Address
	Metadata []byte
}

func WrapISCNFT(n *isc.NFT) ISCNFT {
	return ISCNFT{
		ID:       WrapNFTID(n.ID),
		Issuer:   WrapL1Address(n.Issuer),
		Metadata: n.Metadata,
	}
}

func (n ISCNFT) Unwrap() (*isc.NFT, error) {
	issuer, err := n.Issuer.Unwrap()
	if err != nil {
		return nil, err
	}
	return &isc.NFT{
		ID:       n.ID.Unwrap(),
		Issuer:   issuer,
		Metadata: n.Metadata,
	}, nil
}

func (n ISCNFT) MustUnwrap() *isc.NFT {
	ret, err := n.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// ISCAllowance matches the struct definition in ISC.sol
type ISCAllowance struct {
	BaseTokens uint64
	Tokens     []NativeToken
	Nfts       []NFTID
}

func WrapISCAllowance(a *isc.Allowance) ISCAllowance {
	if a == nil {
		return WrapISCAllowance(isc.NewEmptyAllowance())
	}
	tokens := make([]NativeToken, len(a.Assets.Tokens))
	for i, t := range a.Assets.Tokens {
		tokens[i] = WrapNativeToken(t)
	}
	nfts := make([]NFTID, len(a.NFTs))
	for i, id := range a.NFTs {
		nfts[i] = WrapNFTID(id)
	}
	return ISCAllowance{
		BaseTokens: a.Assets.BaseTokens,
		Tokens:     tokens,
		Nfts:       nfts,
	}
}

func (a ISCAllowance) Unwrap() *isc.Allowance {
	tokens := make(iotago.NativeTokens, len(a.Tokens))
	for i, t := range a.Tokens {
		tokens[i] = t.Unwrap()
	}
	nfts := make([]iotago.NFTID, len(a.Nfts))
	for i, id := range a.Nfts {
		nfts[i] = id.Unwrap()
	}
	return isc.NewAllowance(a.BaseTokens, tokens, nfts)
}

// ISCDictItem matches the struct definition in ISC.sol
type ISCDictItem struct {
	Key   []byte
	Value []byte
}

// ISCDict matches the struct definition in ISC.sol
type ISCDict struct {
	Items []ISCDictItem
}

func WrapISCDict(d dict.Dict) ISCDict {
	items := make([]ISCDictItem, 0, len(d))
	for k, v := range d {
		items = append(items, ISCDictItem{Key: []byte(k), Value: v})
	}
	return ISCDict{Items: items}
}

func (d ISCDict) Unwrap() dict.Dict {
	ret := dict.Dict{}
	for _, item := range d.Items {
		ret[kv.Key(item.Key)] = item.Value
	}
	return ret
}

type ISCFungibleTokens struct {
	BaseTokens uint64
	Tokens     []NativeToken
}

func WrapISCFungibleTokens(fungibleTokens isc.FungibleTokens) ISCFungibleTokens {
	ret := ISCFungibleTokens{
		BaseTokens: fungibleTokens.BaseTokens,
		Tokens:     make([]NativeToken, len(fungibleTokens.Tokens)),
	}

	for i, v := range fungibleTokens.Tokens {
		ret.Tokens[i].ID = WrapNativeTokenID(&v.ID)
		ret.Tokens[i].Amount = v.Amount
	}

	return ret
}

func (t ISCFungibleTokens) Unwrap() *isc.FungibleTokens {
	ret := isc.FungibleTokens{
		BaseTokens: t.BaseTokens,
		Tokens:     make(iotago.NativeTokens, len(t.Tokens)),
	}

	for i, v := range t.Tokens {
		nativeToken := iotago.NativeToken{
			ID:     v.ID.Unwrap(),
			Amount: v.Amount,
		}

		ret.Tokens[i] = &nativeToken
	}

	return &ret
}

type ISCSendMetadata struct {
	TargetContract uint32
	Entrypoint     uint32
	Params         ISCDict
	Allowance      ISCAllowance
	GasBudget      uint64
}

func WrapISCSendMetadata(metadata isc.SendMetadata) ISCSendMetadata {
	ret := ISCSendMetadata{
		GasBudget:      metadata.GasBudget,
		Entrypoint:     uint32(metadata.EntryPoint),
		TargetContract: uint32(metadata.TargetContract),
		Allowance:      WrapISCAllowance(metadata.Allowance),
		Params:         WrapISCDict(metadata.Params),
	}

	return ret
}

func (i ISCSendMetadata) Unwrap() *isc.SendMetadata {
	ret := isc.SendMetadata{
		TargetContract: isc.Hname(i.TargetContract),
		EntryPoint:     isc.Hname(i.Entrypoint),
		Params:         i.Params.Unwrap(),
		Allowance:      i.Allowance.Unwrap(),
		GasBudget:      i.GasBudget,
	}

	return &ret
}

type ISCExpiration struct {
	Time          int64
	ReturnAddress L1Address
}

func WrapISCExpiration(data *isc.Expiration) ISCExpiration {
	if data == nil {
		return ISCExpiration{
			Time:          0,
			ReturnAddress: WrapL1Address(nil),
		}
	}
	var expiryTime int64

	if !data.Time.IsZero() {
		expiryTime = data.Time.UnixMilli()
	}

	ret := ISCExpiration{
		Time:          expiryTime,
		ReturnAddress: WrapL1Address(data.ReturnAddress),
	}

	return ret
}

func (i *ISCExpiration) Unwrap() *isc.Expiration {
	if i == nil {
		return nil
	}

	if i.Time == 0 {
		return nil
	}

	address := i.ReturnAddress.MustUnwrap()

	ret := isc.Expiration{
		ReturnAddress: address,
		Time:          time.UnixMilli(i.Time),
	}

	return &ret
}

type ISCSendOptions struct {
	Timelock   int64
	Expiration ISCExpiration
}

func WrapISCSendOptions(options isc.SendOptions) ISCSendOptions {
	var timeLock int64

	if !options.Timelock.IsZero() {
		timeLock = options.Timelock.UnixMilli()
	}

	ret := ISCSendOptions{
		Timelock:   timeLock,
		Expiration: WrapISCExpiration(options.Expiration),
	}

	return ret
}

func (i *ISCSendOptions) Unwrap() isc.SendOptions {
	var timeLock time.Time

	if i.Timelock > 0 {
		timeLock = time.UnixMilli(i.Timelock)
	}

	ret := isc.SendOptions{
		Timelock:   timeLock,
		Expiration: i.Expiration.Unwrap(),
	}

	return ret
}

type ISCTokenProperties struct {
	Name         string
	TickerSymbol string
	Decimals     uint8
	TotalSupply  *big.Int
}
