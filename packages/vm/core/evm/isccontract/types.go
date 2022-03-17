// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isccontract

import (
	"math/big"

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
