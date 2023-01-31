package isc

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type Assets struct {
	BaseTokens   uint64              `json:"baseTokens"`
	NativeTokens iotago.NativeTokens `json:"nativeTokens"`
	NFTs         []iotago.NFTID      `json:"nfts"`
}

var BaseTokenID = []byte{}

func NewEmptyAssets() *Assets {
	return &Assets{
		NativeTokens: make([]*iotago.NativeToken, 0),
	}
}

func NewAssets(baseTokens uint64, tokens iotago.NativeTokens, NFTs ...iotago.NFTID) *Assets {
	if tokens == nil {
		tokens = make(iotago.NativeTokens, 0)
	}
	ret := &Assets{
		BaseTokens:   baseTokens,
		NativeTokens: tokens,
		NFTs:         make([]iotago.NFTID, 0),
	}
	return ret.AddNFTs(NFTs...)
}

func NewAssetsBaseTokens(amount uint64) *Assets {
	return &Assets{BaseTokens: amount}
}

func NewAssetsForGasFee(p *gas.GasFeePolicy, feeAmount uint64) *Assets {
	if IsEmptyNativeTokenID(p.GasFeeTokenID) {
		return NewAssetsBaseTokens(feeAmount)
	}
	return NewEmptyAssets().AddNativeTokens(p.GasFeeTokenID, feeAmount)
}

func AssetsFromDict(d dict.Dict) (*Assets, error) {
	ret := NewEmptyAssets()
	for key, val := range d {
		if IsBaseToken([]byte(key)) {
			ret.BaseTokens = new(big.Int).SetBytes(d.MustGet(kv.Key(BaseTokenID))).Uint64()
			continue
		}
		id, err := NativeTokenIDFromBytes([]byte(key))
		if err != nil {
			return nil, fmt.Errorf("Assets: %w", err)
		}
		token := &iotago.NativeToken{
			ID:     id,
			Amount: new(big.Int).SetBytes(val),
		}
		ret.NativeTokens = append(ret.NativeTokens, token)
	}
	return ret, nil
}

func AssetsFromNativeTokenSum(baseTokens uint64, tokens iotago.NativeTokenSum) *Assets {
	ret := NewEmptyAssets()
	ret.BaseTokens = baseTokens
	for id, val := range tokens {
		ret.NativeTokens = append(ret.NativeTokens, &iotago.NativeToken{
			ID:     id,
			Amount: val,
		})
	}
	return ret
}

func AssetsFromOutputMap(outs map[iotago.OutputID]iotago.Output) *Assets {
	ret := NewEmptyAssets()
	for _, out := range outs {
		ret.Add(AssetsFromOutput(out))
	}
	return ret
}

func AssetsFromOutput(o iotago.Output) *Assets {
	ret := &Assets{
		BaseTokens:   o.Deposit(),
		NativeTokens: o.NativeTokenList().Clone(),
	}
	return ret
}

func NativeTokenIDFromBytes(data []byte) (iotago.NativeTokenID, error) {
	if len(data) != iotago.NativeTokenIDLength {
		return iotago.NativeTokenID{}, errors.New("NativeTokenIDFromBytes: wrong data length")
	}
	var nativeTokenID iotago.NativeTokenID
	copy(nativeTokenID[:], data)
	return nativeTokenID, nil
}

func MustNativeTokenIDFromBytes(data []byte) iotago.NativeTokenID {
	ret, err := NativeTokenIDFromBytes(data)
	if err != nil {
		panic(fmt.Errorf("MustNativeTokenIDFromBytes: %w", err))
	}
	return ret
}

// returns nil if nil pointer receiver is cloned
func (a *Assets) Clone() *Assets {
	if a == nil {
		return nil
	}

	nfts := make([]iotago.NFTID, len(a.NFTs))
	for i := range a.NFTs {
		nftID := iotago.NFTID{}
		copy(nftID[:], a.NFTs[i][:])
		nfts[i] = nftID
	}

	return &Assets{
		BaseTokens:   a.BaseTokens,
		NativeTokens: a.NativeTokens.Clone(),
		NFTs:         nfts,
	}
}

func (a *Assets) AddNFTs(nfts ...iotago.NFTID) *Assets {
	nftMap := make(map[iotago.NFTID]bool)
	nfts = append(nfts, a.NFTs...)
	for _, nftid := range nfts {
		nftMap[nftid] = true
	}
	a.NFTs = make([]iotago.NFTID, len(nftMap))
	i := 0
	for nftid := range nftMap {
		a.NFTs[i] = nftid
		i++
	}
	return a
}

