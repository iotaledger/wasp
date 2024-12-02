// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	"math/big"
	"time"

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

type (
	CoinType  = string
	CoinValue = uint64
)

// CoinBalance matches the struct definition in ISCTypes.sol
type CoinBalance struct {
	CoinType CoinType
	Amount   CoinValue
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
			CoinType: CoinType(coinType.String()),
			Amount:   CoinValue(amount),
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
		assets.AddCoin(coin.MustTypeFromString(string(b.CoinType)), coin.Value(b.Amount))
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

type ISCHname = uint32

type ISCCallTarget struct {
	ContractHname ISCHname
	EntryPoint    ISCHname
}

type ISCMessage struct {
	Target ISCCallTarget
	Params [][]byte
}

func WrapISCMessage(msg isc.Message) ISCMessage {
	return ISCMessage{
		Target: ISCCallTarget{
			ContractHname: ISCHname(msg.Target.Contract),
			EntryPoint:    ISCHname(msg.Target.EntryPoint),
		},
		Params: msg.Params,
	}
}

func (m ISCMessage) Unwrap() isc.Message {
	return isc.NewMessage(isc.Hname(m.Target.ContractHname), isc.Hname(m.Target.EntryPoint), m.Params)
}

type ISCSendMetadata struct {
	Message   ISCMessage
	Allowance ISCAssets
	GasBudget uint64
}

func WrapISCSendMetadata(metadata *isc.SendMetadata) ISCSendMetadata {
	if metadata == nil {
		return ISCSendMetadata{}
	}
	return ISCSendMetadata{
		Message:   WrapISCMessage(metadata.Message),
		Allowance: WrapISCAssets(metadata.Allowance),
		GasBudget: metadata.GasBudget,
	}
}

func (i ISCSendMetadata) Unwrap() *isc.SendMetadata {
	if i.Message.Target.ContractHname == 0 {
		return nil
	}
	return &isc.SendMetadata{
		Message:   i.Message.Unwrap(),
		Allowance: i.Allowance.Unwrap(),
		GasBudget: i.GasBudget,
	}
}

type ISCSendOptions struct {
	Timelock   int64
	Expiration struct {
		Time          int64
		ReturnAddress cryptolib.Address
	}
}

func WrapISCSendOptions(options *isc.SendOptions) ISCSendOptions {
	var ret ISCSendOptions
	if options == nil {
		return ret
	}
	ret.Timelock = options.Timelock.Unix()
	if options.Expiration == nil {
		return ret
	}
	ret.Expiration.Time = options.Expiration.Time.Unix()
	ret.Expiration.ReturnAddress = *options.Expiration.ReturnAddress
	return ret
}

func (i *ISCSendOptions) Unwrap() isc.SendOptions {
	var timeLock time.Time
	if i.Timelock > 0 {
		timeLock = time.Unix(i.Timelock, 0)
	}
	ret := isc.SendOptions{
		Timelock: timeLock,
	}
	if i.Expiration.Time > 0 {
		ret.Expiration = &isc.Expiration{
			Time:          time.Unix(i.Expiration.Time, 0),
			ReturnAddress: &i.Expiration.ReturnAddress,
		}
	}
	return ret
}

func init() {
	if cryptolib.AddressSize != 32 {
		panic("static check: address length != 32")
	}
}

func WrapIotaAddress(addr *cryptolib.Address) (ret [32]byte) {
	return *addr
}
