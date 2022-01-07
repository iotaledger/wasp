package solo

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

type foundryParams struct {
	ch        *Chain
	user      *cryptolib.KeyPair
	sch       iotago.TokenScheme
	tag       *iotago.TokenTag
	maxSupply *big.Int
}

// CreateFoundryGasBudgetIotas always takes 1000 iotas as gas budget and assets for the call
const (
	CreateFoundryGasBudgetIotas = 1000
	MintTokensGasBudgetIotas    = 1000
	DestroyTokensGasBudgetIotas = 1000
)

func (ch *Chain) NewFoundryParams(maxSupply ...uint64) *foundryParams {
	ret := &foundryParams{
		ch:        ch,
		maxSupply: big.NewInt(1),
	}
	if len(maxSupply) > 0 {
		ret.maxSupply = big.NewInt(int64(maxSupply[0]))
	}
	return ret
}

func (fp *foundryParams) WithMaxSupply(maxSupply *big.Int) *foundryParams {
	fp.maxSupply = maxSupply
	return fp
}

func (fp *foundryParams) WithUser(user *cryptolib.KeyPair) *foundryParams {
	fp.user = user
	return fp
}

func (fp *foundryParams) WithTokenScheme(sch iotago.TokenScheme) *foundryParams {
	fp.sch = sch
	return fp
}

func (fp *foundryParams) WithTag(tag *iotago.TokenTag) *foundryParams {
	fp.tag = tag
	return fp
}

func (fp *foundryParams) CreateFoundry() (uint32, iotago.NativeTokenID, error) {
	par := dict.New()
	if fp.sch != nil {
		par.Set(accounts.ParamTokenScheme, codec.EncodeTokenScheme(fp.sch))
	}
	if fp.tag != nil {
		par.Set(accounts.ParamTokenTag, codec.EncodeTokenTag(*fp.tag))
	}
	if fp.maxSupply != nil {
		par.Set(accounts.ParamMaxSupply, codec.EncodeBigIntAbs(fp.maxSupply))
	}
	user := &fp.ch.OriginatorPrivateKey
	if fp.user != nil {
		user = fp.user
	}
	req := NewCallParamsFromDic(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name, par).
		WithGasBudget(CreateFoundryGasBudgetIotas).
		AddAssetsIotas(CreateFoundryGasBudgetIotas)
	res, err := fp.ch.PostRequestSync(req, user)

	retSN := uint32(0)
	var tokenID iotago.NativeTokenID
	if err == nil {
		resDeco := kvdecoder.New(res)
		retSN = resDeco.MustGetUint32(accounts.ParamFoundrySN)
		tokenID, err = fp.ch.GetNativeTokenIDByFoundrySN(retSN)
	}
	return retSN, tokenID, err
}

func (ch *Chain) MintTokens(foundrySN uint32, amount *big.Int, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
		accounts.ParamFoundrySN, foundrySN,
		accounts.ParamSupplyDeltaAbs, amount,
	).
		WithGasBudget(MintTokensGasBudgetIotas).
		AddAssetsIotas(MintTokensGasBudgetIotas)
	if user == nil {
		user = &ch.OriginatorPrivateKey
	}
	_, err := ch.PostRequestSync(req, user)
	return err
}

func (ch *Chain) DestroyTokens(foundrySN uint32, amount *big.Int, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
		accounts.ParamFoundrySN, foundrySN,
		accounts.ParamSupplyDeltaAbs, amount,
		accounts.ParamDestroyTokens, true,
	).WithGasBudget(DestroyTokensGasBudgetIotas)

	if user == nil {
		user = &ch.OriginatorPrivateKey
	}
	_, err := ch.PostRequestSync(req, user)
	return err
}
