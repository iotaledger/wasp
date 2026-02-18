package l1

import (
	"crypto/sha3"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

// StoredTx holds a committed transaction and its effects.
type StoredTx struct {
	TxData     []byte
	Effects    *graphqltypes.ExecuteTransactionBlockResponse
	Sender     iotago.Address
	Digest     iotago.TransactionDigest
	Signatures []iotago.Base64Data
}

// ObjectStore is an in-memory object database with indices for owner, type,
// dynamic fields, and version history.
type ObjectStore struct {
	mu           sync.RWMutex
	objects      map[iotago.ObjectID]*SimObject
	history      map[iotago.ObjectID]map[uint64]*SimObject // objectID → version → snapshot
	transactions map[iotago.TransactionDigest]*StoredTx
	ownerIndex   map[iotago.Address]map[iotago.ObjectID]struct{}
	dynFields    map[iotago.ObjectID][]DynamicField // parent → children
	deleted      map[iotago.ObjectID]struct{}
}

func NewObjectStore() *ObjectStore {
	return &ObjectStore{
		objects:      make(map[iotago.ObjectID]*SimObject),
		history:      make(map[iotago.ObjectID]map[uint64]*SimObject),
		transactions: make(map[iotago.TransactionDigest]*StoredTx),
		ownerIndex:   make(map[iotago.Address]map[iotago.ObjectID]struct{}),
		dynFields:    make(map[iotago.ObjectID][]DynamicField),
		deleted:      make(map[iotago.ObjectID]struct{}),
	}
}

func (s *ObjectStore) Get(id iotago.ObjectID) (*SimObject, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obj, ok := s.objects[id]
	if !ok {
		return nil, false
	}
	return obj.Clone(), true
}

func (s *ObjectStore) Exists(id iotago.ObjectID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.objects[id]
	return ok
}

func (s *ObjectStore) IsDeleted(id iotago.ObjectID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.deleted[id]
	return ok
}

// Put stores an object, snapshotting the previous version into history.
func (s *ObjectStore) Put(obj *SimObject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.putLocked(obj)
}

func (s *ObjectStore) putLocked(obj *SimObject) {
	id := obj.ID

	if prev, ok := s.objects[id]; ok {
		if _, exists := s.history[id]; !exists {
			s.history[id] = make(map[uint64]*SimObject)
		}
		s.history[id][prev.Version] = prev.Clone()
		s.removeFromOwnerIndexLocked(prev)
	}

	s.objects[id] = obj.Clone()
	delete(s.deleted, id)
	s.addToOwnerIndexLocked(obj)
}

func (s *ObjectStore) Delete(id iotago.ObjectID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteLocked(id)
}

func (s *ObjectStore) deleteLocked(id iotago.ObjectID) {
	if obj, ok := s.objects[id]; ok {
		if _, exists := s.history[id]; !exists {
			s.history[id] = make(map[uint64]*SimObject)
		}
		s.history[id][obj.Version] = obj.Clone()
		s.removeFromOwnerIndexLocked(obj)
		delete(s.objects, id)
	}
	s.deleted[id] = struct{}{}
}

func (s *ObjectStore) addToOwnerIndexLocked(obj *SimObject) {
	var ownerAddr *iotago.Address
	switch {
	case obj.Owner.AddressOwner != nil:
		ownerAddr = obj.Owner.AddressOwner
	case obj.Owner.ObjectOwner != nil:
		ownerAddr = obj.Owner.ObjectOwner
	}
	if ownerAddr != nil {
		if _, ok := s.ownerIndex[*ownerAddr]; !ok {
			s.ownerIndex[*ownerAddr] = make(map[iotago.ObjectID]struct{})
		}
		s.ownerIndex[*ownerAddr][obj.ID] = struct{}{}
	}
}

func (s *ObjectStore) removeFromOwnerIndexLocked(obj *SimObject) {
	var ownerAddr *iotago.Address
	switch {
	case obj.Owner.AddressOwner != nil:
		ownerAddr = obj.Owner.AddressOwner
	case obj.Owner.ObjectOwner != nil:
		ownerAddr = obj.Owner.ObjectOwner
	}
	if ownerAddr != nil {
		if set, ok := s.ownerIndex[*ownerAddr]; ok {
			delete(set, obj.ID)
			if len(set) == 0 {
				delete(s.ownerIndex, *ownerAddr)
			}
		}
	}
}

func (s *ObjectStore) GetAtVersion(id iotago.ObjectID, version uint64) (*SimObject, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if obj, ok := s.objects[id]; ok && obj.Version == version {
		return obj.Clone(), true
	}
	if versions, ok := s.history[id]; ok {
		if obj, ok := versions[version]; ok {
			return obj.Clone(), true
		}
	}
	return nil, false
}

func (s *ObjectStore) StoreTx(digest iotago.TransactionDigest, tx *StoredTx) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactions[digest] = tx
}

