package isc

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"sort"

	"github.com/iotaledger/wasp/clients/iscmove/iscmove_types"
	"github.com/iotaledger/wasp/packages/bigint"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type Assets struct {
	BaseTokens   *big.Int     `json:"baseTokens"`
	NativeTokens NativeTokens `json:"nativeTokens"`
}

var BaseTokenID = []byte{}

func NewAssets(baseTokens *big.Int, tokens NativeTokens) *Assets {
	ret := &Assets{
		BaseTokens:   baseTokens,
		NativeTokens: tokens,
	}
	return ret
}

func AssetsFromAssetsBag(assetsBag iscmove_types.AssetsBagWithBalances) *Assets {
	assets := &Assets{
		BaseTokens:   assetsBag.Balances[suijsonrpc.SuiCoinType].TotalBalance.Int,
		NativeTokens: make(NativeTokens, len(assetsBag.Balances)-1),
	}
	cnt := 0
	for k, v := range assetsBag.Balances {
		if k != suijsonrpc.SuiCoinType {
			assets.NativeTokens[cnt] = &NativeToken{
				CoinType: NativeTokenID(k),
				Amount:   v.TotalBalance.Int,
			}
			cnt++
		}
	}
	return assets
}

func NewAssetsBaseTokens(amount uint64) *Assets {
	return &Assets{BaseTokens: new(big.Int).SetUint64(amount)}
}

func NewEmptyAssets() *Assets {
	return &Assets{}
}

func AssetsFromBytes(b []byte) (*Assets, error) {
	if len(b) == 0 {
		return NewEmptyAssets(), nil
	}
	return rwutil.ReadFromBytes(b, NewEmptyAssets())
}

func AssetsFromDict(d dict.Dict) (*Assets, error) {
	ret := NewEmptyAssets()
	for key, val := range d {
		if IsBaseToken([]byte(key)) {
			ret.BaseTokens = new(big.Int).SetBytes(d.Get(kv.Key(BaseTokenID)))
			continue
		}
		id, err := NativeTokenIDFromBytes([]byte(key))
		if err != nil {
			return nil, fmt.Errorf("Assets: %w", err)
		}
		token := &NativeToken{
			CoinType: id,
			Amount:   new(big.Int).SetBytes(val),
		}
		ret.NativeTokens = append(ret.NativeTokens, token)
	}
	return ret, nil
}

func AssetsFromNativeTokenSum(baseTokens uint64, tokens NativeTokenSum) *Assets {
	ret := NewEmptyAssets()
	ret.BaseTokens = new(big.Int).SetUint64(baseTokens)
	for id, val := range tokens {
		ret.NativeTokens = append(ret.NativeTokens, &NativeToken{
			CoinType: id,
			Amount:   val,
		})
	}
	return ret
}

func MustAssetsFromBytes(b []byte) *Assets {
	ret, err := AssetsFromBytes(b)
	if err != nil {
		panic(err)
	}
	return ret
}

// returns nil if nil pointer receiver is cloned
func (a *Assets) Clone() *Assets {
	if a == nil {
		return nil
	}

	return &Assets{
		BaseTokens:   a.BaseTokens,
		NativeTokens: *a.NativeTokens.Clone(),
	}
}

func (a *Assets) AmountNativeToken(nativeTokenID NativeTokenID) *big.Int {
	for _, t := range a.NativeTokens {
		if t.CoinType == nativeTokenID {
			return t.Amount
		}
	}
	return big.NewInt(0)
}

func (a *Assets) String() string {
	ret := fmt.Sprintf("base tokens: %d", a.BaseTokens)
	if len(a.NativeTokens) > 0 {
		ret += fmt.Sprintf(", tokens (%d):", len(a.NativeTokens))
	}
	for _, nt := range a.NativeTokens {
		ret += fmt.Sprintf("\n       %s: %d", nt.CoinType, nt.Amount)
	}
	return ret
}

func (a *Assets) Bytes() []byte {
	return rwutil.WriteToBytes(a)
}

func (a *Assets) Equals(b *Assets) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if !bigint.Equal(a.BaseTokens, b.BaseTokens) {
		return false
	}

	if len(a.NativeTokens) != len(b.NativeTokens) {
		return false
	}
	bTokensSet := b.NativeTokens.MustSet()
	for _, nativeToken := range a.NativeTokens {
		if !bigint.Equal(nativeToken.Amount, bTokensSet[nativeToken.CoinType].Amount) {
			return false
		}
	}
	return true
}

