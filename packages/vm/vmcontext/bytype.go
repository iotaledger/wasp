package vmcontext

import "github.com/iotaledger/wasp/packages/iscp/requestdata"

func (vmctx *VMContext) checkReplay(req requestdata.RequestData, requestIndex uint16) bool {
	// checks replay in the state.
	// - off ledger checks as usual with requestID in blocklog receipts + nonce
	// - extended outputs checks:
	// --- if extended output is produced by the chain itself (sender field) then special flag in metadata indicates if it is account or request
	// --- if it is external, checked in the blocklog receipts
	// - NFT output checks if NFT is known already in the NFT table
	// - Foundry output checks if NFT is known already in the NFT table.
	// --- If it is not known it is ignored
	// ---
	_ = req.Request().ID()
}

func (vmctx *VMContext) preprocessRequestData(req requestdata.RequestData, requestIndex uint16) bool {
	switch req.Type() {
	case requestdata.TypeSimpleOutput:
		// consume it an assign all assets to owner's account
		// no need to invoke SC

		return true
	case requestdata.TypeOffLedger:
		// prepare off ledger
		return false
	case requestdata.TypeExtendedOutput:
		// prepare extended
		return false
	case requestdata.TypeNFTOutput:
		// prepare NFT request
		return false
	case requestdata.TypeFoundryOutput:
		// do not consume. Check consistency in the state
		// no need to invoke SC
		return true
	case requestdata.TypeAliasOutput:
		// do not consume. It is unexpected.
		// assign ownership to the owner
		// no need to invoke SC
		return true
	case requestdata.TypeUnknownOutput:
		// do not consume.
		// Assign ownership to the owner
		// no need to invoke SC
		return true
	case requestdata.TypeUnknown:
		// an error. probably panic
		// no need to invoke SC
		return true
	}
	panic("wrong request data type")
}

func (vmctx *VMContext) preprocessSimple() {}
