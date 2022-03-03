package vmtxbuilder

import (
	"bytes"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

type nftIncluded struct {
	ID iotago.NFTID
	// dustDepositCharged bool
	input       iotago.UTXOInput  // if in != nil
	in          *iotago.NFTOutput // if nil it means NFT does not exist and will be created
	out         *iotago.NFTOutput // NFT output (can be used by the chain)
	sentOutside bool
}

// 3 cases of handling NFTs in txbuilder
// - NFT comes in
// - NFT goes out
// - NFT comes in and goes out in the same block
// all cases need 1 input and 1 output, but in the last case we don't need to keep the "accounting" for the NFT

func (n *nftIncluded) clone() *nftIncluded {
	return &nftIncluded{
		ID: n.ID,
		// dustDepositCharged: n.dustDepositCharged,
		input: n.input,
		in:    cloneInternalNFTOutputOrNil(n.in),
		out:   cloneInternalNFTOutputOrNil(n.out),
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

func (txb *AnchorTransactionBuilder) NFTOutputs() []*iotago.NFTOutput {
	outs := make([]*iotago.NFTOutput, len(txb.nftsIncluded))
	for i, nft := range txb.nftsSorted() {
		outs[i] = nft.out
	}
	return outs
}

func (txb *AnchorTransactionBuilder) NFTOutputsToBeUpdated() (toBeAdded, toBeRemoved []*iotago.NFTOutput) {
	toBeAdded = make([]*iotago.NFTOutput, 0, len(txb.nftsIncluded))
	toBeRemoved = make([]*iotago.NFTOutput, 0, len(txb.nftsIncluded))
	txb.inputs()
	for _, nft := range txb.nftsSorted() {
		if nft.in != nil {
			// to remove if input is not nil (nft exists in accounting), and its sent to outside the chain
			toBeRemoved = append(toBeRemoved, nft.out)
			continue
		}
		if nft.sentOutside {
			// do nothing if input is nil (doesn't exist in accounting) and its sent outside (comes in and leaves on the same block)
			continue
		}
		// to add if input is nil (doesn't exist in accounting), and its not sent outside the chain
		toBeAdded = append(toBeAdded, nft.out)
	}
	return toBeAdded, toBeRemoved
}

func (txb *AnchorTransactionBuilder) consumeNFT(o *iotago.NFTOutput) int64 {
	dustDeposit := int64(txb.dustDepositAssumption.NFTOutput)

	// keep the number of iotas in the output == required dust deposit
	// take all native tokens out of the NFT output
	txb.subDeltaIotasFromTotal(txb.dustDepositAssumption.NFTOutput)

	out := o.Clone().(*iotago.NFTOutput)
	out.Amount = uint64(dustDeposit)
	out.NativeTokens = nil
	out.Conditions = iotago.UnlockConditions{
		&iotago.AddressUnlockCondition{
			Address: txb.anchorOutput.AliasID.ToAddress(),
		},
	}

	toInclude := &nftIncluded{
		ID:          o.NFTID,
		in:          nil,
		out:         out,
		sentOutside: false,
	}

	txb.nftsIncluded[o.NFTID] = toInclude
	return -dustDeposit
}

func (txb *AnchorTransactionBuilder) sendNFT(o *iotago.NFTOutput) int64 {
	if txb.nftsIncluded[o.NFTID] != nil {
		// NFT comes in and out in the same block
		txb.nftsIncluded[o.NFTID].sentOutside = true
		txb.nftsIncluded[o.NFTID].out = o
		return 0
	}

	if txb.InputsAreFull() {
		panic(vmexceptions.ErrInputLimitExceeded)
	}
	if txb.outputsAreFull() {
		panic(vmexceptions.ErrOutputLimitExceeded)
	}

	// using NFT already owned by the chain
	in, input := txb.loadNFTOutput(&o.NFTID)
	toInclude := &nftIncluded{
		ID:          o.NFTID,
		in:          in,
		out:         o,
		sentOutside: false,
	}

	if input != nil {
		toInclude.input = *input
	}
	txb.nftsIncluded[o.NFTID] = toInclude
	return int64(txb.dustDepositAssumption.NFTOutput)
}
