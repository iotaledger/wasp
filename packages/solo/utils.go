package solo

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GrantDeployPermission gives permission to the specified agentID to deploy SCs into the chain
func (ch *Chain) GrantDeployPermission(keyPair *cryptolib.KeyPair, deployerAgentID *iscp.AgentID) error {
	if keyPair == nil {
		keyPair = &ch.OriginatorPrivateKey
	}

	req := NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name, root.ParamDeployer, deployerAgentID).WithIotas(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// RevokeDeployPermission removes permission of the specified agentID to deploy SCs into the chain
func (ch *Chain) RevokeDeployPermission(keyPair *cryptolib.KeyPair, deployerAgentID *iscp.AgentID) error {
	if keyPair == nil {
		keyPair = &ch.OriginatorPrivateKey
	}

	req := NewCallParams(root.Contract.Name, root.FuncRevokeDeployPermission.Name, root.ParamDeployer, deployerAgentID).WithIotas(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

func (ch *Chain) ContractAgentID(name string) *iscp.AgentID {
	return iscp.NewAgentID(ch.ChainID.AsAddress(), iscp.Hn(name))
}
