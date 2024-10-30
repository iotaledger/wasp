package solo

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func (ch *Chain) ContractAgentID(name string) isc.AgentID {
	return isc.NewContractAgentID(ch.ChainID, isc.Hn(name))
}

// Warning: if the same `req` is passed in different occasions, the resulting request will have different IDs (because the ledger state is different)
func ISCRequestFromCallParams(ch *Chain, req *CallParams, keyPair *cryptolib.KeyPair) (isc.Request, error) {
	panic("TODO")
	/*
		reqID, err := ch.RequestFromParamsToLedger(req, keyPair)
		if err != nil {
			return nil, err
		}
		requestsFromSignedTx, err := isc.RequestsInTransaction(tx)
		if err != nil {
			return nil, err
		}
		return requestsFromSignedTx[ch.ChainID][0], nil
	*/
}
