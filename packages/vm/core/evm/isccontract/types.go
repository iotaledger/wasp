// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isccontract

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// ISCChainID matches the type definition in ISC.sol
type ISCChainID [iscp.ChainIDLength]byte

func init() {
	if iscp.ChainIDLength != 20 {
		panic("static check: ChainID length does not match bytes20 in ISC.sol")
	}
}

func WrapISCChainID(c *iscp.ChainID) (ret ISCChainID) {
	copy(ret[:], c.Bytes())
	return
}

func (c ISCChainID) Unwrap() (*iscp.ChainID, error) {
	return iscp.ChainIDFromBytes(c[:])
}

func (c ISCChainID) MustUnwrap() *iscp.ChainID {
	ret, err := c.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// IotaNativeTokenID matches the struct definition in ISC.sol
type IotaNativeTokenID struct {
	Data []byte
}

func WrapIotaNativeTokenID(id *iotago.NativeTokenID) IotaNativeTokenID {
	return IotaNativeTokenID{Data: id[:]}
}

func (a IotaNativeTokenID) Unwrap() (ret iotago.NativeTokenID) {
	copy(ret[:], a.Data)
	return
}

// IotaNativeToken matches the struct definition in ISC.sol
type IotaNativeToken struct {
	ID     IotaNativeTokenID
	Amount *big.Int
}

func WrapIotaNativeToken(nt *iotago.NativeToken) IotaNativeToken {
	return IotaNativeToken{
		ID:     WrapIotaNativeTokenID(&nt.ID),
		Amount: nt.Amount,
	}
}

func (nt IotaNativeToken) Unwrap() *iotago.NativeToken {
	return &iotago.NativeToken{
		ID:     nt.ID.Unwrap(),
		Amount: nt.Amount,
	}
}

// IotaAddress matches the struct definition in ISC.sol
type IotaAddress struct {
	Data []byte
}

func WrapIotaAddress(a iotago.Address) IotaAddress {
	return IotaAddress{Data: iscp.BytesFromAddress(a)}
}

func (a IotaAddress) Unwrap() (iotago.Address, error) {
	ret, _, err := iscp.AddressFromBytes(a.Data)
	return ret, err
}

func (a IotaAddress) MustUnwrap() iotago.Address {
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

func WrapISCAgentID(a iscp.AgentID) ISCAgentID {
	return ISCAgentID{Data: a.Bytes()}
}

func (a ISCAgentID) Unwrap() (iscp.AgentID, error) {
	return iscp.AgentIDFromBytes(a.Data)
}

func (a ISCAgentID) MustUnwrap() iscp.AgentID {
	ret, err := a.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// IotaNFTID matches the type definition in ISC.sol
type IotaNFTID [iotago.NFTIDLength]byte

func init() {
	if iotago.NFTIDLength != 20 {
		panic("static check: NFTID length does not match bytes20 in ISC.sol")
	}
}

func WrapIotaNFTID(c iotago.NFTID) (ret IotaNFTID) {
	copy(ret[:], c[:])
	return
}

func (c IotaNFTID) Unwrap() (ret iotago.NFTID) {
	copy(ret[:], c[:])
	return
}

// ISCNFT matches the struct definition in ISC.sol
type ISCNFT struct {
	ID       IotaNFTID
	Issuer   IotaAddress
	Metadata []byte
}

func WrapISCNFT(n *iscp.NFT) ISCNFT {
	return ISCNFT{
		ID:       WrapIotaNFTID(n.ID),
		Issuer:   WrapIotaAddress(n.Issuer),
		Metadata: n.Metadata,
	}
}

func (n ISCNFT) Unwrap() (*iscp.NFT, error) {
	issuer, err := n.Issuer.Unwrap()
	if err != nil {
		return nil, err
	}
	return &iscp.NFT{
		ID:       n.ID.Unwrap(),
		Issuer:   issuer,
		Metadata: n.Metadata,
	}, nil
}

func (n ISCNFT) MustUnwrap() *iscp.NFT {
	ret, err := n.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
}

// ISCAllowance matches the struct definition in ISC.sol
type ISCAllowance struct {
	Iotas  uint64
	Tokens []IotaNativeToken
	NFTs   []IotaNFTID
}

func WrapISCAllowance(a *iscp.Allowance) ISCAllowance {
	tokens := make([]IotaNativeToken, len(a.Assets.Tokens))
	for i, t := range a.Assets.Tokens {
		tokens[i] = WrapIotaNativeToken(t)
	}
	nfts := make([]IotaNFTID, len(a.NFTs))
	for i, id := range a.NFTs {
		nfts[i] = WrapIotaNFTID(id)
	}
	return ISCAllowance{
		Iotas:  a.Assets.Iotas,
		Tokens: tokens,
		NFTs:   nfts,
	}
}

func (a ISCAllowance) Unwrap() *iscp.Allowance {
	tokens := make(iotago.NativeTokens, len(a.Tokens))
	for i, t := range a.Tokens {
		tokens[i] = t.Unwrap()
	}
	nfts := make([]iotago.NFTID, len(a.NFTs))
	for i, id := range a.NFTs {
		nfts[i] = id.Unwrap()
	}
	return iscp.NewAllowance(a.Iotas, tokens, nfts)
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
	Iotas  uint64
	Tokens []IotaNativeToken
}

func WrapISCFungibleTokens(fungibleTokens iscp.FungibleTokens) ISCFungibleTokens {
	ret := ISCFungibleTokens{
		Iotas:  fungibleTokens.Iotas,
		Tokens: make([]IotaNativeToken, len(fungibleTokens.Tokens)),
	}

	for i, v := range fungibleTokens.Tokens {
		ret.Tokens[i].ID = WrapIotaNativeTokenID(&v.ID)
		ret.Tokens[i].Amount = v.Amount
	}

	return ret
}

func (t ISCFungibleTokens) Unwrap() *iscp.FungibleTokens {
	ret := iscp.FungibleTokens{
		Iotas:  t.Iotas,
		Tokens: make(iotago.NativeTokens, len(t.Tokens)),
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

func WrapISCSendMetadata(metadata iscp.SendMetadata) ISCSendMetadata {
	ret := ISCSendMetadata{
		GasBudget:      metadata.GasBudget,
		Entrypoint:     uint32(metadata.EntryPoint),
		TargetContract: uint32(metadata.TargetContract),
		Allowance:      WrapISCAllowance(metadata.Allowance),
		Params:         WrapISCDict(metadata.Params),
	}

	return ret
}

func (i ISCSendMetadata) Unwrap() *iscp.SendMetadata {
	ret := iscp.SendMetadata{
		TargetContract: iscp.Hname(i.TargetContract),
		EntryPoint:     iscp.Hname(i.Entrypoint),
		Params:         i.Params.Unwrap(),
		Allowance:      i.Allowance.Unwrap(),
		GasBudget:      i.GasBudget,
	}

	return &ret
}

type ISCTimeData struct {
	MilestoneIndex uint32
	Time           int64
}

func WrapISCTimeData(data *iscp.TimeData) ISCTimeData {
	ret := ISCTimeData{
		MilestoneIndex: data.MilestoneIndex,
		Time:           data.Time.UnixMilli(),
	}

	return ret
}

func (i ISCTimeData) Unwrap() *iscp.TimeData {
	if i.MilestoneIndex == 0 && i.Time == 0 {
		return nil
	}

	ret := iscp.TimeData{
		MilestoneIndex: i.MilestoneIndex,
		Time:           time.UnixMilli(i.Time),
	}

	return &ret
}

type ISCExpiration struct {
	MilestoneIndex uint32
	Time           int64
	ReturnAddress  IotaAddress
}

func WrapISCExpiration(data *iscp.Expiration) ISCExpiration {
	ret := ISCExpiration{
		MilestoneIndex: data.MilestoneIndex,
		Time:           data.Time.UnixMilli(),
		ReturnAddress:  WrapIotaAddress(data.ReturnAddress),
	}

	return ret
}

func (i *ISCExpiration) Unwrap() *iscp.Expiration {
	if i == nil {
		return nil
	}

	if i.MilestoneIndex == 0 && i.Time == 0 {
		return nil
	}

	address := i.ReturnAddress.MustUnwrap()

	ret := iscp.Expiration{
		ReturnAddress: address,
		TimeData: iscp.TimeData{
			MilestoneIndex: i.MilestoneIndex,
			Time:           time.UnixMilli(i.Time),
		},
	}

	return &ret
}

type ISCSendOptions struct {
	Timelock   ISCTimeData
	Expiration ISCExpiration
}

func WrapISCSendOptions(options iscp.SendOptions) ISCSendOptions {
	ret := ISCSendOptions{
		Timelock:   WrapISCTimeData(options.Timelock),
		Expiration: WrapISCExpiration(options.Expiration),
	}

	return ret
}

func (i *ISCSendOptions) Unwrap() iscp.SendOptions {
	ret := iscp.SendOptions{
		Timelock:   i.Timelock.Unwrap(),
		Expiration: i.Expiration.Unwrap(),
	}

	return ret
}

type ISCRequestParameters struct {
	TargetAddress            IotaAddress
	FungibleTokens           ISCFungibleTokens
	AdjustMinimumDustDeposit bool
	Metadata                 ISCSendMetadata
	SendOptions              ISCSendOptions
}

func WrapISCRequestParameters(parameters iscp.RequestParameters) ISCRequestParameters {
	ret := ISCRequestParameters{
		TargetAddress:            WrapIotaAddress(parameters.TargetAddress),
		FungibleTokens:           WrapISCFungibleTokens(*parameters.FungibleTokens),
		AdjustMinimumDustDeposit: parameters.AdjustToMinimumDustDeposit,
		Metadata:                 WrapISCSendMetadata(*parameters.Metadata),
		SendOptions:              WrapISCSendOptions(parameters.Options),
	}

	return ret
}

func (i *ISCRequestParameters) Unwrap() iscp.RequestParameters {
	ret := iscp.RequestParameters{
		TargetAddress:              i.TargetAddress.MustUnwrap(),
		FungibleTokens:             i.FungibleTokens.Unwrap(),
		AdjustToMinimumDustDeposit: i.AdjustMinimumDustDeposit,
		Metadata:                   i.Metadata.Unwrap(),
		Options:                    i.SendOptions.Unwrap(),
	}

	return ret
}
