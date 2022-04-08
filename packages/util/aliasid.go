package util

import iotago "github.com/iotaledger/iota.go/v3"

func AliasIDFromAliasOutput(out *iotago.AliasOutput, outID iotago.OutputID) iotago.AliasID {
	if out.AliasID.Empty() {
		// NFT outputs might not have an NFTID defined yet (when initially minted, the NFTOutput will have an empty NFTID, so we need to compute it)
		return iotago.AliasIDFromOutputID(outID)
	} else {
		return out.AliasID
	}
}
