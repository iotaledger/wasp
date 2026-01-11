package graphqltypes_test

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

func TestCoins_PickIOTACoinsWithGas(t *testing.T) {
	// coins 1,2,3,4,5
	testCoins := graphqltypes.Coins{
		{Balance: graphqltypes.NewBigInt(3)},
		{Balance: graphqltypes.NewBigInt(5)},
		{Balance: graphqltypes.NewBigInt(1)},
		{Balance: graphqltypes.NewBigInt(4)},
		{Balance: graphqltypes.NewBigInt(2)},
	}
	type args struct {
		amount     *big.Int
		gasAmount  uint64
		pickMethod int
	}
	tests := []struct {
		name    string
		cs      graphqltypes.Coins
		args    args
		want    graphqltypes.Coins
		want1   *graphqltypes.Coin
		wantErr bool
	}{
		{
			name: "case success 1",
			cs:   testCoins,
			args: args{
				amount:     new(big.Int),
				gasAmount:  0,
				pickMethod: graphqltypes.PickMethodSmaller,
			},
			want:    nil,
			want1:   nil,
			wantErr: false,
		},
		{
			name: "case success 2",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(1),
				gasAmount:  2,
				pickMethod: graphqltypes.PickMethodSmaller,
			},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(1)}},
			want1:   &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(2)},
			wantErr: false,
		},
		{
			name: "case success 3",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(4),
				gasAmount:  2,
				pickMethod: graphqltypes.PickMethodSmaller,
			},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(1)}, {Balance: graphqltypes.NewBigInt(3)}},
			want1:   &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(2)},
			wantErr: false,
		},
		{
			name: "case success 4",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(6),
				gasAmount:  2,
				pickMethod: graphqltypes.PickMethodSmaller,
			},
			want: graphqltypes.Coins{
				{Balance: graphqltypes.NewBigInt(1)},
				{Balance: graphqltypes.NewBigInt(3)},
				{Balance: graphqltypes.NewBigInt(4)},
			},
			want1:   &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(2)},
			wantErr: false,
		},
		{
			name: "case error 1",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(6),
				gasAmount:  6,
				pickMethod: graphqltypes.PickMethodSmaller,
			},
			want:    graphqltypes.Coins{},
			want1:   nil,
			wantErr: true,
		},
		{
			name: "case error 1",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(100),
				gasAmount:  3,
				pickMethod: graphqltypes.PickMethodSmaller,
			},
			want:    graphqltypes.Coins{},
			want1:   &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(3)},
			wantErr: true,
		},
		{
			name: "case bigger 1",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(3),
				gasAmount:  3,
				pickMethod: graphqltypes.PickMethodBigger,
			},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(5)}},
			want1:   &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(3)},
			wantErr: false,
		},
		{
			name: "case order 1",
			cs:   testCoins,
			args: args{
				amount:     big.NewInt(3),
				gasAmount:  3,
				pickMethod: graphqltypes.PickMethodByOrder,
			},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(5)}},
			want1:   &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(3)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, got1, err := tt.cs.PickIOTACoinsWithGas(tt.args.amount, tt.args.gasAmount, tt.args.pickMethod)
				if (err != nil) != tt.wantErr {
					t.Errorf("Coins.PickIOTACoinsWithGas() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) != 0 && len(tt.want) != 0 {
					if !reflect.DeepEqual(got, tt.want) {
						t.Errorf("Coins.PickIOTACoinsWithGas() got: %v, want %v", got, tt.want)
					}
				}
				if !reflect.DeepEqual(got1, tt.want1) {
					t.Errorf("Coins.PickIOTACoinsWithGas() got1: %v, want %v", got1, tt.want1)
				}
			},
		)
	}
}

