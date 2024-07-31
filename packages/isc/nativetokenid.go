package isc

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func NativeTokenIDFromBytes(data []byte) (ret NativeTokenID, err error) {
	ret = NativeTokenID(data)
	return ret, nil
}

func MustNativeTokenIDFromBytes(data []byte) NativeTokenID {
	ret, err := NativeTokenIDFromBytes(data)
	if err != nil {
		panic(fmt.Errorf("MustNativeTokenIDFromBytes: %w", err))
	}
	return ret
}

func NativeTokenIDToBytes(tokenID NativeTokenID) []byte {
	return []byte(tokenID[:])
}

type NativeTokenID suijsonrpc.CoinType

func (n *NativeTokenID) Bytes() []byte {
	return NativeTokenIDToBytes(*n)
}

type NativeToken struct {
	CoinType NativeTokenID
	Amount   *big.Int
}

// NativeTokensSet is a set of NativeToken(s).
type NativeTokensSet map[NativeTokenID]*NativeToken

// NativeTokens is a set of NativeToken.
type NativeTokens []*NativeToken

func (ns *NativeTokens) MustSet() NativeTokensSet {
	m := make(NativeTokensSet)
	for _, n := range *ns {
		_n := n
		m[n.CoinType] = _n
	}
	return m
}

func (ns NativeTokens) Clone() *NativeTokens {
	ret := make(NativeTokens, len(ns))
	for i, n := range ns {
		ret[i] = n
	}
	return &ret
}

type NativeTokenSum map[NativeTokenID]*big.Int
