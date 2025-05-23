package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
)

func (s *StateReader) GetAssets(agentID isc.AgentID) *isc.Assets {
	return &isc.Assets{
		Coins:   s.getFungibleTokens(accountKey(agentID)),
		Objects: isc.NewObjectSet(s.getAccountObjects(agentID)...),
	}
}

func (s *StateReader) GetTotalAssets() *isc.Assets {
	return &isc.Assets{
		Coins:   s.getFungibleTokens(L2TotalsAccount),
		Objects: isc.NewObjectSet(s.getL2TotalObjects()...),
	}
}