func (s *ObjectStore) GetTx(digest iotago.TransactionDigest) (*StoredTx, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tx, ok := s.transactions[digest]
	return tx, ok
}

func (s *ObjectStore) GetByOwner(addr iotago.Address) []*SimObject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set, ok := s.ownerIndex[addr]
	if !ok {
		return nil
	}
	result := make([]*SimObject, 0, len(set))
	for id := range set {
		if obj, ok := s.objects[id]; ok {
			result = append(result, obj.Clone())
		}
	}
	return result
}

func (s *ObjectStore) GetByOwnerAndType(addr iotago.Address, typeFilter string) []*SimObject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set, ok := s.ownerIndex[addr]
	if !ok {
		return nil
	}
	result := make([]*SimObject, 0)
	for id := range set {
		if obj, ok := s.objects[id]; ok {
			if matchesType(obj.Type, typeFilter) {
				result = append(result, obj.Clone())
			}
		}
	}
	return result
}

func (s *ObjectStore) GetCoinsByOwner(addr iotago.Address, coinType string) []*SimObject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set, ok := s.ownerIndex[addr]
	if !ok {
		return nil
	}
	wantType := CoinTypeString(coinType)
	result := make([]*SimObject, 0)
	for id := range set {
		if obj, ok := s.objects[id]; ok {
			if matchesType(obj.Type, wantType) {
				result = append(result, obj.Clone())
			}
		}
	}
	return result
}

func (s *ObjectStore) SumCoinBalance(addr iotago.Address, coinType string) uint64 {
	coins := s.GetCoinsByOwner(addr, coinType)
	var total uint64
	for _, c := range coins {
		total += DecodeCoinObjectBalance(c.Data)
	}
	return total
}

func (s *ObjectStore) GetAllCoinBalances(addr iotago.Address) map[string]uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set, ok := s.ownerIndex[addr]
	if !ok {
		return nil
	}
	result := make(map[string]uint64)
	for id := range set {
		if obj, ok := s.objects[id]; ok {
			if ct, ok := extractCoinType(obj.Type); ok {
				result[ct] += DecodeCoinObjectBalance(obj.Data)
			}
		}
	}
	return result
}

func (s *ObjectStore) AddDynamicField(df DynamicField) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dynFields[df.ParentID] = append(s.dynFields[df.ParentID], df)
}

func (s *ObjectStore) GetDynamicFields(parentID iotago.ObjectID) []DynamicField {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fields := s.dynFields[parentID]
	result := make([]DynamicField, len(fields))
	copy(result, fields)
	return result
}

func (s *ObjectStore) GetDynamicField(parentID iotago.ObjectID, nameTypeRepr string, nameJSON string) (*DynamicField, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.dynFields[parentID] {
		if f.Name.TypeRepr == nameTypeRepr && string(f.Name.Json) == nameJSON {
			return &f, true
		}
	}
	return nil, false
}

func (s *ObjectStore) RemoveDynamicField(parentID iotago.ObjectID, nameTypeRepr string, nameJSON string) (*DynamicField, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fields := s.dynFields[parentID]
	for i, f := range fields {
		if f.Name.TypeRepr == nameTypeRepr && string(f.Name.Json) == nameJSON {
			removed := f
			s.dynFields[parentID] = append(fields[:i], fields[i+1:]...)
			return &removed, true
		}
	}
	return nil, false
}

