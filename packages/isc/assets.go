package isc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math/big"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/bigint"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// TODO: maybe it is not ok to consider this constant?
const BaseTokenType = CoinType(suijsonrpc.SuiCoinType)

// CoinType is the string representation of a Sui coin type, e.g. `0x2::sui::SUI`
type CoinType string

func (c CoinType) Write(w io.Writer) error {
	rt := lo.Must(sui.NewResourceType(string(c)))
	if rt.SubType != nil {
		panic("cointype with subtype is unsupported")
	}
	ww := rwutil.NewWriter(w)
	ww.WriteN(rt.Address[:])
	ww.WriteString(rt.Module)
	ww.WriteString(rt.ObjectName)
	return ww.Err
}

func (c *CoinType) Read(r io.Reader) error {
	rt := sui.ResourceType{
		Address: &sui.Address{},
	}
	rr := rwutil.NewReader(r)
	rr.ReadN(rt.Address[:])
	rt.Module = rr.ReadString()
	rt.ObjectName = rr.ReadString()
	*c = CoinType(rt.ShortString())
	return rr.Err
}

func (c CoinType) String() string {
	return string(c)
}

func (c CoinType) Bytes() []byte {
	return rwutil.WriteToBytes(c)
}

func CoinTypeFromBytes(b []byte) (CoinType, error) {
	var r CoinType
	_, err := rwutil.ReadFromBytes(b, &r)
	return r, err
}

type CoinBalances map[CoinType]*big.Int

func NewCoinBalances() CoinBalances {
	return make(CoinBalances)
}

func CoinBalancesFromDict(d dict.Dict) (CoinBalances, error) {
	ret := NewCoinBalances()
	for key, val := range d {
		coinType, err := CoinTypeFromBytes([]byte(key))
		if err != nil {
			return nil, fmt.Errorf("AssetsFromCoinsDict: %w", err)
		}
		ret.Add(coinType, new(big.Int).SetBytes(val))
	}
	return ret, nil
}

func (c CoinBalances) IterateSorted(f func(CoinType, *big.Int) bool) {
	types := lo.Keys(c)
	slices.Sort(types)
	for _, coinType := range types {
		if !f(coinType, c[coinType]) {
			return
		}
	}
}

func (c CoinBalances) ToDict() dict.Dict {
	ret := dict.New()
	for coinType, amount := range c {
		ret.Set(kv.Key(coinType.Bytes()), amount.Bytes())
	}
	return ret
}

func (c *CoinBalances) Read(r io.Reader) error {
	*c = NewCoinBalances()
	rr := rwutil.NewReader(r)
	n := rr.ReadSize32()
	for i := 0; i < n; i++ {
		var coinType CoinType
		rr.Read(&coinType)
		amount := rr.ReadBigUint()
		c.Add(coinType, amount)
	}
	return rr.Err
}

func (c CoinBalances) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteSize32(len(c))
	c.IterateSorted(func(t CoinType, a *big.Int) bool {
		ww.Write(t)
		ww.WriteBigUint(a)
		return true
	})
	return ww.Err
}

func (c CoinBalances) Add(coinType CoinType, amount *big.Int) CoinBalances {
	if amount.Sign() == 0 {
		return c
	}
	if _, ok := c[coinType]; !ok {
		c[coinType] = new(big.Int).Set(amount)
		return c
	}
	c[coinType].Add(c[coinType], amount)
	return c
}

func (c CoinBalances) Sub(coinType CoinType, amount *big.Int) CoinBalances {
	c[coinType] = bigint.Sub(c[coinType], amount)
	s := c[coinType].Sign()
	if s < 0 {
		panic("negative coin balance")
	}
	if s == 0 {
		delete(c, coinType)
	}
	return c
}

func (c CoinBalances) Get(coinType CoinType) *big.Int {
	r := c[coinType]
	if r == nil {
		return big.NewInt(0)
	}
	return r
}

type CoinJSON struct {
	CoinType CoinType           `json:"coinType" swagger:"required"`
	Balance  *suijsonrpc.BigInt `json:"balance" swagger:"required"`
}

func (c *CoinBalances) UnmarshalJSON(b []byte) error {
	var coins []CoinJSON
	err := json.Unmarshal(b, &coins)
	if err != nil {
		return err
	}
	*c = NewCoinBalances()
	for _, coin := range coins {
		c.Add(coin.CoinType, coin.Balance.Int)
	}
	return nil
}

func (c CoinBalances) MarshalJSON() ([]byte, error) {
	var coins []CoinJSON
	c.IterateSorted(func(t CoinType, a *big.Int) bool {
		coins = append(coins, CoinJSON{
			CoinType: t,
			Balance:  &suijsonrpc.BigInt{Int: a},
		})
		return true
	})
	return json.Marshal(coins)
}

func (c CoinBalances) Equals(b CoinBalances) bool {
	if len(c) != len(b) {
		return false
	}
	for coinType, amount := range c {
		bal := b[coinType]
		if !bigint.Equal(bal, amount) {
			return false
		}
	}
	return true
}

type ObjectIDSet map[sui.ObjectID]struct{}

func NewObjectIDSet() ObjectIDSet {
	return make(map[sui.ObjectID]struct{})
}

func (o ObjectIDSet) Add(id sui.ObjectID) {
	o[id] = struct{}{}
}

func (o ObjectIDSet) Has(id sui.ObjectID) bool {
	_, ok := o[id]
	return ok
}

func (o ObjectIDSet) Sorted() []sui.ObjectID {
	ids := lo.Keys(o)
	slices.SortFunc(ids, func(a, b sui.ObjectID) int { return bytes.Compare(a[:], b[:]) })
	return ids
}

