package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
)

type MintNFTTransactionParams struct {
	IssuerKeyPair     *cryptolib.KeyPair
	Collection        *iotago.NFTOutput
	Target            iotago.Address
	ImmutableMetadata [][]byte
	UnspentOutputs    iotago.OutputSet
	UnspentOutputIDs  iotago.OutputIDs
}

func NewMintNFTsTransaction(par MintNFTTransactionParams) (*iotago.Transaction, error) {
	senderAddress := par.IssuerKeyPair.Address()

	var issuerAddress iotago.Address = senderAddress
	var nftsOut map[iotago.NFTID]bool
	if par.Collection != nil {
		issuerAddress = par.Collection.NFTID.ToAddress()
		nftsOut[par.Collection.NFTID] = true
	}

	storageDeposit := uint64(0)
	var outputs iotago.Outputs

	for _, immutableMetadata := range par.ImmutableMetadata {
		out := &iotago.NFTOutput{
			NFTID: iotago.NFTID{},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: par.Target},
			},
			ImmutableFeatures: iotago.Features{
				&iotago.IssuerFeature{Address: issuerAddress},
				&iotago.MetadataFeature{Data: immutableMetadata},
			},
		}

		d := parameters.L1().Protocol.RentStructure.MinRent(out)
		out.Amount = d
		storageDeposit += d

		outputs = append(outputs, out)
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
