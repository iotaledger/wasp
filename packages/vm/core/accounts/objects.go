package accounts

import (
	"bytes"
	"fmt"

	"github.com/samber/lo"

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

func (s *StateWriter) removeObjectOwner(objectID iotago.ObjectID, agentID isc.AgentID) (iotago.ObjectType, bool) {
	// remove the mapping of ObjectID => owner
	objectMap := s.objectToOwnerMap()
	if bytes.Compare(objectMap.GetAt(objectID[:]), agentID.Bytes()) != 0 {
		return iotago.ObjectType{}, false
	}

	// remove the mapping of agentID => {ObjectID => ObjectType}
	objects := s.accountToObjectsMap(agentID)
	tBin := objects.GetAt(objectID[:])
	if tBin == nil {
		return iotago.ObjectType{}, false
	}

	t := lo.Must(iotago.ObjectTypeFromBytes(tBin))
	objectMap.DelAt(objectID[:])
	objects.DelAt(objectID[:])
	return t, true
}

func (s *StateWriter) setObjectOwner(objectID iotago.ObjectID, t iotago.ObjectType, agentID isc.AgentID) {
	// add to the mapping of ObjectID => owner
	objectMap := s.objectToOwnerMap()
	objectMap.SetAt(objectID[:], agentID.Bytes())

	// add to the mapping of agentID => {ObjectID => ObjectType}
	objects := s.accountToObjectsMap(agentID)
	objects.SetAt(objectID[:], codec.Encode(t))
}

// CreditObjectToAccount credits an Object to the on chain ledger
func (s *StateWriter) CreditObjectToAccount(agentID isc.AgentID, objectID iotago.ObjectID, t iotago.ObjectType, chainID isc.ChainID) {
	s.setObjectOwner(objectID, t, agentID)
	s.touchAccount(agentID, chainID)
}

// DebitObjectFromAccount removes an Object from an account.
// If the account does not own the object, it panics.
func (s *StateWriter) DebitObjectFromAccount(agentID isc.AgentID, objectID iotago.ObjectID, chainID isc.ChainID) iotago.ObjectType {
	t, ok := s.removeObjectOwner(objectID, agentID)
	if !ok {
		panic(fmt.Errorf("cannot debit Object %s from %s: %w", objectID.String(), agentID, ErrNotEnoughFunds))
	}
	s.touchAccount(agentID, chainID)
	return t
}

func collectObjects(m *collections.ImmutableMap) []isc.L1Object {
	var ret []isc.L1Object
	m.Iterate(func(idBytes []byte, tBytes []byte) bool {
		id := lo.Must(codec.Decode[iotago.ObjectID](idBytes))
		t := lo.Must(codec.Decode[iotago.ObjectType](tBytes))
		ret = append(ret, lo.T2(id, t))
		return true
	})
	return ret
}

func (s *StateReader) getAccountObjects(agentID isc.AgentID) []isc.L1Object {
	return collectObjects(s.accountToObjectsMapR(agentID))
}

func (s *StateReader) getL2TotalObjects() []isc.L1Object {
	return collectObjects(s.objectToOwnerMapR())
}

// GetAccountObjects returns all Objects belonging to the agentID on the state
func (s *StateReader) GetAccountObjects(agentID isc.AgentID) []isc.L1Object {
	return s.getAccountObjects(agentID)
}

func (s *StateReader) GetTotalL2Objects() []isc.L1Object {
	return s.getL2TotalObjects()
}

func (s *StateReader) GetObject(id iotago.ObjectID, chID isc.ChainID) (isc.L1Object, bool) {
	owner := s.objectToOwnerMapR().GetAt(id[:])
	if owner == nil {
		return isc.L1Object{}, false
	}
	aid := lo.Must(codec.Decode[isc.AgentID](owner))
	t := lo.Must(iotago.ObjectTypeFromBytes(s.accountCoinBalancesMapR(accountKey(aid, chID)).GetAt(id[:])))
	return lo.T2(id, t), true
}
