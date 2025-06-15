package isc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"math/big"
	"slices"
	"strings"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
)

type CoinBalances struct {
	items map[coin.Type]coin.Value `bcs:"export"`
}

func NewCoinBalances() CoinBalances {
	return CoinBalances{
		items: make(map[coin.Type]coin.Value),
	}
}

// Iterate returns a deterministic iterator
func (c CoinBalances) Iterate() iter.Seq2[coin.Type, coin.Value] {
	return func(yield func(coin.Type, coin.Value) bool) {
		for _, k := range slices.SortedFunc(maps.Keys(c.items), coin.CompareTypes) {
			if !yield(k, c.items[k]) {
				return
			}
		}
	}
}

func (c CoinBalances) Bytes() []byte {
	return bcs.MustMarshal(&c)
}

func CoinBalancesFromBytes(b []byte) (CoinBalances, error) {
	return bcs.Unmarshal[CoinBalances](b)
}

func (c CoinBalances) Add(coinType coin.Type, amount coin.Value) CoinBalances {
	return c.Set(coinType, c.Get(coinType)+amount)
}

func (c CoinBalances) Set(coinType coin.Type, amount coin.Value) CoinBalances {
	if amount == 0 {
		delete(c.items, coinType)
		return c
	}

	c.items[coinType] = amount
	return c
}

func (c CoinBalances) AddBaseTokens(amount coin.Value) CoinBalances {
	return c.Add(coin.BaseTokenType, amount)
}

func (c CoinBalances) Sub(coinType coin.Type, amount coin.Value) CoinBalances {
	v := c.Get(coinType)
	if v < amount {
		panic("negative coin balance")
	}
	return c.Set(coinType, v-amount)
}

func (c CoinBalances) ToAssets() *Assets {
	return &Assets{
		Coins:   c,
		Objects: NewObjectSet(),
	}
}

func (c CoinBalances) Get(coinType coin.Type) coin.Value {
	return c.items[coinType]
}

func (c CoinBalances) BaseTokens() coin.Value {
	return c.items[coin.BaseTokenType]
}

func (c CoinBalances) NativeTokens() CoinBalances {
	ret := NewCoinBalances()
	for t, v := range c.items {
		// Exclude BaseTokens
		if coin.BaseTokenType.MatchesStringType(t.String()) {
			continue
		}
		ret.items[t] = v
	}
	return ret
}

func (c CoinBalances) IsEmpty() bool {
	return len(c.items) == 0
}

func (c CoinBalances) Size() int {
	return len(c.items)
}

type CoinJSON struct {
	CoinType coin.TypeJSON `json:"coinType" swagger:"required"`
	Balance  string        `json:"balance" swagger:"required,desc(The balance (uint64 as string))"`
}

func (c *CoinBalances) JSON() []CoinJSON {
	var coins []CoinJSON
	for t, v := range c.Iterate() {
		coins = append(coins, CoinJSON{
			CoinType: t.ToTypeJSON(),
			Balance:  v.String(),
		})
	}
	return coins
}

func (c *CoinBalances) UnmarshalJSON(b []byte) error {
	var coins []CoinJSON
	err := json.Unmarshal(b, &coins)
	if err != nil {
		return err
	}
	*c = NewCoinBalances()
	for _, cc := range coins {
		value := lo.Must(coin.ValueFromString(cc.Balance))
		c.Add(cc.CoinType.ToType(), value)
	}
	return nil
}

func (c CoinBalances) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.JSON())
}

func (c CoinBalances) Equals(b CoinBalances) bool {
	if len(c.items) != len(b.items) {
		return false
	}
	for coinType, amount := range c.items {
		bal := b.items[coinType]
		if bal != amount {
			return false
		}
	}
	return true
}

func (c CoinBalances) String() string {
	s := lo.MapToSlice(c.items, func(coinType coin.Type, amount coin.Value) string {
		return fmt.Sprintf("%s: %d", coinType, amount)
	})
	return fmt.Sprintf("CoinBalances{%s}", strings.Join(s, ", "))
}

func (c CoinBalances) Clone() CoinBalances {
	r := NewCoinBalances()
	for coinType, amount := range c.items {
		r.Add(coinType, amount)
	}
	return r
}

// IotaObject represents a non-coin object originally created on L1
type IotaObject struct {
	ID   iotago.ObjectID
	Type iotago.ObjectType
}

func NewIotaObject(id iotago.ObjectID, t iotago.ObjectType) IotaObject {
	return IotaObject{id, t}
}

func (o IotaObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.JSON())
}