// Spend subtracts assets from the current set.
// Mutates receiver `a` !
// If budget is not enough, returns false and leaves receiver untouched
func (a *Assets) Spend(toSpend *Assets) bool {
	if a.IsEmpty() {
		return toSpend.IsEmpty()
	}
	if toSpend.IsEmpty() {
		return true
	}
	if a.Equals(toSpend) {
		a.BaseTokens = big.NewInt(0)
		a.NativeTokens = nil
		return true
	}
	if bigint.Less(a.BaseTokens, toSpend.BaseTokens) {
		return false
	}

	targetSet := a.NativeTokens.Clone().MustSet()

	for _, nativeToken := range toSpend.NativeTokens {
		curr, ok := targetSet[nativeToken.CoinType]
		if !ok || bigint.Less(curr.Amount, nativeToken.Amount) {
			return false
		}
		curr.Amount = bigint.Sub(curr.Amount, nativeToken.Amount)
	}

	// budget is enough
	a.BaseTokens = bigint.Sub(a.BaseTokens, toSpend.BaseTokens)
	a.NativeTokens = a.NativeTokens[:0]
	for _, nativeToken := range targetSet {
		if util.IsZeroBigInt(nativeToken.Amount) {
			continue
		}
		a.NativeTokens = append(a.NativeTokens, nativeToken)
	}
	return true
}

func (a *Assets) Add(b *Assets) *Assets {
	a.BaseTokens = bigint.Add(a.BaseTokens, b.BaseTokens)
	resultTokens := a.NativeTokens.MustSet()
	for _, nativeToken := range b.NativeTokens {
		if resultTokens[nativeToken.CoinType] != nil {
			resultTokens[nativeToken.CoinType].Amount = bigint.Add(
				resultTokens[nativeToken.CoinType].Amount,
				nativeToken.Amount,
			)
			continue
		}
		resultTokens[nativeToken.CoinType] = nativeToken
	}
	a.NativeTokens = nativeTokensFromSet(resultTokens)
	return a
}

func (a *Assets) IsEmpty() bool {
	return a == nil || a.BaseTokens == nil || (bigint.Equal(a.BaseTokens, big.NewInt(0)) &&
		len(a.NativeTokens) == 0)
}

func (a *Assets) AddBaseTokens(amount uint64) *Assets {
	a.BaseTokens = bigint.Add(a.BaseTokens, big.NewInt(0).SetUint64(amount))
	return a
}

func (a *Assets) AddNativeTokens(nativeTokenID NativeTokenID, amount *big.Int) *Assets {
	b := NewAssets(big.NewInt(0), NativeTokens{
		&NativeToken{
			CoinType: nativeTokenID,
			Amount:   amount,
		},
	})
	return a.Add(b)
}

func (a *Assets) ToDict() dict.Dict {
	ret := dict.New()
	ret.Set(kv.Key(BaseTokenID), a.BaseTokens.Bytes())
	for _, nativeToken := range a.NativeTokens {
		ret.Set(kv.Key(nativeToken.CoinType[:]), nativeToken.Amount.Bytes())
	}
	return ret
}

func nativeTokensFromSet(nativeTokenSet NativeTokensSet) NativeTokens {
	ret := make(NativeTokens, len(nativeTokenSet))
	i := 0
	for _, nativeToken := range nativeTokenSet {
		ret[i] = nativeToken
		i++
	}
	return ret
}

// IsBaseToken return whether a given tokenID represents the base token
func IsBaseToken(tokenID []byte) bool {
	return bytes.Equal(tokenID, BaseTokenID)
}

// Since we are encoding a nil assets pointer with a byte already,
// we may as well use more of the byte to compress the data further.
// We're adding 3 flags to indicate the presence of the subcomponents
// of the assets so that we may skip reading/writing them altogether.
const (
	hasBaseTokens   = 0x80
	hasNativeTokens = 0x40
	hasNFTs         = 0x20
)

func (a *Assets) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	flags := rr.ReadByte()
	if flags == 0x00 {
		return rr.Err
	}
	if (flags & hasBaseTokens) != 0 {
		a.BaseTokens = big.NewInt(0).SetUint64(rr.ReadAmount64())
	}
	if (flags & hasNativeTokens) != 0 {
		size := rr.ReadSize16()
		a.NativeTokens = make(NativeTokens, size)
		for i := range a.NativeTokens {
			nativeToken := new(NativeToken)
			a.NativeTokens[i] = nativeToken
			nativeToken.CoinType = NativeTokenID(rr.ReadString())
			nativeToken.Amount = rr.ReadUint256()
		}
	}
	return rr.Err
}

func (a *Assets) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	if a.IsEmpty() {
		ww.WriteByte(0x00)
		return ww.Err
	}

	var flags byte
	if !bigint.Equal(a.BaseTokens, big.NewInt(0)) {
		flags |= hasBaseTokens
	}
	if len(a.NativeTokens) != 0 {
		flags |= hasNativeTokens
	}

	ww.WriteByte(flags)
	if (flags & hasBaseTokens) != 0 {
		ww.WriteAmount64(a.BaseTokens.Uint64())
	}
	if (flags & hasNativeTokens) != 0 {
		ww.WriteSize16(len(a.NativeTokens))
		sort.Slice(a.NativeTokens, func(lhs, rhs int) bool {
			return a.NativeTokens[lhs].CoinType < a.NativeTokens[rhs].CoinType
		})
		for _, nativeToken := range a.NativeTokens {
			ww.WriteString(string(nativeToken.CoinType))
			ww.WriteUint256(nativeToken.Amount)
		}
	}
	return ww.Err
}
