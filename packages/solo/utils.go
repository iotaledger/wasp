package solo

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GrantDeployPermission gives permission to the specified agentID to deploy SCs into the chain
func (ch *Chain) GrantDeployPermission(sigScheme signaturescheme.SignatureScheme, deployerAgentID coretypes.AgentID) error {
	if sigScheme == nil {
		sigScheme = ch.OriginatorSigScheme
	}

	req := NewCallParams(root.Interface.Name, root.FuncGrantDeploy, root.ParamDeployer, deployerAgentID)
	_, err := ch.PostRequestSync(req, sigScheme)
	return err
}

// RevokeDeployPermission removes permission of the specified agentID to deploy SCs into the chain
func (ch *Chain) RevokeDeployPermission(sigScheme signaturescheme.SignatureScheme, deployerAgentID coretypes.AgentID) error {
	if sigScheme == nil {
		sigScheme = ch.OriginatorSigScheme
	}

	req := NewCallParams(root.Interface.Name, root.FuncRevokeDeploy, root.ParamDeployer, deployerAgentID)
	_, err := ch.PostRequestSync(req, sigScheme)
	return err
}
