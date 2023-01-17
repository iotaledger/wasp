package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

type MintNFTsTransactionParams struct {
	IssuerKeyPair      *cryptolib.KeyPair
	CollectionOutputID *iotago.OutputID
	Target             iotago.Address
	ImmutableMetadata  [][]byte
	UnspentOutputs     iotago.OutputSet
	UnspentOutputIDs   iotago.OutputIDs
}

func NewMintNFTsTransaction(par MintNFTsTransactionParams) (*iotago.Transaction, error) {
	senderAddress := par.IssuerKeyPair.Address()

	storageDeposit := uint64(0)
	var outputs iotago.Outputs

	var issuerAddress iotago.Address = senderAddress
	nftsOut := make(map[iotago.NFTID]bool)

	addOutput := func(out *iotago.NFTOutput) {
		d := parameters.L1().Protocol.RentStructure.MinRent(out)
		out.Amount = d
		storageDeposit += d

		outputs = append(outputs, out)
	}

	if par.CollectionOutputID != nil {
		collectionOutputID := *par.CollectionOutputID
		collectionOutput := par.UnspentOutputs[*par.CollectionOutputID].(*iotago.NFTOutput)
		collectionID := util.NFTIDFromNFTOutput(collectionOutput, collectionOutputID)
		issuerAddress = collectionID.ToAddress()
		nftsOut[collectionID] = true

		out := collectionOutput.Clone().(*iotago.NFTOutput)
		out.NFTID = collectionID
		addOutput(out)
	}

	for _, immutableMetadata := range par.ImmutableMetadata {
		addOutput(&iotago.NFTOutput{
			NFTID: iotago.NFTID{},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: par.Target},
			},
			ImmutableFeatures: iotago.Features{
				&iotago.IssuerFeature{Address: issuerAddress},
				&iotago.MetadataFeature{Data: immutableMetadata},
			},
		})
	}

	inputIDs, remainder, err := computeInputsAndRemainder(senderAddress, storageDeposit, nil, nftsOut, par.UnspentOutputs, par.UnspentOutputIDs)
	if err != nil {
		return nil, err
	}
	if remainder != nil {
		outputs = append(outputs, remainder)
	}

	inputsCommitment := inputIDs.OrderedSet(par.UnspentOutputs).MustCommitment()
	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, par.IssuerKeyPair, parameters.L1().Protocol.NetworkID())
}
