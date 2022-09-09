package util

import iotago "github.com/iotaledger/iota.go/v3"

func NFTIDFromNFTOutput(out *iotago.NFTOutput, outID iotago.OutputID) iotago.NFTID {
	if out.NFTID.Empty() {
		// NFT outputs might not have an NFTID defined yet (when initially minted, the NFTOutput will have an empty NFTID, so we need to compute it)
		return iotago.NFTIDFromOutputID(outID)
	}
	return out.NFTID
}
