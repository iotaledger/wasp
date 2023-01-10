package isc

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
)

type Allowance struct {
	Assets *FungibleTokens `json:"tokens"`
	NFTs   []iotago.NFTID  `json:"nfts"`
}

func NewEmptyAllowance() *Allowance {
	return &Allowance{
		Assets: NewEmptyFungibleTokens(),
		NFTs:   make([]iotago.NFTID, 0),
	}
}

func NewAllowance(baseTokens uint64, tokens iotago.NativeTokens, nfts []iotago.NFTID) *Allowance {
	return &Allowance{
		Assets: NewFungibleTokens(baseTokens, tokens),
		NFTs:   nfts,
	}
}

func NewAllowanceBaseTokens(baseTokens uint64) *Allowance {
	return NewAllowance(baseTokens, nil, nil)
}

func NewAllowanceFungibleTokens(ftokens *FungibleTokens) *Allowance {
	return &Allowance{
		Assets: ftokens,
	}
}

// returns nil if nil pointer receiver is cloned
func (a *Allowance) Clone() *Allowance {
	if a == nil {
		return nil
	}

	nfts := make([]iotago.NFTID, len(a.NFTs))
	for i := range a.NFTs {
		nftID := iotago.NFTID{}
		copy(nftID[:], a.NFTs[i][:])
		nfts[i] = nftID
	}

	return &Allowance{
		Assets: a.Assets.Clone(),
		NFTs:   nfts,
	}
}

func (a *Allowance) SpendFromBudget(toSpend *Allowance) bool {
	if a.IsEmpty() {
		return toSpend.IsEmpty()
	}
	if !a.Assets.SpendFromFungibleTokenBudget(toSpend.Assets) {
		return false
	}
	nftSet := a.NFTSet()
	for _, id := range toSpend.NFTs {
		if !nftSet[id] {
			return false
		}
		nftSet[id] = false
	}

	tmp := a.NFTs[:0] // reuse the array
	for id, keep := range nftSet {
		cp := id // otherwise, taking pointer of loop parameter is a bug
		if keep {
			tmp = append(tmp, cp)
		}
	}
	a.NFTs = tmp

	return true
}

func AllowanceFromBytes(b []byte) (*Allowance, error) {
	return AllowanceFromMarshalUtil(marshalutil.New(b))
}

func (a *Allowance) Bytes() []byte {
	mu := marshalutil.New()
	a.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (a *Allowance) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteBool(a.IsEmpty())
	if a.IsEmpty() {
		return
	}
	a.Assets.WriteToMarshalUtil(mu)
	mu.WriteUint16(uint16(len(a.NFTs)))
	for _, id := range a.NFTs {
		mu.WriteBytes(id[:])
	}
}

func AllowanceFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Allowance, error) {
	empty, err := mu.ReadBool()
	if err != nil {
		return nil, err
	}
	if empty {
		return NewEmptyAllowance(), nil
	}
	assets, err := FungibleTokensFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	nNFTs, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	nfts := make([]iotago.NFTID, nNFTs)
	for i := 0; i < int(nNFTs); i++ {
		b, err := mu.ReadBytes(iotago.NFTIDLength)
		if err != nil {
			return nil, err
		}
		var id iotago.NFTID
		copy(id[:], b)
		nfts[i] = id
	}

	a := &Allowance{
		Assets: assets,
		NFTs:   nfts,
	}
	return a, nil
}

func (a *Allowance) NFTSet() map[iotago.NFTID]bool {
	ret := map[iotago.NFTID]bool{}
	for _, nft := range a.NFTs {
		ret[nft] = true
	}
	return ret
}

func (a *Allowance) IsEmpty() bool {
	return a == nil || (a.Assets.IsEmpty() && len(a.NFTs) == 0)
}

func (a *Allowance) Add(b *Allowance) *Allowance {
	a.Assets.Add(b.Assets)
	a.AddNFTs(b.NFTs...)
	return a
}

func (a *Allowance) AddBaseTokens(amount uint64) *Allowance {
	a.Assets.BaseTokens += amount
	return a
}

func (a *Allowance) AddNativeTokens(nativeTokenID iotago.NativeTokenID, amount interface{}) *Allowance {
	a.Assets.AddNativeTokens(nativeTokenID, amount)
	return a
}

func (a *Allowance) AddNFTs(nfts ...iotago.NFTID) *Allowance {
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

func (a *Allowance) String() string {
	ret := a.Assets.String()
	for _, nftid := range a.NFTs {
		ret += fmt.Sprintf("\n NFTID: %s", nftid.String())
	}
	return ret
}

func (a *Allowance) fillEmptyNFTIDs(output iotago.Output, outputID iotago.OutputID) *Allowance {
	if a == nil {
		return nil
	}

	nftOutput, ok := output.(*iotago.NFTOutput)
	if !ok {
		return a
	}

	// see if there is an empty NFTID in allowance (this can happpen if the NTF is minted as a request to the chain)
	for i, nftID := range a.NFTs {
		if nftID.Empty() {
			a.NFTs[i] = util.NFTIDFromNFTOutput(nftOutput, outputID)
		}
	}
	return a
}
