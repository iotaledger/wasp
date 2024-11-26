// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	"math/big"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func init() {
	if isc.ChainIDLength != 32 {
		panic("static check: ChainID length does not match bytes32 in ISCTypes.sol")
	}
}

// CoinBalance matches the struct definition in ISCTypes.sol
type CoinBalance struct {
	CoinType coin.Type
	Amount   coin.Value
}

// ISCAgentID matches the struct definition in ISCTypes.sol
type ISCAgentID struct {
	Data []byte
}

func WrapISCAgentID(a isc.AgentID) ISCAgentID {
	return ISCAgentID{Data: a.Bytes()}
}

func (a ISCAgentID) Unwrap() (isc.AgentID, error) {
	return isc.AgentIDFromBytes(a.Data)
}

// TokenIDFromIotaObjectID returns the uint256 tokenID for ERC721
func TokenIDFromIotaObjectID(o iotago.ObjectID) *big.Int {
	return new(big.Int).SetBytes(o[:])
}

// IRC27NFTMetadata matches the struct definition in ISCTypes.sol
type IRC27NFTMetadata struct {
	Standard    string
	Version     string
	MimeType    string
	Uri         string //nolint:revive // "URI" would break serialization
	Name        string
	Description string
}

func WrapIRC27NFTMetadata(m *isc.IRC27NFTMetadata) IRC27NFTMetadata {
	return IRC27NFTMetadata{
		Standard:    m.Standard,
		Version:     m.Version,
		MimeType:    m.MIMEType,
		Uri:         m.URI,
		Name:        m.Name,
		Description: m.Description,
	}
}

// ISCAssets matches the struct definition in ISCTypes.sol
type ISCAssets struct {
	Coins   []CoinBalance
	Objects []iotago.ObjectID
}

func WrapISCAssets(a *isc.Assets) ISCAssets {
	var ret ISCAssets
	a.Coins.IterateSorted(func(coinType coin.Type, amount coin.Value) bool {
		ret.Coins = append(ret.Coins, CoinBalance{
			CoinType: coinType,
			Amount:   amount,
		})
		return true
	})
	a.Objects.IterateSorted(func(id iotago.ObjectID) bool {
		ret.Objects = append(ret.Objects, id)
		return true
	})
	return ret
}

func (a ISCAssets) Unwrap() *isc.Assets {
	assets := isc.NewEmptyAssets()
	for _, b := range a.Coins {
		assets.AddCoin(b.CoinType, b.Amount)
	}
	for _, id := range a.Objects {
		assets.AddObject(id)
	}
	return assets
}

// ISCDictItem matches the struct definition in ISCTypes.sol
type ISCDictItem struct {
	Key   []byte
	Value []byte
}

// ISCDict matches the struct definition in ISCTypes.sol
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

type ISCCallTarget struct {
	Contract   isc.Hname
	EntryPoint isc.Hname
}

type ISCMessage struct {
	Target ISCCallTarget
	Params [][]byte
}

func WrapISCMessage(msg isc.Message) ISCMessage {
	return ISCMessage{
		Target: ISCCallTarget{
			Contract:   msg.Target.Contract,
			EntryPoint: msg.Target.EntryPoint,
		},
		Params: msg.Params,
	}
}

func (m ISCMessage) Unwrap() isc.Message {
	return isc.NewMessage(m.Target.Contract, m.Target.EntryPoint, m.Params)
}

type ISCSendMetadata struct {
	Message   ISCMessage
	Allowance ISCAssets
	GasBudget uint64
}

func WrapISCSendMetadata(metadata isc.SendMetadata) ISCSendMetadata {
	return ISCSendMetadata{
		Message:   WrapISCMessage(metadata.Message),
		Allowance: WrapISCAssets(metadata.Allowance),
		GasBudget: metadata.GasBudget,
	}
}

func (i ISCSendMetadata) Unwrap() *isc.SendMetadata {
	return &isc.SendMetadata{
		Message:   i.Message.Unwrap(),
		Allowance: i.Allowance.Unwrap(),
		GasBudget: i.GasBudget,
	}
}

func init() {
	if cryptolib.AddressSize != 32 {
		panic("static check: address length != 32")
	}
}

func WrapIotaAddress(addr *cryptolib.Address) (ret [32]byte) {
	return *addr
}
