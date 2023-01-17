package transaction

import (
	"encoding/json"

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

// IRC27NFTMetadata represents an NFT metadata according to IRC27.
// See: https://github.com/Kami-Labs/tips/blob/main/tips/TIP-0027/tip-0027.md#nft-schema
type IRC27NFTMetadata struct {
	Standard string `json:"standard"`
	Version  string `json:"version"`
	MIMEType string `json:"type"`
	URI      string `json:"uri"`
	Name     string `json:"name"`
}

func NewIRC27NFTMetadata(mimeType, uri, name string) *IRC27NFTMetadata {
	return &IRC27NFTMetadata{
		Standard: "IRC27",
		Version:  "v1.0",
		MIMEType: mimeType,
		URI:      uri,
		Name:     name,
	}
}

func (m *IRC27NFTMetadata) Bytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *IRC27NFTMetadata) MustBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func IRC27NFTMetadataFromBytes(b []byte) (*IRC27NFTMetadata, error) {
	var m IRC27NFTMetadata
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
