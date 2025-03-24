package validation

import (
	"fmt"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

const (
	newSchema = allmigrations.SchemaVersionMigratedRebased
)

func oldAgentIDToStr(agentID old_isc.AgentID) string {
	switch agentID := agentID.(type) {
	case *old_isc.AddressAgentID:
		return fmt.Sprintf("AddressAgentID(%v)", agentID.Address().String())
	case *old_isc.ContractAgentID:
		return fmt.Sprintf("ContractAgentID(%v, %v)", agentID.ChainID().String(), agentID.Hname())
	case *old_isc.EthereumAddressAgentID:
		return fmt.Sprintf("EthereumAddressAgentID(%v, %v)", agentID.ChainID().String(), agentID.EthAddress().String())
	case *old_isc.NilAgentID:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDIsNil: %v", agentID))
	default:
		panic(fmt.Sprintf("Unknown agent ID kind: %v/%T = %v", agentID.Kind(), agentID, agentID))
	}
}

func newAgentIDToStr(agentID isc.AgentID) string {
	switch agentID := agentID.(type) {
	case *isc.AddressAgentID:
		return fmt.Sprintf("AddressAgentID(%v)", agentID.Address().String())
	case *isc.ContractAgentID:
		return fmt.Sprintf("ContractAgentID(%v, %v)", agentID.ChainID().String(), agentID.Hname())
	case *isc.EthereumAddressAgentID:
		return fmt.Sprintf("EthereumAddressAgentID(%v, %v)", agentID.ChainID().String(), agentID.EthAddress().String())
	case *isc.NilAgentID:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDIsNil: %v", agentID))
	default:
		panic(fmt.Sprintf("Unknown agent ID kind: %v/%T = %v", agentID.Kind(), agentID, agentID))
	}
}