func (o IotaObject) JSON() IotaObjectJSON {
	return IotaObjectJSON{
		ID:   o.ID.ToHex(),
		Type: o.Type.ToTypeJSON(),
	}
}

type IotaObjectJSON struct {
	ID   string                `json:"id" swagger:"required,desc(Hex-encoded object ID)"`
	Type iotago.ObjectTypeJSON `json:"type" swagger:"required"`
}

type ObjectSet struct {
	items map[iotago.ObjectID]iotago.ObjectType `bcs:"export"`
}

func NewObjectSet(objs ...IotaObject) ObjectSet {
	items := make(map[iotago.ObjectID]iotago.ObjectType, len(objs))

	for _, obj := range objs {
		items[obj.ID] = obj.Type
	}

	return ObjectSet{
		items: items,
	}
}

func (o ObjectSet) IsEmpty() bool {
	return len(o.items) == 0
}

func (o ObjectSet) Size() int {
	return len(o.items)
}

func (o ObjectSet) Add(obj IotaObject) {
	o.items[obj.ID] = obj.Type
}

func (o ObjectSet) AddAll(obj []IotaObject) {
	for _, iotaObject := range obj {
		o.Add(iotaObject)
	}
}

func (o ObjectSet) Has(id iotago.ObjectID) bool {
	_, ok := o.items[id]
	return ok
}

func (o ObjectSet) KeysSorted() []iotago.ObjectID {
	ids := lo.Keys(o.items)
	slices.SortFunc(ids, func(a, b iotago.ObjectID) int { return bytes.Compare(a[:], b[:]) })
	return ids
}

func (o ObjectSet) Sorted() []IotaObject {
	var ret []IotaObject
	for _, id := range o.KeysSorted() {
		ret = append(ret, NewIotaObject(id, o.items[id]))
	}
	return ret
}

// Iterate returns a deterministic iterator
func (o ObjectSet) Iterate() iter.Seq[IotaObject] {
	return func(yield func(IotaObject) bool) {
		for _, k := range slices.SortedFunc(maps.Keys(o.items), func(a, b iotago.ObjectID) int {
			return bytes.Compare(a[:], b[:])
		}) {
			if !yield(IotaObject{
				ID:   k,
				Type: o.items[k],
			}) {
				return
			}
		}
	}
}

func (o ObjectSet) Clone() ObjectSet {
	r := NewObjectSet()
	for id, t := range o.items {
		r.Add(NewIotaObject(id, t))
	}
	return r
}

func (o *ObjectSet) JSON() []IotaObjectJSON {
	objs := make([]IotaObjectJSON, 0, len(o.items))
	for obj := range o.Iterate() {
		objs = append(objs, obj.JSON())
	}
	return objs
}

func (o *ObjectSet) UnmarshalJSON(b []byte) error {
	var objs []IotaObject
	err := json.Unmarshal(b, &objs)
	if err != nil {
		return err
	}
	*o = NewObjectSet()
	for _, obj := range objs {
		o.Add(NewIotaObject(obj.ID, obj.Type))
	}
	return nil
}

func (o ObjectSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.JSON())
}

