// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isccontract

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"math/big"
	"time"
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
	IotaAddress IotaAddress
	Hname       uint32
}

func WrapISCAgentID(a *iscp.AgentID) ISCAgentID {
	return ISCAgentID{
		IotaAddress: WrapIotaAddress(a.Address()),
		Hname:       uint32(a.Hname()),
	}
}

func (a ISCAgentID) Unwrap() (*iscp.AgentID, error) {
	addr, err := a.IotaAddress.Unwrap()
	if err != nil {
		return nil, err
	}
	return iscp.NewAgentID(addr, iscp.Hname(a.Hname)), nil
}

func (a ISCAgentID) MustUnwrap() *iscp.AgentID {
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

func WrapISCNFTID(c iotago.NFTID) (ret IotaNFTID) {
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
		ID:       WrapISCNFTID(n.ID),
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

func (a ISCNFT) MustUnwrap() *iscp.NFT {
	ret, err := a.Unwrap()
	if err != nil {
		panic(err)
	}
	return ret
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

type IotaFungibleTokens struct {
	Iotas  uint64
	Tokens []IotaNativeToken
}

func WrapIotaFungibleTokens(fungibleTokens iscp.FungibleTokens) IotaFungibleTokens {
	ret := IotaFungibleTokens{
		Iotas:  fungibleTokens.Iotas,
		Tokens: make([]IotaNativeToken, len(fungibleTokens.Tokens)),
	}

	for i, v := range fungibleTokens.Tokens {
		ret.Tokens[i].ID = WrapIotaNativeTokenID(&v.ID)
		ret.Tokens[i].Amount = v.Amount
	}

	return ret
}

func (t IotaFungibleTokens) Unwrap() *iscp.FungibleTokens {
	ret := iscp.FungibleTokens{
		Iotas: t.Iotas,
	}

	for i, v := range t.Tokens {
		ret.Tokens[i].ID = v.ID.Unwrap()
		ret.Tokens[i].Amount = v.Amount
	}

	return &ret
}

type IotaAllowance struct {
	Assets IotaFungibleTokens
	NFTs   []IotaNFTID
}

func WrapIotaAllowance(allowance iscp.Allowance) IotaAllowance {
	nftIds := make([]IotaNFTID, 0)

	for _, nft := range allowance.NFTs {
		nftIds = append(nftIds, WrapISCNFTID(nft))
	}

	ret := IotaAllowance{
		NFTs:   nftIds,
		Assets: WrapIotaFungibleTokens(*allowance.Assets),
	}

	return ret
}

func (a IotaAllowance) Unwrap() *iscp.Allowance {
	nftIDs := make([]iotago.NFTID, 0)

	for _, nftID := range a.NFTs {
		nftIDs = append(nftIDs, nftID.Unwrap())
	}

	ret := iscp.Allowance{
		Assets: a.Assets.Unwrap(),
		NFTs:   nftIDs,
	}

	return &ret
}

type ISCSendMetadata struct {
	TargetContract iscp.Hname
	Entrypoint     iscp.Hname
	// TODO: Params
	Allowance IotaAllowance
	GasBudget uint64
}

func WrapISCSendMetadata(metadata iscp.SendMetadata) ISCSendMetadata {
	ret := ISCSendMetadata{
		GasBudget:      metadata.GasBudget,
		Entrypoint:     metadata.EntryPoint,
		TargetContract: metadata.TargetContract,
		Allowance:      WrapIotaAllowance(*metadata.Allowance),
	}

	return ret
}

func (i ISCSendMetadata) Unwrap() *iscp.SendMetadata {
	ret := iscp.SendMetadata{
		TargetContract: i.TargetContract,
		EntryPoint:     i.Entrypoint,
		Params:         nil,
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
	ret := iscp.TimeData{
		MilestoneIndex: i.MilestoneIndex,
		Time:           time.UnixMilli(i.Time),
	}

	return &ret
}

type ISCExpiration struct {
	ISCTimeData
	ReturnAddress IotaAddress
}

func WrapISCExpiration(data *iscp.Expiration) ISCExpiration {
	ret := ISCExpiration{
		ISCTimeData: ISCTimeData{
			MilestoneIndex: data.MilestoneIndex,
			Time:           data.Time.UnixMilli(),
		},
		ReturnAddress: WrapIotaAddress(data.ReturnAddress),
	}

	return ret
}

func (i ISCExpiration) Unwrap() *iscp.Expiration {
	ret := iscp.Expiration{
		ReturnAddress: i.ReturnAddress.MustUnwrap(),
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
	FungibleTokens           IotaFungibleTokens
	AdjustMinimumDustDeposit bool
	Metadata                 ISCSendMetadata
	SendOptions              ISCSendOptions
}

func WrapISCRequestParameters(parameters iscp.RequestParameters) ISCRequestParameters {
	ret := ISCRequestParameters{
		TargetAddress:            WrapIotaAddress(parameters.TargetAddress),
		FungibleTokens:           WrapIotaFungibleTokens(*parameters.FungibleTokens),
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
