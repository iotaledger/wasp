package isctest

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
)

func RandomStateAnchor() isc.StateAnchor {
	anchor := iscmovetest.RandomAnchor()
	anchorRef := iscmove.RefWithObject[iscmove.Anchor]{
		ObjectRef: *iotatest.RandomObjectRef(),
		Object:    &anchor,
	}
	return isc.NewStateAnchor(&anchorRef, cryptolib.NewRandomAddress(), *iotatest.RandomAddress())
}

// RandomChainID creates a random chain ID. Used for testing only
func RandomChainID(seed ...[]byte) isc.ChainID {
	var h hashing.HashValue
	if len(seed) > 0 {
		h = hashing.HashData(seed[0])
	} else {
		h = hashing.PseudoRandomHash(nil)
	}
	chainID, err := isc.ChainIDFromBytes(h[:isc.ChainIDLength])
	if err != nil {
		panic(err)
	}
	return chainID
}

func RandomRequestWithRef() *iscmove.RefWithObject[iscmove.Request] {
	ref := iotatest.RandomObjectRef()
	a := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{ID: *iotatest.RandomAddress(), Size: 1},
		Balances:  make(iscmove.AssetsBagBalances),
	}
	a.Balances[iotajsonrpc.IotaCoinType] = &iotajsonrpc.Balance{CoinType: iotajsonrpc.IotaCoinType, TotalBalance: iotajsonrpc.NewBigInt(1000)}
	return &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *ref,
		Object: &iscmove.Request{
			ID:        *ref.ObjectID,
			Sender:    cryptolib.NewRandomAddress(),
			AssetsBag: a,
			Message: iscmove.Message{
				Contract: 123,
				Function: 456,
				Args:     [][]byte{[]byte("testarg1"), []byte("testarg2")},
			},
			Allowance: iscmove.Assets{Coins: iscmove.CoinBalances{iotajsonrpc.IotaCoinType: 111, "TEST_A": 222}},
			GasBudget: 1000,
		},
	}
}

func RandomOnLedgerRequest() isc.OnLedgerRequest {
	req := RandomRequestWithRef()
	onReq, err := isc.OnLedgerFromRequest(req, cryptolib.NewRandomAddress())
	if err != nil {
		panic(err)
	}
	return onReq
}

func RandomOnLedgerDepositRequest(senders ...*cryptolib.Address) isc.OnLedgerRequest {
	sender := cryptolib.NewRandomAddress()
	if len(senders) != 0 {
		sender = senders[0]
	}
	ref := iotatest.RandomObjectRef()
	a := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{ID: *iotatest.RandomAddress(), Size: 1},
		Balances:  make(iscmove.AssetsBagBalances),
	}
	a.Balances[iotajsonrpc.IotaCoinType] = &iotajsonrpc.Balance{CoinType: iotajsonrpc.IotaCoinType, TotalBalance: iotajsonrpc.NewBigInt(1000)}
	req := iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *ref,
		Object: &iscmove.Request{
			ID:        *ref.ObjectID,
			Sender:    sender,
			AssetsBag: a,
			Message: iscmove.Message{
				Contract: uint32(isc.Hn("accounts")),
				Function: uint32(isc.Hn("deposit")),
			},
			Allowance: iscmove.Assets{Coins: iscmove.CoinBalances{iotajsonrpc.IotaCoinType: 10000}},
			GasBudget: 100000,
		},
	}
	onReq, err := isc.OnLedgerFromRequest(&req, sender)
	if err != nil {
		panic(err)
	}
	return onReq
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() isc.AgentID {
	return isc.NewContractAgentID(RandomChainID(), isc.Hn("testName"))
}
