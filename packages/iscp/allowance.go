package iscp

import (
	"fmt"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
)

type Allowance struct {
	Assets *Assets
	NFTs   []iotago.NFTID
}

func NewEmptyAllowance() *Allowance {
	return &Allowance{
		Assets: NewEmptyAssets(),
		NFTs:   make([]iotago.NFTID, 0),
	}
}

func NewAllowance(iotas uint64, tokens iotago.NativeTokens, NFTs []iotago.NFTID) *Allowance {
	return &Allowance{
		Assets: NewAssets(iotas, tokens),
		NFTs:   NFTs,
	}
}

func NewAllowanceIotas(iotas uint64) *Allowance {
	return NewAllowance(iotas, nil, nil)
}

func NewAllowanceFungibleTokens(ftokens *Assets) *Allowance {
	return &Allowance{
		Assets: ftokens,
	}
}

func (a *Allowance) Clone() *Allowance {
	if a == nil {
		return nil
	}
	nfts := make([]iotago.NFTID, len(a.NFTs))
	for i, nft := range a.NFTs {
		id := nft
		nfts[i] = id
	}
	return &Allowance{
		Assets: a.Assets.Clone(),
		NFTs:   nfts,
	}
}

func (a *Allowance) SpendFromBudget(toSpend *Allowance) bool {
	a.Assets.SpendFromBudget(toSpend.Assets)
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

func (a *Allowance) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	a.Assets.WriteToMarshalUtil(mu)
	mu.WriteUint16(uint16(len(a.NFTs)))
	for _, id := range a.NFTs {
		mu.WriteBytes(id[:])
	}
}

func AllowanceFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Allowance, error) {
	assets, err := AssetsFromMarshalUtil(mu)
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
	a.NFTs = append(a.NFTs, b.NFTs...)
	return a
}

func (a *Allowance) AddIotas(amount uint64) *Allowance {
	a.Assets.Iotas += amount
	return a
}

func (a *Allowance) AddNativeTokens(tokenID iotago.NativeTokenID, amount interface{}) *Allowance {
	a.Assets.AddNativeTokens(tokenID, amount)
	return a
}

func (a *Allowance) AddNFTs(nfts ...iotago.NFTID) *Allowance {
	a.NFTs = make([]iotago.NFTID, len(nfts))
	copy(a.NFTs, nfts)
	return a
}

func (a *Allowance) String() string {
	ret := a.Assets.String()
	for _, nftid := range a.NFTs {
		ret += fmt.Sprintf("\n NFTID: %s", nftid.String())
	}
	return ret
}

func (a *Allowance) fillEmptyNFTIDs(o iotago.Output, utxoInput *iotago.UTXOInput) *Allowance {
	if a == nil {
		return nil
	}
	nftOut, ok := o.(*iotago.NFTOutput)
	if !ok {
		return a
	}
	// see if there is an empty NFTID in allowance (this can happpen if the NTF is minted as a request to the chain)
	for i, nft := range a.NFTs {
		if nft.Empty() {
			a.NFTs[i] = util.NFTIDFromNFTOutput(nftOut, utxoInput.ID())
		}
	}
	return a
}
