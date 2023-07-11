package isc

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Assets struct {
	BaseTokens   uint64              `json:"baseTokens"`
	NativeTokens iotago.NativeTokens `json:"nativeTokens"`
	NFTs         []iotago.NFTID      `json:"nfts"`
}

var BaseTokenID = []byte{}

func NewAssets(baseTokens uint64, tokens iotago.NativeTokens, nfts ...iotago.NFTID) *Assets {
	ret := &Assets{
		BaseTokens:   baseTokens,
		NativeTokens: tokens,
	}
	if len(nfts) != 0 {
		ret.AddNFTs(nfts...)
	}
	return ret
}

func NewAssetsBaseTokens(amount uint64) *Assets {
	return &Assets{BaseTokens: amount}
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
			ret.BaseTokens = new(big.Int).SetBytes(d.Get(kv.Key(BaseTokenID))).Uint64()
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

func AssetsFromOutput(o iotago.Output) *Assets {
	ret := &Assets{
		BaseTokens:   o.Deposit(),
		NativeTokens: o.NativeTokenList().Clone(),
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
	return rwutil.WriteToBytes(a)
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
		if !bNFTS[nft] {
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
		a.NFTs = nil
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

	// see if there is an empty NFTID in the assets (this can happen if the NTF is minted as a request to the chain)
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
		a.BaseTokens = rr.ReadAmount64()
	}
	if (flags & hasNativeTokens) != 0 {
		size := rr.ReadSize16()
		a.NativeTokens = make(iotago.NativeTokens, size)
		for i := range a.NativeTokens {
			nativeToken := new(iotago.NativeToken)
			a.NativeTokens[i] = nativeToken
			rr.ReadN(nativeToken.ID[:])
			nativeToken.Amount = rr.ReadUint256()
		}
	}
	if (flags & hasNFTs) != 0 {
		size := rr.ReadSize16()
		a.NFTs = make([]iotago.NFTID, size)
		for i := range a.NFTs {
			rr.ReadN(a.NFTs[i][:])
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
	if a.BaseTokens != 0 {
		flags |= hasBaseTokens
	}
	if len(a.NativeTokens) != 0 {
		flags |= hasNativeTokens
	}
	if len(a.NFTs) != 0 {
		flags |= hasNFTs
	}

	ww.WriteByte(flags)
	if (flags & hasBaseTokens) != 0 {
		ww.WriteAmount64(a.BaseTokens)
	}
	if (flags & hasNativeTokens) != 0 {
		ww.WriteSize16(len(a.NativeTokens))
		sort.Slice(a.NativeTokens, func(lhs, rhs int) bool {
			return bytes.Compare(a.NativeTokens[lhs].ID[:], a.NativeTokens[rhs].ID[:]) < 0
		})
		for _, nativeToken := range a.NativeTokens {
			ww.WriteN(nativeToken.ID[:])
			ww.WriteUint256(nativeToken.Amount)
		}
	}
	if (flags & hasNFTs) != 0 {
		ww.WriteSize16(len(a.NFTs))
		sort.Slice(a.NFTs, func(lhs, rhs int) bool {
			return bytes.Compare(a.NFTs[lhs][:], a.NFTs[rhs][:]) < 0
		})
		for _, nft := range a.NFTs {
			ww.WriteN(nft[:])
		}
	}
	return ww.Err
}