func (o ObjectSet) Equals(b ObjectSet) bool {
	if len(o.items) != len(b.items) {
		return false
	}
	for id := range o.items {
		_, ok := b.items[id]
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
	Objects ObjectSet `json:"objects" swagger:"required"`
}

func NewEmptyAssets() *Assets {
	return &Assets{
		Coins:   NewCoinBalances(),
		Objects: NewObjectSet(),
	}
}

func NewAssets(baseTokens coin.Value) *Assets {
	return NewEmptyAssets().AddCoin(coin.BaseTokenType, baseTokens)
}

func AssetsFromAssetsBagWithBalances(assetsBag *iscmove.AssetsBagWithBalances) (*Assets, error) {
	assets := NewEmptyAssets()
	for cointype, coinval := range assetsBag.Coins.Iterate() {
		assets.Coins.Add(coin.MustTypeFromString(cointype.String()), coin.Value(coinval))
	}
	for objectID, t := range assetsBag.Objects.Iterate() {
		assets.Objects.Add(NewIotaObject(objectID, t))
	}
	return assets, nil
}

func AssetsFromISCMove(assets *iscmove.Assets) (*Assets, error) {
	ret := NewEmptyAssets()
	for k, v := range assets.Coins.Iterate() {
		coinType, err := coin.TypeFromString(k.String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse string to coin.Type: %w", err)
		}
		ret.Coins.Add(coinType, coin.Value(v))
	}
	for id, t := range assets.Objects.Iterate() {
		ret.Objects.Add(NewIotaObject(id, t))
	}
	return ret, nil
}

func AssetsFromBytes(b []byte) (*Assets, error) {
	return bcs.Unmarshal[*Assets](b)
}

// Size returns the number of coins and objects in the assets
func (a *Assets) Size() int {
	return len(a.Coins.items) + len(a.Objects.items)
}

func (a *Assets) Clone() *Assets {
	if a == nil {
		return nil
	}
	r := NewEmptyAssets()
	r.Coins = a.Coins.Clone()
	r.Objects = a.Objects.Clone()
	return r
}

func (a *Assets) AddCoin(coinType coin.Type, amount coin.Value) *Assets {
	a.Coins.Add(coinType, amount)
	return a
}

func (a *Assets) AddObject(obj IotaObject) *Assets {
	a.Objects.Add(obj)
	return a
}

func (a *Assets) CoinBalance(coinType coin.Type) coin.Value {
	return a.Coins.Get(coinType)
}

func (a *Assets) String() string {
	s := lo.MapToSlice(a.Coins.items, func(coinType coin.Type, amount coin.Value) string {
		return fmt.Sprintf("%s: %d", coinType, amount)
	})
	s = append(s, lo.MapToSlice(a.Objects.items, func(id iotago.ObjectID, t iotago.ObjectType) string {
		return fmt.Sprintf("%s: %s", t, id)
	})...)
	return fmt.Sprintf("Assets{%s}", strings.Join(s, ", "))
}

func (a *Assets) Bytes() []byte {
	return bcs.MustMarshal(a)
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
	for coinType, spendAmount := range toSpend.Coins.items {
		available, ok := a.Coins.items[coinType]
		if !ok || available < spendAmount {
			return false
		}
	}
	for id := range toSpend.Objects.items {
		if !a.Objects.Has(id) {
			return false
		}
	}

	// budget is enough
	for coinType, spendAmount := range toSpend.Coins.items {
		a.Coins.Sub(coinType, spendAmount)
	}
	for id := range toSpend.Objects.items {
		delete(a.Objects.items, id)
	}
	return true
}

func (a *Assets) Add(b *Assets) *Assets {
	for coinType, amount := range b.Coins.items {
		a.Coins.Add(coinType, amount)
	}
	for id, t := range b.Objects.items {
		a.Objects.Add(NewIotaObject(id, t))
	}
	return a
}

func (a *Assets) IsEmpty() bool {
	return len(a.Coins.items) == 0 && len(a.Objects.items) == 0
}

func (a *Assets) AddBaseTokens(amount coin.Value) *Assets {
	a.Coins.Add(coin.BaseTokenType, amount)
	return a
}

func (a *Assets) SetBaseTokens(amount coin.Value) *Assets {
	a.Coins.Set(coin.BaseTokenType, amount)
	return a
}

func (a *Assets) BaseTokens() coin.Value {
	if a == nil {
		return 0
	}
	return a.Coins.Get(coin.BaseTokenType)
}

func (a *Assets) AsISCMove() *iscmove.Assets {
	r := iscmove.NewEmptyAssets()
	for coinType, amount := range a.Coins.Iterate() {
		if amount > 0 {
			r.SetCoin(
				iotajsonrpc.CoinType(coinType.String()),
				iotajsonrpc.CoinValue(amount),
			)
		}
	}
	for o := range a.Objects.Iterate() {
		r.AddObject(o.ID, o.Type)
	}
	return r
}

func (a *Assets) AsAssetsBagWithBalances(b *iscmove.AssetsBag) *iscmove.AssetsBagWithBalances {
	return &iscmove.AssetsBagWithBalances{
		AssetsBag: *b,
		Assets:    *a.AsISCMove(),
	}
}

// JSONTokenScheme is for now a 1:1 copy of the Stardusts version
type JSONTokenScheme struct {
	Type          int    `json:"type"`
	MintedSupply  string `json:"mintedTokens"`
	MeltedTokens  string `json:"meltedTokens"`
	MaximumSupply string `json:"maximumSupply"`
}

type SimpleTokenScheme struct {
	// The amount of tokens which has been minted.
	MintedTokens *big.Int
	// The amount of tokens which has been melted.
	MeltedTokens *big.Int
	// The maximum supply of tokens controlled.
	MaximumSupply *big.Int
}

func (s *SimpleTokenScheme) Clone() *SimpleTokenScheme {
	return &SimpleTokenScheme{
		MintedTokens:  new(big.Int).Set(s.MintedTokens),
		MeltedTokens:  new(big.Int).Set(s.MeltedTokens),
		MaximumSupply: new(big.Int).Set(s.MaximumSupply),
	}
}
