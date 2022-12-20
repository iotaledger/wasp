package isc

import (
	"bytes"
	"fmt"
	"math/big"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// FungibleTokens is used as assets in the UTXO and as tokens in transfer
type FungibleTokens struct {
	BaseTokens   uint64              `json:"baseTokens"`
	NativeTokens iotago.NativeTokens `json:"nativeTokens"`
}

var BaseTokenID = []byte{}

func NewEmptyFungibleTokens() *FungibleTokens {
	return &FungibleTokens{
		NativeTokens: make([]*iotago.NativeToken, 0),
	}
}

func NewFungibleTokens(baseTokens uint64, tokens iotago.NativeTokens) *FungibleTokens {
	return &FungibleTokens{
		BaseTokens:   baseTokens,
		NativeTokens: tokens,
	}
}

func NewFungibleBaseTokens(amount uint64) *FungibleTokens {
	return &FungibleTokens{BaseTokens: amount}
}

func NewFungibleTokensForGasFee(p *gas.GasFeePolicy, feeAmount uint64) *FungibleTokens {
	if p.GasFeeTokenID == nil {
		return NewFungibleBaseTokens(feeAmount)
	}
	return NewEmptyFungibleTokens().AddNativeTokens(*p.GasFeeTokenID, feeAmount)
}

func FungibleTokensFromDict(d dict.Dict) (*FungibleTokens, error) {
	ret := NewEmptyFungibleTokens()
	for key, val := range d {
		if IsBaseToken([]byte(key)) {
			ret.BaseTokens = new(big.Int).SetBytes(d.MustGet(kv.Key(BaseTokenID))).Uint64()
			continue
		}
		id, err := NativeTokenIDFromBytes([]byte(key))
		if err != nil {
			return nil, xerrors.Errorf("FungibleTokensFromDict: %w", err)
		}
		token := &iotago.NativeToken{
			ID:     id,
			Amount: new(big.Int).SetBytes(val),
		}
		ret.NativeTokens = append(ret.NativeTokens, token)
	}
	return ret, nil
}

func FungibleTokensFromNativeTokenSum(baseTokens uint64, tokens iotago.NativeTokenSum) *FungibleTokens {
	ret := NewEmptyFungibleTokens()
	ret.BaseTokens = baseTokens
	for id, val := range tokens {
		ret.NativeTokens = append(ret.NativeTokens, &iotago.NativeToken{
			ID:     id,
			Amount: val,
		})
	}
	return ret
}

func FungibleTokensFromOutputMap(outs map[iotago.OutputID]iotago.Output) *FungibleTokens {
	ret := NewEmptyFungibleTokens()
	for _, out := range outs {
		ret.Add(FungibleTokensFromOutput(out))
	}
	return ret
}

func FungibleTokensFromOutput(o iotago.Output) *FungibleTokens {
	ret := &FungibleTokens{
		BaseTokens:   o.Deposit(),
		NativeTokens: o.NativeTokenList().Clone(),
	}
	return ret
}

func NativeTokenIDFromBytes(data []byte) (iotago.NativeTokenID, error) {
	if len(data) != iotago.NativeTokenIDLength {
		return iotago.NativeTokenID{}, xerrors.New("NativeTokenIDFromBytes: wrong data length")
	}
	var tokenID iotago.NativeTokenID
	copy(tokenID[:], data)
	return tokenID, nil
}

func MustNativeTokenIDFromBytes(data []byte) iotago.NativeTokenID {
	ret, err := NativeTokenIDFromBytes(data)
	if err != nil {
		panic(xerrors.Errorf("MustNativeTokenIDFromBytes: %w", err))
	}
	return ret
}

// returns nil if nil pointer receiver is cloned
func (a *FungibleTokens) Clone() *FungibleTokens {
	if a == nil {
		return nil
	}

	return &FungibleTokens{
		BaseTokens:   a.BaseTokens,
		NativeTokens: a.NativeTokens.Clone(),
	}
}

func (a *FungibleTokens) AmountNativeToken(tokenID *iotago.NativeTokenID) *big.Int {
	for _, t := range a.NativeTokens {
		if t.ID == *tokenID {
			return t.Amount
		}
	}
	return big.NewInt(0)
}

func (a *FungibleTokens) String() string {
	ret := fmt.Sprintf("base tokens: %d", a.BaseTokens)
	if len(a.NativeTokens) > 0 {
		ret += fmt.Sprintf(", tokens (%d):", len(a.NativeTokens))
	}
	for _, nt := range a.NativeTokens {
		ret += fmt.Sprintf("\n       %s: %d", nt.ID.String(), nt.Amount)
	}
	return ret
}

func (a *FungibleTokens) Bytes() []byte {
	mu := marshalutil.New()
	a.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (a *FungibleTokens) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteUint64(a.BaseTokens)
	tokenBytes, err := serializer.NewSerializer().WriteSliceOfObjects(&a.NativeTokens, serializer.DeSeriModePerformLexicalOrdering, nil, serializer.SeriLengthPrefixTypeAsUint16, &NativeAssetsSerializationArrayRules, func(err error) error {
		return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
	}).Serialize()
	if err != nil {
		panic(fmt.Errorf("unexpected error serializing native tokens: %w", err))
	}
	mu.WriteUint16(uint16(len(tokenBytes)))
	mu.WriteBytes(tokenBytes)
}

func FungibleTokensFromMarshalUtil(mu *marshalutil.MarshalUtil) (*FungibleTokens, error) {
	ret := &FungibleTokens{
		NativeTokens: make(iotago.NativeTokens, 0),
	}
	var err error
	if ret.BaseTokens, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	tokenBytesLength, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	tokenBytes, err := mu.ReadBytes(int(tokenBytesLength))
	if err != nil {
		return nil, err
	}
	_, err = serializer.NewDeserializer(tokenBytes).
		ReadSliceOfObjects(&ret.NativeTokens, serializer.DeSeriModePerformLexicalOrdering, nil, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, &NativeAssetsSerializationArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for alias output: %w", err)
		}).Done()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *FungibleTokens) Equals(b *FungibleTokens) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.BaseTokens != b.BaseTokens {
		return false
	}
	if len(a.NativeTokens) != len(b.NativeTokens) {
		return false
	}
	bTokensSet := b.NativeTokens.MustSet()
	for _, nativeToken := range a.NativeTokens {
		if nativeToken.Amount.Cmp(bTokensSet[nativeToken.ID].Amount) != 0 {
			return false
		}
	}
	return true
}

