package vmtxbuilder

import (
	"bytes"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
)

type nftIncluded struct {
	ID                 iotago.NFTID
	dustDepositCharged bool
	input              iotago.UTXOInput  // if in != nil
	in                 *iotago.NFTOutput // if nil it means NFT does not exist and will be created
	out                *iotago.NFTOutput // current balance of the token_id on the chain
}

// 3 cases of handling NFTs in txbuilder
// - NFT comes in
// - NFT goes out
// - NFT comes in and goes out in the same block
// all cases need 1 input and 1 output, but in the last case we don't need to keep the "accounting" for the NFT

func (n *nftIncluded) clone() *nftIncluded {
	return &nftIncluded{
		ID:                 n.ID,
		dustDepositCharged: n.dustDepositCharged,
		input:              n.input,
		in:                 cloneInternalNFTOutputOrNil(n.in),
		out:                cloneInternalNFTOutputOrNil(n.out),
	}
}

func cloneInternalNFTOutputOrNil(o *iotago.NFTOutput) *iotago.NFTOutput {
	if o == nil {
		return nil
	}
	return o.Clone().(*iotago.NFTOutput)
}

func (txb *AnchorTransactionBuilder) nftsSorted() []*nftIncluded {
	ret := make([]*nftIncluded, 0, len(txb.nftsIncluded))
	for _, nft := range txb.nftsIncluded {
		ret = append(ret, nft)
	}
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i].ID[:], ret[j].ID[:]) == -1
	})
	return ret
}

func (txb *AnchorTransactionBuilder) NFTOutputsToBeUpdated() (toBeAdded, toBeRemoved []*iotago.NFTOutput) {
	toBeAdded = make([]*iotago.NFTOutput, 0, len(txb.nftsIncluded))
	toBeRemoved = make([]*iotago.NFTOutput, 0, len(txb.nftsIncluded))
	txb.inputs()
	for _, nft := range txb.nftsSorted() {
		// TODO

		// to add if input is nil (doesn't exist in accounting), and its not sent outside the chain
		// to remove if input is not nil (exists in accounting), and its sent to outside the chain
		// do nothing if input is nil (doesn't exist in accounting) and its sent outside (comes in and leaves on the same block)
		if nft.producesOutput() {
			toBeAdded = append(toBeAdded, nft.out)
		} else if nft.requiresInput() {
			toBeRemoved = append(toBeRemoved, nft.out)
		}
	}
	return toBeAdded, toBeRemoved
}
