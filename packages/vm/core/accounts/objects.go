package accounts

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func objectsMapKey(agentID isc.AgentID) string {
	return prefixObjects + string(agentID.Bytes())
}

func (s *StateReader) accountToObjectsMapR(agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, objectsMapKey(agentID))
}

func (s *StateWriter) accountToObjectsMap(agentID isc.AgentID) *collections.Map {
	return collections.NewMap(s.state, objectsMapKey(agentID))
}

func (s *StateWriter) objectToOwnerMap() *collections.Map {
	return collections.NewMap(s.state, keyObjectOwner)
}

func (s *StateReader) objectToOwnerMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyObjectOwner)
}

func (s *StateReader) hasObject(agentID isc.AgentID, objectID iotago.ObjectID) bool {
	return s.accountToObjectsMapR(agentID).HasAt(objectID[:])
}

func (s *StateWriter) removeObjectOwner(objectID iotago.ObjectID, agentID isc.AgentID) bool {
	// remove the mapping of ObjectID => owner
	objectMap := s.objectToOwnerMap()
	if !objectMap.HasAt(objectID[:]) {
		return false
	}
	objectMap.DelAt(objectID[:])

	// add to the mapping of agentID => []ObjectIDs
	objects := s.accountToObjectsMap(agentID)
	if !objects.HasAt(objectID[:]) {
		return false
	}
	objects.DelAt(objectID[:])
	return true
}

func (s *StateWriter) setObjectOwner(objectID iotago.ObjectID, agentID isc.AgentID) {
	// add to the mapping of ObjectID => owner
	objectMap := s.objectToOwnerMap()
	objectMap.SetAt(objectID[:], agentID.Bytes())

	// add to the mapping of agentID => []ObjectIDs
	objects := s.accountToObjectsMap(agentID)
	objects.SetAt(objectID[:], codec.Encode(true))
}

// CreditObjectToAccount credits an Object to the on chain ledger
func (s *StateWriter) CreditObjectToAccount(agentID isc.AgentID, objectID iotago.ObjectID) {
	s.setObjectOwner(objectID, agentID)
	s.touchAccount(agentID, chainID)
}

// DebitObjectFromAccount removes an Object from an account.
// If the account does not own the object, it panics.
func (s *StateWriter) DebitObjectFromAccount(agentID isc.AgentID, objectID iotago.ObjectID) {
	if !s.removeObjectOwner(objectID, agentID) {
		panic(fmt.Errorf("cannot debit Object %s from %s: %w", objectID.String(), agentID, ErrNotEnoughFunds))
	}
	s.touchAccount(agentID)
}

func collectObjectIDs(m *collections.ImmutableMap) []iotago.ObjectID {
	var ret []iotago.ObjectID
	m.Iterate(func(idBytes []byte, val []byte) bool {
		id := iotago.ObjectID{}
		copy(id[:], idBytes)
		ret = append(ret, id)
		return true
	})
	return ret
}

func (s *StateReader) getAccountObjects(agentID isc.AgentID) []iotago.ObjectID {
	return collectObjectIDs(s.accountToObjectsMapR(agentID))
}

func (s *StateReader) getL2TotalObjects() []iotago.ObjectID {
	return collectObjectIDs(s.objectToOwnerMapR())
}

// GetAccountObjects returns all Objects belonging to the agentID on the state
func (s *StateReader) GetAccountObjects(agentID isc.AgentID) []iotago.ObjectID {
	return s.getAccountObjects(agentID)
}

func (s *StateReader) GetTotalL2Objects() []iotago.ObjectID {
	return s.getL2TotalObjects()
}