func (o ObjectIDSet) IterateSorted(f func(sui.ObjectID) bool) {
	for _, id := range o.Sorted() {
		if !f(id) {
			return
		}
	}
}

func (o *ObjectIDSet) Read(r io.Reader) error {
	*o = NewObjectIDSet()
	rr := rwutil.NewReader(r)
	n := rr.ReadSize32()
	for i := 0; i < n; i++ {
		var id sui.ObjectID
		rr.ReadN(id[:])
		o.Add(id)
	}
	return rr.Err
}

func (o ObjectIDSet) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteSize32(len(o))
	o.IterateSorted(func(id sui.ObjectID) bool {
		ww.WriteN(id[:])
		return true
	})
	return ww.Err
}

func (o *ObjectIDSet) UnmarshalJSON(b []byte) error {
	var ids []sui.ObjectID
	err := json.Unmarshal(b, &ids)
	if err != nil {
		return err
	}
	*o = NewObjectIDSet()
	for _, id := range ids {
		o.Add(id)
	}
	return nil
}

func (o ObjectIDSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Sorted())
}

func (o ObjectIDSet) Equals(b ObjectIDSet) bool {
	if len(o) != len(b) {
		return false
	}
	for id := range o {
		_, ok := b[id]
		if !ok {
			return false
		}
	}
	return true
}

type Assets struct {
	// Coins is a set of coin balances
	Coins CoinBalances `json:"coins" swagger:"required"`
	// Objects is a set of non-Coin object IDs (e.g. NFTs)
	Objects ObjectIDSet `json:"objects" swagger:"required"`
}

func NewEmptyAssets() *Assets {
	return &Assets{
		Coins:   NewCoinBalances(),
		Objects: NewObjectIDSet(),
	}
}

func NewAssets(baseTokens *big.Int) *Assets {
	return NewEmptyAssets().AddCoin(BaseTokenType, baseTokens)
}

func AssetsFromAssetsBag(assetsBag iscmove.AssetsBagWithBalances) *Assets {
	assets := NewEmptyAssets()
	for k, v := range assetsBag.Balances {
		assets.Coins.Add(CoinType(k), v.TotalBalance.Int)
	}
	return assets
}

func NewAssetsBaseTokens(amount uint64) *Assets {
	return NewAssets(big.NewInt(0).SetUint64(amount))
}

func AssetsFromBytes(b []byte) (*Assets, error) {
	return rwutil.ReadFromBytes(b, NewEmptyAssets())
}

func (a *Assets) Clone() *Assets {
	r := NewEmptyAssets()
	for coinType, amount := range a.Coins {
		r.Coins.Add(coinType, amount)
	}
	r.Objects = maps.Clone(a.Objects)
	return r
}

func (a *Assets) AddCoin(coinType CoinType, amount *big.Int) *Assets {
	a.Coins.Add(coinType, amount)
	return a
}

func (a *Assets) AddObject(id sui.ObjectID) *Assets {
	a.Objects.Add(id)
	return a
}

func (a *Assets) CoinBalance(coinType CoinType) *big.Int {
	return a.Coins.Get(coinType)
}

func (a *Assets) String() string {
	s := lo.MapToSlice(a.Coins, func(coinType CoinType, amount *big.Int) string {
		return fmt.Sprintf("%s: %s", coinType, amount.Text(10))
	})
	s = append(s, lo.MapToSlice(a.Objects, func(id sui.ObjectID, _ struct{}) string {
		return id.ShortString()
	})...)
	return fmt.Sprintf("Assets{%s}", strings.Join(s, ", "))
}

func (a *Assets) Bytes() []byte {
	return rwutil.WriteToBytes(a)
}

func (a *Assets) Equals(b *Assets) bool {
	if a == b {
		return true
	}
	if !a.Coins.Equals(b.Coins) {
		return false
	}
	if !a.Objects.Equals(b.Objects) {
		return false
	}
	return true
}

// Spend subtracts assets from the current set, mutating the receiver.
// If the budget is not enough, returns false and leaves receiver untouched.
func (a *Assets) Spend(toSpend *Assets) bool {
	// check budget
	for coinType, spendAmount := range toSpend.Coins {
		available, ok := a.Coins[coinType]
		if !ok || bigint.Less(available, spendAmount) {
			return false
		}
	}
	for id := range toSpend.Objects {
		if !a.Objects.Has(id) {
			return false
		}
	}

	// budget is enough
	for coinType, spendAmount := range toSpend.Coins {
		a.Coins.Sub(coinType, spendAmount)
	}
	for id := range toSpend.Objects {
		delete(a.Objects, id)
	}
	return true
}

func (a *Assets) Add(b *Assets) *Assets {
	for coinType, amount := range b.Coins {
		a.Coins.Add(coinType, amount)
	}
	for id := range b.Objects {
		a.Objects.Add(id)
	}
	return a
}

func (a *Assets) IsEmpty() bool {
	return len(a.Coins) == 0 && len(a.Objects) == 0
}

func (a *Assets) AddBaseTokens(amount *big.Int) *Assets {
	a.Coins.Add(BaseTokenType, amount)
	return a
}

func (a *Assets) BaseTokens() *big.Int {
	return a.Coins.Get(BaseTokenType)
}

func (a *Assets) Read(r io.Reader) error {
	*a = Assets{}
	rr := rwutil.NewReader(r)
	rr.Read(&a.Coins)
	rr.Read(&a.Objects)
	return rr.Err
}

func (a *Assets) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(a.Coins)
	ww.Write(a.Objects)
	return ww.Err
}
