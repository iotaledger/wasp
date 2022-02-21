package vmtxbuilder

import (
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

func (txb *AnchorTransactionBuilder) nftsSorted() []*nftIncluded {
	ret := make([]*nftIncluded, 0, len(txb.invokedFoundries))
	for _, f := range txb.nftsIncluded {
		if !f.requiresInput() && !f.producesOutput() {
			continue
		}
		ret = append(ret, f)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].serialNumber < ret[j].serialNumber
	})
	return ret
}

func (txb *AnchorTransactionBuilder) NFTOutputsToBeUpdated() (toBeAdded, toBeRemoved []*iotago.NFTOutput) {
	toBeAdded = make([]iotago.NFTID, 0, len(txb.nftsIncluded))
	toBeRemoved = make([]iotago.NFTID, 0, len(txb.nftsIncluded))
	txb.inputs()
	for _, nt := range txb.nativeTokenOutputsSorted() {
		if nt.producesOutput() {
			toBeAdded = append(toBeAdded, nt.tokenID)
		} else if nt.requiresInput() {
			toBeRemoved = append(toBeRemoved, nt.tokenID)
		}
	}
	return toBeAdded, toBeRemoved
}
