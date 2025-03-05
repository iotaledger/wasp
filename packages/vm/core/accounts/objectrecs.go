package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func (s *StateWriter) objectRecordsMap() *collections.Map {
	return collections.NewMap(s.state, KeyObjectRecords)
}

func (s *StateReader) objectRecordsMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, KeyObjectRecords)
}

func (s *StateWriter) SaveObject(rec *ObjectRecord) {
	s.objectRecordsMap().SetAt(rec.ID[:], rec.Bytes())
}

func (s *StateWriter) DeleteObject(id iotago.ObjectID) {
	s.objectRecordsMap().DelAt(id[:])
}

func (s *StateReader) GetObject(id iotago.ObjectID) *ObjectRecord {
	data := s.objectRecordsMapR().GetAt(id[:])
	if data == nil {
		return nil
	}
	return lo.Must(ObjectRecordFromBytes(data, id))
}

func (s *StateReader) GetObjectBCS(objectID iotago.ObjectID) []byte {
	o := s.GetObject(objectID)
	if o == nil {
		return nil
	}
	return o.BCS
}
