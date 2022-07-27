package solo

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GrantDeployPermission gives permission to the specified agentID to deploy SCs into the chain
func (ch *Chain) GrantDeployPermission(keyPair *cryptolib.KeyPair, deployerAgentID isc.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}

	req := NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name, root.ParamDeployer, deployerAgentID).AddBaseTokens(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// RevokeDeployPermission removes permission of the specified agentID to deploy SCs into the chain
func (ch *Chain) RevokeDeployPermission(keyPair *cryptolib.KeyPair, deployerAgentID isc.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}

	req := NewCallParams(root.Contract.Name, root.FuncRevokeDeployPermission.Name, root.ParamDeployer, deployerAgentID).AddBaseTokens(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

func (ch *Chain) ContractAgentID(name string) isc.AgentID {
	return isc.NewContractAgentID(ch.ChainID, isc.Hn(name))
}

// Warning: if the same `req` is passed in different occasions, the resulting request will have different IDs (because the ledger state is different)
func NewIscpRequestFromCallParams(ch *Chain, req *CallParams, keyPair *cryptolib.KeyPair) (isc.Request, error) {
	tx, _, err := ch.RequestFromParamsToLedger(req, keyPair)
	if err != nil {
		return nil, err
	}
	requestsFromSignedTx, err := isc.RequestsInTransaction(tx)
	if err != nil {
		return nil, err
	}
	return requestsFromSignedTx[*ch.ChainID][0], nil
}
