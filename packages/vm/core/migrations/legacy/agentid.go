// Package legacy defines legacy migrations
package legacy

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

/*
AgentIDToBytes returns the correct formatting of an AgentID based on the specified SchemaVersion
Stardust AgentID (For ED25519 addresses) was
  - AgentIDKind (Ethereum, ED25519, Contract) (byte)
  - AdressKind (User address, Alias, NFT, Native) (byte)
  - Address [32]Byte

Rebased AgentID is:
  - AgentIDKind (Ethereum, ED25519, Contract) (byte)
  - Address [32]Byte

Ethereum/Nil/Contract AgentIDs remain untouched.
*/
func AgentIDToBytes(v isc.SchemaVersion, id isc.AgentID) []byte {
	if v > allmigrations.SchemaVersionMigratedRebased {
		return id.Bytes()
	}

	if id.Kind() != isc.AgentIDKindAddress {
		return id.Bytes()
	}

	agentIDBytes := id.Bytes()

	resultBytes := make([]byte, 0)
	resultBytes = append(resultBytes, []byte{1, 0}...)
	resultBytes = append(resultBytes, agentIDBytes[1:]...)

	return resultBytes
}