func TestCoins_PickCoins(t *testing.T) {
	// coins 1,2,3,4,5
	testCoins := graphqltypes.Coins{
		{Balance: graphqltypes.NewBigInt(3)},
		{Balance: graphqltypes.NewBigInt(5)},
		{Balance: graphqltypes.NewBigInt(1)},
		{Balance: graphqltypes.NewBigInt(4)},
		{Balance: graphqltypes.NewBigInt(2)},
	}
	type args struct {
		amount     *big.Int
		pickMethod int
	}
	tests := []struct {
		name    string
		cs      graphqltypes.Coins
		args    args
		want    graphqltypes.Coins
		wantErr bool
	}{
		{
			name:    "smaller 1",
			cs:      testCoins,
			args:    args{amount: big.NewInt(2), pickMethod: graphqltypes.PickMethodSmaller},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(1)}, {Balance: graphqltypes.NewBigInt(2)}},
			wantErr: false,
		},
		{
			name: "smaller 2",
			cs:   testCoins,
			args: args{amount: big.NewInt(4), pickMethod: graphqltypes.PickMethodSmaller},
			want: graphqltypes.Coins{
				{Balance: graphqltypes.NewBigInt(1)},
				{Balance: graphqltypes.NewBigInt(2)},
				{Balance: graphqltypes.NewBigInt(3)},
			},
			wantErr: false,
		},
		{
			name:    "bigger 1",
			cs:      testCoins,
			args:    args{amount: big.NewInt(2), pickMethod: graphqltypes.PickMethodBigger},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(5)}},
			wantErr: false,
		},
		{
			name:    "bigger 2",
			cs:      testCoins,
			args:    args{amount: big.NewInt(6), pickMethod: graphqltypes.PickMethodBigger},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(5)}, {Balance: graphqltypes.NewBigInt(4)}},
			wantErr: false,
		},
		{
			name:    "pick by order 1",
			cs:      testCoins,
			args:    args{amount: big.NewInt(6), pickMethod: graphqltypes.PickMethodByOrder},
			want:    graphqltypes.Coins{{Balance: graphqltypes.NewBigInt(3)}, {Balance: graphqltypes.NewBigInt(5)}},
			wantErr: false,
		},
		{
			name:    "pick by order 2",
			cs:      testCoins,
			args:    args{amount: big.NewInt(15), pickMethod: graphqltypes.PickMethodByOrder},
			want:    testCoins,
			wantErr: false,
		},
		{
			name:    "pick error",
			cs:      testCoins,
			args:    args{amount: big.NewInt(16), pickMethod: graphqltypes.PickMethodByOrder},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.cs.PickCoins(tt.args.amount, tt.args.pickMethod)
				if (err != nil) != tt.wantErr {
					t.Errorf("Coins.PickCoins() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Coins.PickCoins(): %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestPickupCoins(t *testing.T) {
	coin := func(n uint64) *graphqltypes.Coin {
		return &graphqltypes.Coin{Balance: graphqltypes.NewBigInt(uint64(n)), CoinType: graphqltypes.IotaCoinType}
	}

	type args struct {
		inputCoins   *graphqltypes.CoinPage
		targetAmount *big.Int
		gasBudget    uint64
		limit        int
		moreCount    int
	}
	tests := []struct {
		name    string
		args    args
		want    *graphqltypes.PickedCoins
		wantErr error
	}{
		{
			name: "moreCount = 3",
			args: args{
				inputCoins: &graphqltypes.CoinPage{
					Data: []*graphqltypes.Coin{
						coin(1e3), coin(1e5), coin(1e2), coin(1e4),
					},
				},
				targetAmount: big.NewInt(1e3),
				moreCount:    3,
			},
			want: &graphqltypes.PickedCoins{
				Coins: []*graphqltypes.Coin{
					coin(1e3), coin(1e5), coin(1e2),
				},
				TotalAmount:  big.NewInt(1e3 + 1e5 + 1e2),
				TargetAmount: big.NewInt(1e3),
			},
		},
		{
			name: "large gas",
			args: args{
				inputCoins: &graphqltypes.CoinPage{
					Data: []*graphqltypes.Coin{
						coin(1e3), coin(1e5), coin(1e2), coin(1e4),
					},
				},
				targetAmount: big.NewInt(1e3),
				gasBudget:    1e9,
				moreCount:    3,
			},
			want: &graphqltypes.PickedCoins{
				Coins: []*graphqltypes.Coin{
					coin(1e3), coin(1e5), coin(1e2), coin(1e4),
				},
				TotalAmount:  big.NewInt(1e3 + 1e5 + 1e2 + 1e4),
				TargetAmount: big.NewInt(1e3),
			},
		},
		{
			name: "ErrNoCoinsFound",
			args: args{
				inputCoins: &graphqltypes.CoinPage{
					Data: []*graphqltypes.Coin{},
				},
				targetAmount: big.NewInt(101000),
			},
			wantErr: graphqltypes.ErrNoCoinsFound,
		},
		{
			name: "ErrInsufficientBalance",
			args: args{
				inputCoins: &graphqltypes.CoinPage{
					Data: []*graphqltypes.Coin{
						coin(1e5), coin(1e6), coin(1e4),
					},
				},
				targetAmount: big.NewInt(1e9),
			},
			wantErr: graphqltypes.ErrInsufficientBalance,
		},
		{
			name: "ErrNeedMergeCoin 1",
			args: args{
				inputCoins: &graphqltypes.CoinPage{
					Data: []*graphqltypes.Coin{
						coin(1e5), coin(1e6), coin(1e4),
					},
					HasNextPage: true,
				},
				targetAmount: big.NewInt(1e9),
			},
			wantErr: graphqltypes.ErrNeedMergeCoin,
		},
		{
			name: "ErrNeedMergeCoin 2",
			args: args{
				inputCoins: &graphqltypes.CoinPage{
					Data: []*graphqltypes.Coin{
						coin(1e5), coin(1e6), coin(1e4), coin(1e5),
					},
					HasNextPage: false,
				},
				targetAmount: big.NewInt(1e6 + 1e5*2 + 1e3),
				limit:        3,
			},
			wantErr: graphqltypes.ErrNeedMergeCoin,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := graphqltypes.PickupCoins(
					tt.args.inputCoins,
					tt.args.targetAmount,
					tt.args.gasBudget,
					tt.args.limit,
					tt.args.moreCount,
				)
				require.Equal(t, err, tt.wantErr)
				require.Equal(t, got, tt.want)
			},
		)
	}
}

func TestUnmarshalCoinFields(t *testing.T) {
	s := []byte(`{"balance":"46952212","id":{"id": "0x0679bceafb254938dc123032e6d2d3c1a3e650a0c681bf0d997d38ff7eb88738"}}`)
	var coinFields graphqltypes.CoinFields
	err := json.Unmarshal(s, &coinFields)
	require.NoError(t, err)
	testObjectID := iotago.MustObjectIDFromHex("0x0679bceafb254938dc123032e6d2d3c1a3e650a0c681bf0d997d38ff7eb88738")
	require.Equal(t, uint64(46952212), coinFields.Balance.Uint64())
	require.Equal(t, testObjectID, coinFields.ID.ID)
}
