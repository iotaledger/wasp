package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func (s *StateWriter) objectRecordsMap() *collections.Map {
	return collections.NewMap(s.state, keyObjectRecords)
}

func (s *StateReader) objectRecordsMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyObjectRecords)
}

func (s *StateWriter) SaveObject(rec *ObjectRecord) {
	s.objectRecordsMap().SetAt(rec.ID[:], rec.Bytes())
}

func (s *StateWriter) DeleteObject(id sui.ObjectID) {
	s.objectRecordsMap().DelAt(id[:])
}

func (s *StateReader) GetObject(id sui.ObjectID) *ObjectRecord {
	data := s.objectRecordsMapR().GetAt(id[:])
	if data == nil {
		return nil
	}
	return lo.Must(ObjectRecordFromBytes(data, id))
}

func (s *StateReader) GetObjectBCS(objectID sui.ObjectID) []byte {
	o := s.GetObject(objectID)
	if o == nil {
		return nil
	}
	return o.BCS
}
