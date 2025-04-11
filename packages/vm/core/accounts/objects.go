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
	return PrefixObjects + string(agentID.Bytes())
}

func (s *StateReader) accountToObjectsMapR(agentID isc.AgentID) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, objectsMapKey(agentID))
}

func (s *StateWriter) accountToObjectsMap(agentID isc.AgentID) *collections.Map {
	return collections.NewMap(s.state, objectsMapKey(agentID))
}

func (s *StateWriter) objectToOwnerMap() *collections.Map {
	return collections.NewMap(s.state, KeyObjectOwner)
}

func (s *StateReader) objectToOwnerMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, KeyObjectOwner)
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

func (s *StateWriter) setObjectOwner(obj isc.IotaObject, agentID isc.AgentID) {
	// add to the mapping of ObjectID => owner
	objectMap := s.objectToOwnerMap()
	objectMap.SetAt(obj.ID[:], agentID.Bytes())

	// add to the mapping of agentID => {ObjectID => ObjectType}
	objects := s.accountToObjectsMap(agentID)
	objects.SetAt(obj.ID[:], codec.Encode(obj.Type))
}

// CreditObjectToAccount credits an Object to the on chain ledger
func (s *StateWriter) CreditObjectToAccount(agentID isc.AgentID, obj isc.IotaObject) {
	s.setObjectOwner(obj, agentID)
	s.touchAccount(agentID)
}

// DebitObjectFromAccount removes an Object from an account.
// If the account does not own the object, it panics.
func (s *StateWriter) DebitObjectFromAccount(agentID isc.AgentID, objectID iotago.ObjectID) iotago.ObjectType {
	t, ok := s.removeObjectOwner(objectID, agentID)
	if !ok {
		panic(fmt.Errorf("cannot debit Object %s from %s: %w", objectID.String(), agentID, ErrNotEnoughFunds))
	}
	s.touchAccount(agentID)
	return t
}

func collectObjects(m *collections.ImmutableMap) []isc.IotaObject {
	var ret []isc.IotaObject
	m.Iterate(func(idBytes []byte, tBytes []byte) bool {
		id := lo.Must(codec.Decode[iotago.ObjectID](idBytes))
		t := lo.Must(codec.Decode[iotago.ObjectType](tBytes))
		ret = append(ret, isc.NewIotaObject(id, t))
		return true
	})
	return ret
}

func (s *StateReader) getAccountObjects(agentID isc.AgentID) []isc.IotaObject {
	return collectObjects(s.accountToObjectsMapR(agentID))
}

func (s *StateReader) getL2TotalObjects() []isc.IotaObject {
	return collectObjects(s.objectToOwnerMapR())
}

// GetAccountObjects returns all Objects belonging to the agentID on the state
func (s *StateReader) GetAccountObjects(agentID isc.AgentID) []isc.IotaObject {
	return s.getAccountObjects(agentID)
}

func (s *StateReader) GetTotalL2Objects() []isc.IotaObject {
	return s.getL2TotalObjects()
}

func (s *StateReader) GetObject(id iotago.ObjectID) (isc.IotaObject, bool) {
	owner := s.objectToOwnerMapR().GetAt(id[:])
	if owner == nil {
		return isc.IotaObject{}, false
	}
	aid := lo.Must(codec.Decode[isc.AgentID](owner))
	t := lo.Must(iotago.ObjectTypeFromBytes(s.accountCoinBalancesMapR(AccountKey(aid)).GetAt(id[:])))
	return isc.NewIotaObject(id, t), true
}

func (s *StateReader) GetObjectsToOwnerMap() map[iotago.ObjectID]isc.AgentID {
	ret := make(map[iotago.ObjectID]isc.AgentID)
	s.objectToOwnerMapR().Iterate(func(k []byte, v []byte) bool {
		id := iotago.ObjectID{}
		copy(id[:], k)
		agentID, err := isc.AgentIDFromBytes(v)
		if err != nil {
			panic(fmt.Errorf("cannot convert %v to AgentID: %w", v, err))
		}
		ret[id] = agentID
		return true
	})
	return ret
}
