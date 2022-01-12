package iscp

import (
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
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
			Amount: maxVal,
		},
	}

	assets := NewAssets(1, tokens)
	bytes := assets.Bytes()
	assets2, err := AssetsFromMarshalUtil(marshalutil.New(bytes))
	require.NoError(t, err)
	require.Equal(t, assets.Iotas, assets2.Iotas)
	require.Equal(t, len(assets.Tokens), len(assets2.Tokens))
	for i := range tokens {
		require.Equal(t, assets.Tokens[i], assets2.Tokens[i])
	}
}

func TestAssets_FitsAllowance(t *testing.T) {
	var a *Assets
	var allowance *Assets
	require.True(t, a.MustFitsTheBudget(allowance))

	allowance = &Assets{Iotas: 1}
	require.True(t, a.MustFitsTheBudget(allowance))
	require.False(t, allowance.MustFitsTheBudget(a))

	a = &Assets{Iotas: 1}
	require.True(t, a.MustFitsTheBudget(allowance))
	a = &Assets{Iotas: 2}
	require.False(t, a.MustFitsTheBudget(allowance))
	tokenID1 := tpkg.RandNativeToken().ID
	tokenID2 := tpkg.RandNativeToken().ID

	a = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	allowance = a
	require.True(t, a.MustFitsTheBudget(allowance))

	a = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	allowance = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	require.True(t, a.MustFitsTheBudget(allowance))
	require.False(t, allowance.MustFitsTheBudget(a))

	a = &Assets{
		Iotas: 2,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	allowance = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	require.False(t, a.MustFitsTheBudget(allowance))

	a = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	allowance = &Assets{
		Iotas: 10,
		Tokens: iotago.NativeTokens{
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}

	require.False(t, a.MustFitsTheBudget(allowance))
	a = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	allowance = &Assets{
		Iotas: 10,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(1)},
		},
	}
	require.False(t, a.MustFitsTheBudget(allowance))

	a = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	allowance = &Assets{
		Iotas: 10,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(1)},
			{ID: tokenID1, Amount: big.NewInt(10)},
		},
	}
	_, err := a.FitsTheBudget(allowance)
	require.Error(t, err)

	a = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(0)},
		},
	}
	allowance = &Assets{
		Iotas: 10,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(1)},
		},
	}
	_, err = a.FitsTheBudget(allowance)
	require.Error(t, err)
}

func TestAssets_SpendBudget(t *testing.T) {
	var toSpend *Assets
	var budget *Assets
	require.True(t, budget.SpendFromBudget(toSpend))
	require.True(t, budget.IsEmpty())
	require.True(t, budget.IsEmpty())

	budget = &Assets{Iotas: 1}
	require.True(t, budget.SpendFromBudget(toSpend))
	require.False(t, toSpend.SpendFromBudget(budget))

	budget = &Assets{Iotas: 10}
	require.True(t, budget.SpendFromBudget(budget))
	require.True(t, budget.IsEmpty())

	budget = &Assets{Iotas: 2}
	toSpend = &Assets{Iotas: 1}
	require.True(t, budget.SpendFromBudget(toSpend))
	require.True(t, budget.Equals(&Assets{1, nil}))

	budget = &Assets{Iotas: 1}
	toSpend = &Assets{Iotas: 2}
	require.False(t, budget.SpendFromBudget(toSpend))
	require.True(t, budget.Equals(&Assets{1, nil}))

	tokenID1 := tpkg.RandNativeToken().ID
	tokenID2 := tpkg.RandNativeToken().ID

	budget = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	toSpend = budget.Clone()
	require.True(t, budget.SpendFromBudget(toSpend))
	require.True(t, budget.IsEmpty())

	budget = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	cloneBudget := budget.Clone()
	toSpend = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(10)},
		},
	}
	require.False(t, budget.SpendFromBudget(toSpend))
	require.True(t, budget.Equals(cloneBudget))

	budget = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	toSpend = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}
	expected := &Assets{
		Iotas: 0,
		Tokens: iotago.NativeTokens{
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	require.True(t, budget.SpendFromBudget(toSpend))
	require.True(t, budget.Equals(expected))

	budget = &Assets{
		Iotas: 10,
		Tokens: iotago.NativeTokens{
			{ID: tokenID2, Amount: big.NewInt(1)},
		},
	}
	toSpend = &Assets{
		Iotas: 1,
		Tokens: iotago.NativeTokens{
			{ID: tokenID1, Amount: big.NewInt(5)},
		},
	}

	require.False(t, budget.SpendFromBudget(toSpend))
}