// SpendFromFungibleTokenBudget subtracts fungible tokens from budget.
// Mutates receiver `a` !
// If budget is not enough, returns false and leaves receiver untouched
func (a *FungibleTokens) SpendFromFungibleTokenBudget(toSpend *FungibleTokens) bool {
	if a.IsEmpty() {
		return toSpend.IsEmpty()
	}
	if toSpend.IsEmpty() {
		return true
	}
	if a.Equals(toSpend) {
		a.BaseTokens = 0
		a.NativeTokens = nil
		return true
	}
	if a.BaseTokens < toSpend.BaseTokens {
		return false
	}
	targetSet := a.NativeTokens.Clone().MustSet()

	for _, nativeToken := range toSpend.NativeTokens {
		curr, ok := targetSet[nativeToken.ID]
		if !ok || curr.Amount.Cmp(nativeToken.Amount) < 0 {
			return false
		}
		curr.Amount.Sub(curr.Amount, nativeToken.Amount)
	}
	// budget is enough
	a.BaseTokens -= toSpend.BaseTokens
	a.NativeTokens = a.NativeTokens[:0]
	for _, nativeToken := range targetSet {
		if util.IsZeroBigInt(nativeToken.Amount) {
			continue
		}
		a.NativeTokens = append(a.NativeTokens, nativeToken)
	}
	return true
}

func (a *FungibleTokens) Add(b *FungibleTokens) *FungibleTokens {
	a.BaseTokens += b.BaseTokens
	resultTokens := a.NativeTokens.MustSet()
	for _, nativeToken := range b.NativeTokens {
		if resultTokens[nativeToken.ID] != nil {
			resultTokens[nativeToken.ID].Amount.Add(
				resultTokens[nativeToken.ID].Amount,
				nativeToken.Amount,
			)
			continue
		}
		resultTokens[nativeToken.ID] = nativeToken
	}
	a.NativeTokens = nativeTokensFromSet(resultTokens)
	return a
}

func (a *FungibleTokens) IsEmpty() bool {
	return a == nil || a.BaseTokens == 0 && len(a.NativeTokens) == 0
}

func (a *FungibleTokens) AddBaseTokens(amount uint64) *FungibleTokens {
	a.BaseTokens += amount
	return a
}

func (a *FungibleTokens) AddNativeTokens(tokenID iotago.NativeTokenID, amount interface{}) *FungibleTokens {
	b := NewFungibleTokens(0, iotago.NativeTokens{
		&iotago.NativeToken{
			ID:     tokenID,
			Amount: util.ToBigInt(amount),
		},
	})
	return a.Add(b)
}

func (a *FungibleTokens) ToDict() dict.Dict {
	ret := dict.New()
	ret.Set(kv.Key(BaseTokenID), new(big.Int).SetUint64(a.BaseTokens).Bytes())
	for _, nativeToken := range a.NativeTokens {
		ret.Set(kv.Key(nativeToken.ID[:]), nativeToken.Amount.Bytes())
	}
	return ret
}

func nativeTokensFromSet(nativeTokenSet iotago.NativeTokensSet) iotago.NativeTokens {
	ret := make(iotago.NativeTokens, len(nativeTokenSet))
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

var NativeAssetsSerializationArrayRules = iotago.NativeTokenArrayRules()

// TODO this could be refactored to use `AmountNativeToken`
func FindNativeTokenBalance(nts iotago.NativeTokens, id *iotago.NativeTokenID) *big.Int {
	for _, nt := range nts {
		if nt.ID == *id {
			return nt.Amount
		}
	}
	return nil
}
