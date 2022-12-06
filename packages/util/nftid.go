package util

import iotago "github.com/iotaledger/iota.go/v3"

func NFTIDFromNFTOutput(nftOutput *iotago.NFTOutput, outputID iotago.OutputID) iotago.NFTID {
	if nftOutput.NFTID.Empty() {
		// NFT outputs might not have an NFTID defined yet (when initially minted, the NFTOutput will have an empty NFTID, so we need to compute it)
		return iotago.NFTIDFromOutputID(outputID)
	}
	return nftOutput.NFTID
}