func (a *Assets) AmountNativeToken(nativeTokenID iotago.NativeTokenID) *big.Int {
	for _, t := range a.NativeTokens {
		if t.ID == nativeTokenID {
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
		ret += fmt.Sprintf("\n       %s: %d", nt.ID.String(), nt.Amount)
	}
	for _, nftid := range a.NFTs {
		ret += fmt.Sprintf("\n NFTID: %s", nftid.String())
	}
	return ret
}

func (a *Assets) Bytes() []byte {
	mu := marshalutil.New()
	a.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

var NativeAssetsSerializationArrayRules = iotago.NativeTokenArrayRules()

func (a *Assets) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteBool(a.IsEmpty())
	if a.IsEmpty() {
		return
	}
	mu.WriteUint64(a.BaseTokens)
	tokenBytes, err := serializer.NewSerializer().WriteSliceOfObjects(&a.NativeTokens, serializer.DeSeriModePerformLexicalOrdering, nil, serializer.SeriLengthPrefixTypeAsUint16, &NativeAssetsSerializationArrayRules, func(err error) error {
		return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
	}).Serialize()
	if err != nil {
		panic(fmt.Errorf("unexpected error serializing native tokens: %w", err))
	}
	mu.WriteUint16(uint16(len(tokenBytes)))
	mu.WriteBytes(tokenBytes)
	mu.WriteUint16(uint16(len(a.NFTs)))
	for _, id := range a.NFTs {
		mu.WriteBytes(id[:])
	}
}

func MustAssetsFromBytes(b []byte) *Assets {
	if len(b) == 0 {
		return NewEmptyAssets()
	}
	ret, err := AssetsFromMarshalUtil(marshalutil.New(b))
	if err != nil {
		panic(err)
	}
	return ret
}

func AssetsFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Assets, error) {
	ret := NewEmptyAssets()
	empty, err := mu.ReadBool()
	if err != nil {
		return nil, err
	}
	if empty {
		return ret, nil
	}
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
	nNFTs, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	ret.NFTs = make([]iotago.NFTID, nNFTs)
	for i := 0; i < int(nNFTs); i++ {
		b, err := mu.ReadBytes(iotago.NFTIDLength)
		if err != nil {
			return nil, err
		}
		var id iotago.NFTID
		copy(id[:], b)
		ret.NFTs[i] = id
	}
	return ret, nil
}

func (a *Assets) Equals(b *Assets) bool {
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
	if len(a.NFTs) != len(b.NFTs) {
		return false
	}

	bNFTS := b.NFTSet()
	for _, nft := range a.NFTs {
		if ok, _ := bNFTS[nft]; !ok {
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
		a.BaseTokens = 0
		a.NativeTokens = nil
		a.NFTs = make([]iotago.NFTID, 0)
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

	nftSet := a.NFTSet()
	for _, nftID := range toSpend.NFTs {
		if !nftSet[nftID] {
			return false
		}
		delete(nftSet, nftID)
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
	a.NFTs = make([]iotago.NFTID, len(nftSet))
	i := 0
	for nftID := range nftSet {
		a.NFTs[i] = nftID
		i++
	}
	return true
}

func (a *Assets) NFTSet() map[iotago.NFTID]bool {
	ret := map[iotago.NFTID]bool{}
	for _, nft := range a.NFTs {
		ret[nft] = true
	}
	return ret
}

func (a *Assets) Add(b *Assets) *Assets {
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
	a.AddNFTs(b.NFTs...)
	return a
}

func (a *Assets) IsEmpty() bool {
	return a == nil || (a.BaseTokens == 0 &&
		len(a.NativeTokens) == 0 &&
		len(a.NFTs) == 0)
}

func (a *Assets) AddBaseTokens(amount uint64) *Assets {
	a.BaseTokens += amount
	return a
}

func (a *Assets) AddNativeTokens(nativeTokenID iotago.NativeTokenID, amount interface{}) *Assets {
	b := NewAssets(0, iotago.NativeTokens{
		&iotago.NativeToken{
			ID:     nativeTokenID,
			Amount: util.ToBigInt(amount),
		},
	})
	return a.Add(b)
}

func (a *Assets) ToDict() dict.Dict {
	ret := dict.New()
	ret.Set(kv.Key(BaseTokenID), new(big.Int).SetUint64(a.BaseTokens).Bytes())
	for _, nativeToken := range a.NativeTokens {
		ret.Set(kv.Key(nativeToken.ID[:]), nativeToken.Amount.Bytes())
	}
	return ret
}

func (a *Assets) fillEmptyNFTIDs(output iotago.Output, outputID iotago.OutputID) *Assets {
	if a == nil {
		return nil
	}

	nftOutput, ok := output.(*iotago.NFTOutput)
	if !ok {
		return a
	}

	// see if there is an empty NFTID in the assets (this can happpen if the NTF is minted as a request to the chain)
	for i, nftID := range a.NFTs {
		if nftID.Empty() {
			a.NFTs[i] = util.NFTIDFromNFTOutput(nftOutput, outputID)
		}
	}
	return a
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
