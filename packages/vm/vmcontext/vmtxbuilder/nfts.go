package vmtxbuilder

import (
	"bytes"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

type nftIncluded struct {
	ID          iotago.NFTID
	input       *iotago.UTXOInput // only available when the input is already accounted for (NFT was deposited in a previous block)
	in          *iotago.NFTOutput
	out         *iotago.NFTOutput
	sentOutside bool
}

// 3 cases of handling NFTs in txbuilder
// - NFT comes in
// - NFT goes out
// - NFT comes in and goes out in the same block
// all cases need 1 input and 1 output, but in the last case we don't need to keep the "accounting" for the NFT

func (n *nftIncluded) clone() *nftIncluded {
	return &nftIncluded{
		ID:    n.ID,
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
	outs := make([]*iotago.NFTOutput, 0)
	for _, nft := range txb.nftsSorted() {
		if !nft.sentOutside {
			// outputs sent outside are already added to txb.postedOutputs
			outs = append(outs, nft.out)
		}
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

func (txb *AnchorTransactionBuilder) consumeNFT(o *iotago.NFTOutput, utxoInput iotago.UTXOInput) int64 {
	dustDeposit := int64(txb.dustDepositAssumption.NFTOutput)

	// keep the number of base tokens in the output == required dust deposit
	// take all native tokens out of the NFT output
	txb.subDeltaBaseTokensFromTotal(txb.dustDepositAssumption.NFTOutput)

	out := o.Clone().(*iotago.NFTOutput)
	out.Amount = uint64(dustDeposit)
	chainAddr := txb.anchorOutput.AliasID.ToAddress()
	out.NativeTokens = nil
	out.Conditions = iotago.UnlockConditions{
		&iotago.AddressUnlockCondition{
			Address: chainAddr,
		},
	}
	out.Features = iotago.Features{
		&iotago.SenderFeature{
			Address: chainAddr,
		},
	}

	if out.NFTID.Empty() {
		// nft was just minted to the chain
		out.NFTID = iotago.NFTIDFromOutputID(utxoInput.ID())
	}

	toInclude := &nftIncluded{
		ID:          out.NFTID,
		in:          nil,
		out:         out,
		sentOutside: false,
	}

	txb.nftsIncluded[o.NFTID] = toInclude
	return -dustDeposit
}

func (txb *AnchorTransactionBuilder) sendNFT(o *iotago.NFTOutput) int64 {
	if txb.outputsAreFull() {
		panic(vmexceptions.ErrOutputLimitExceeded)
	}

	if txb.nftsIncluded[o.NFTID] != nil {
		// NFT comes in and out in the same block
		txb.nftsIncluded[o.NFTID].sentOutside = true
		txb.nftsIncluded[o.NFTID].out = o
	} else {
		if txb.InputsAreFull() {
			panic(vmexceptions.ErrInputLimitExceeded)
		}

		// using NFT already owned by the chain
		in, input := txb.loadNFTOutput(o.NFTID)
		toInclude := &nftIncluded{
			ID:          o.NFTID,
			in:          in,
			input:       input,
			out:         o,
			sentOutside: true,
		}

		txb.nftsIncluded[o.NFTID] = toInclude
	}
	txb.addDeltaBaseTokensToTotal(txb.dustDepositAssumption.NFTOutput)
	return int64(txb.dustDepositAssumption.NFTOutput)
}
