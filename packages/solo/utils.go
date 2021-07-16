package solo

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GrantDeployPermission gives permission to the specified agentID to deploy SCs into the chain
func (ch *Chain) GrantDeployPermission(keyPair *ed25519.KeyPair, deployerAgentID iscp.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorKeyPair
	}

	req := NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name, root.ParamDeployer, deployerAgentID)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// RevokeDeployPermission removes permission of the specified agentID to deploy SCs into the chain
func (ch *Chain) RevokeDeployPermission(keyPair *ed25519.KeyPair, deployerAgentID iscp.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorKeyPair
	}

	req := NewCallParams(root.Contract.Name, root.FuncRevokeDeployPermission.Name, root.ParamDeployer, deployerAgentID)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

func (ch *Chain) ContractAgentID(name string) *iscp.AgentID {
	return iscp.NewAgentID(ch.ChainID.AsAddress(), iscp.Hn(name))
}
