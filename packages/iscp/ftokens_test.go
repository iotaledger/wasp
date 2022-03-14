package iscp

import (
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
)

func TestMarshalling(t *testing.T) {
	maxVal, e := big.NewInt(2), big.NewInt(256)
	maxVal.Exp(maxVal, e, nil)
	maxVal.Sub(maxVal, big.NewInt(1))

	tokens := iotago.NativeTokens{
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{1},
			Amount: big.NewInt(100),
		},
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{2},
			Amount: big.NewInt(200),
		},
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{3},
			Amount: util.MaxUint256,
		},
	}

	assets := NewFungibleTokens(1, tokens)
	bytes := assets.Bytes()
	assets2, err := FungibleTokensFromMarshalUtil(marshalutil.New(bytes))
	require.NoError(t, err)
	require.Equal(t, assets.Iotas, assets2.Iotas)
	require.Equal(t, len(assets.Tokens), len(assets2.Tokens))
	for i := range tokens {
		require.Equal(t, assets.Tokens[i], assets2.Tokens[i])
	}
}

func TestAssets_SpendBudget(t *testing.T) {
	var toSpend *FungibleTokens
	var budget *FungibleTokens
	require.True(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.True(t, budget.IsEmpty())
	require.True(t, budget.IsEmpty())

	budget = &FungibleTokens{Iotas: 1}
	require.True(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.False(t, toSpend.SpendFromFungibleTokenBudget(budget))

	budget = &FungibleTokens{Iotas: 10}
	require.True(t, budget.SpendFromFungibleTokenBudget(budget))
	require.True(t, budget.IsEmpty())

	budget = &FungibleTokens{Iotas: 2}
	toSpend = &FungibleTokens{Iotas: 1}
	require.True(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.True(t, budget.Equals(&FungibleTokens{1, nil}))

	budget = &FungibleTokens{Iotas: 1}
	toSpend = &FungibleTokens{Iotas: 2}
	require.False(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.True(t, budget.Equals(&FungibleTokens{1, nil}))

	tokenID1 := tpkg.RandNativeToken().ID
	tokenID2 := tpkg.RandNativeToken().ID

	budget = &FungibleTokens{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	toSpend = budget.Clone()
	require.True(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.True(t, budget.IsEmpty())

	budget = &FungibleTokens{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	cloneBudget := budget.Clone()
	toSpend = &FungibleTokens{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(10)},
		},
	}
	require.False(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.True(t, budget.Equals(cloneBudget))

	budget = &FungibleTokens{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	toSpend = &FungibleTokens{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	expected := &FungibleTokens{
		Iotas: 0,
		Tokens: iotago.NativeTokens{
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	require.True(t, budget.SpendFromFungibleTokenBudget(toSpend))
	require.True(t, budget.Equals(expected))

	budget = &FungibleTokens{
		Iotas: 10,
		Tokens: iotago.NativeTokens{
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	toSpend = &FungibleTokens{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}

	require.False(t, budget.SpendFromFungibleTokenBudget(toSpend))
}