func (s *ObjectStore) UpdateDynamicFieldValue(parentID iotago.ObjectID, nameTypeRepr string, nameJSON string, newValueObjID iotago.ObjectID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, f := range s.dynFields[parentID] {
		if f.Name.TypeRepr == nameTypeRepr && string(f.Name.Json) == nameJSON {
			s.dynFields[parentID][i].ValueObjID = newValueObjID
			return true
		}
	}
	return false
}

// FreshID generates a new ObjectID from the transaction digest and a counter,
// matching the Rust derive_id() function.
func FreshID(txDigest iotago.TransactionDigest, counter *uint64) iotago.ObjectID {
	h := sha3.New256()
	h.Write(txDigest[:])
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, *counter)
	h.Write(buf)
	*counter++
	var id iotago.ObjectID
	copy(id[:], h.Sum(nil))
	return id
}

func ComputeDigest(data []byte) iotago.Digest {
	h := hashing.HashDataBlake2b(data)
	return iotago.Digest(h)
}

// NextLamportVersion computes max(versions) + 1.
func NextLamportVersion(versions ...uint64) uint64 {
	var max uint64
	for _, v := range versions {
		if v > max {
			max = v
		}
	}
	return max + 1
}

// PutBatch stores multiple objects atomically.
func (s *ObjectStore) PutBatch(objs []*SimObject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, obj := range objs {
		s.putLocked(obj)
	}
}

// DeleteBatch deletes multiple objects atomically.
func (s *ObjectStore) DeleteBatch(ids []iotago.ObjectID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		s.deleteLocked(id)
	}
}

// PresetCoinObject creates a coin object owned by addr and stores it.
func (s *ObjectStore) PresetCoinObject(id iotago.ObjectID, owner iotago.Address, coinType string, balance uint64, txDigest iotago.TransactionDigest) {
	data := encodeCoinObject(id, balance)
	fullType := CoinTypeString(coinType)
	obj := &SimObject{
		ID:         id,
		Version:    1,
		Digest:     ComputeDigest(data),
		Owner:      SimOwner{AddressOwner: &owner},
		Type:       fullType,
		Data:       data,
		PreviousTx: txDigest,
	}
	s.Put(obj)
}

// matchesType checks if the object type matches the filter.
// Supports exact match and prefix match (e.g. "pkg::mod::Type" matches "pkg::mod::Type<...>").
func matchesType(objType, filter string) bool {
	if objType == filter {
		return true
	}
	normObj := normalizeTypeString(objType)
	normFilter := normalizeTypeString(filter)
	if normObj == normFilter {
		return true
	}
	if strings.HasPrefix(normObj, normFilter) {
		return true
	}
	return false
}

func normalizeTypeString(s string) string {
	return s
}

func extractCoinType(objType string) (string, bool) {
	const marker = "::coin::Coin<"
	idx := strings.Index(objType, marker)
	if idx < 0 {
		return "", false
	}
	inner := objType[idx+len(marker):]
	if len(inner) > 0 && inner[len(inner)-1] == '>' {
		inner = inner[:len(inner)-1]
	}
	return inner, true
}

// encodeCoinObject BCS-marshals a full Coin object using iscmoveclient.MoveCoin.
func encodeCoinObject(id iotago.ObjectID, balance uint64) []byte {
	data, err := bcs.Marshal(&iscmoveclient.MoveCoin{ID: id, Balance: balance})
	if err != nil {
		panic(fmt.Sprintf("encodeCoinObject: BCS marshal failed: %v", err))
	}
	return data
}

// DecodeCoinObjectBalance extracts the balance from a BCS-encoded Coin object (MoveCoin: ID + Balance).
func DecodeCoinObjectBalance(data []byte) uint64 {
	coin, err := bcs.Unmarshal[iscmoveclient.MoveCoin](data)
	if err != nil {
		return 0
	}
	return coin.Balance
}

// DecodeBalanceValue extracts the balance from a BCS-encoded Balance<T> value (u64).
func DecodeBalanceValue(data []byte) uint64 {
	bal, err := bcs.Unmarshal[uint64](data)
	if err != nil {
		return 0
	}
	return bal
}
