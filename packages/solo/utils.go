package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GrantDeployPermission gives permission to the specified agentID to deploy SCs into the chain
func (ch *Chain) GrantDeployPermission(keyPair *ed25519.KeyPair, deployerAgentID coretypes.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorKeyPair
	}

	req := NewCallParams(root.Interface.Name, root.FuncGrantDeploy, root.ParamDeployer, deployerAgentID)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// RevokeDeployPermission removes permission of the specified agentID to deploy SCs into the chain
func (ch *Chain) RevokeDeployPermission(keyPair *ed25519.KeyPair, deployerAgentID coretypes.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorKeyPair
	}

	req := NewCallParams(root.Interface.Name, root.FuncRevokeDeploy, root.ParamDeployer, deployerAgentID)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

func CloneBalances(m map[ledgerstate.Color]uint64) map[ledgerstate.Color]uint64 {
	ret := make(map[ledgerstate.Color]uint64)
	for c, b := range m {
		ret[c] = b
	}
	return ret
}
