// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/key"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"

	"github.com/iotaledger/wasp/packages/registry/chainrecord"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/database"
)

// region Registry /////////////////////////////////////////////////////////

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	suite tcrypto.Suite
	log   *logger.Logger
	store kvstore.KVStore
}

// New creates new instance of the registry implementation.
func NewRegistry(suite tcrypto.Suite, log *logger.Logger, store kvstore.KVStore) *Impl {
	if store == nil {
		store = database.GetRegistryKVStore()
	}
	ret := &Impl{
		suite: suite,
		log:   log.Named("registry"),
		store: store,
	}

	return ret
}

// endregion ////////////////////////////////////////////////////////

// region ChainRecordProvider /////////////////////////////////////////////////////////

func MakeChainRecordDbKey(chainID *chainid.ChainID) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, chainID.Bytes())
}

func (r *Impl) GetChainRecordByChainID(chainID *chainid.ChainID) (*chainrecord.ChainRecord, error) {
	data, err := r.store.Get(MakeChainRecordDbKey(chainID))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return chainrecord.ChainRecordFromBytes(data)
}

func (r *Impl) GetChainRecords() ([]*chainrecord.ChainRecord, error) {
	ret := make([]*chainrecord.ChainRecord, 0)

	err := r.store.Iterate([]byte{dbkeys.ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		if rec, err1 := chainrecord.ChainRecordFromBytes(value); err1 == nil {
			ret = append(ret, rec)
		}
		return true
	})
	return ret, err
}

func (r *Impl) UpdateChainRecord(chainID *chainid.ChainID, f func(*chainrecord.ChainRecord) bool) (*chainrecord.ChainRecord, error) {
	rec, err := r.GetChainRecordByChainID(chainID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("no chain record found for chainID %s", chainID.String())
	}
	if f(rec) {
		err = r.SaveChainRecord(rec)
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func (r *Impl) ActivateChainRecord(chainID *chainid.ChainID) (*chainrecord.ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *chainrecord.ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func (r *Impl) DeactivateChainRecord(chainID *chainid.ChainID) (*chainrecord.ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *chainrecord.ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func (r *Impl) SaveChainRecord(rec *chainrecord.ChainRecord) error {
	key := dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, rec.ChainID.Bytes())
	return r.store.Set(key, rec.Bytes())
}

// endregion ///////////////////////////////////////////////////////////////

// region CommitteeRegistryProvider  ///////////////////////////////////////////////////////////

func dbKeyCommitteeRecord(addr ledgerstate.Address) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeCommitteeRecord, addr.Bytes())
}

func (r *Impl) GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	data, err := r.store.Get(dbKeyCommitteeRecord(addr))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return committee_record.CommitteeRecordFromBytes(data)
}

func (r *Impl) SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error {
	return r.store.Set(dbKeyCommitteeRecord(rec.Address), rec.Bytes())
}

// endregion  //////////////////////////////////////////////////////////////////////

// region DKShareRegistryProvider ////////////////////////////////////////////////////

// SaveDKShare implements dkg.DKShareRegistryProvider.
func (r *Impl) SaveDKShare(dkShare *tcrypto.DKShare) error {
	var err error
	var exists bool
	dbKey := dbKeyForDKShare(dkShare.Address)

	if exists, err = r.store.Has(dbKey); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}
	return r.store.Set(dbKey, dkShare.Bytes())
}

// LoadDKShare implements dkg.DKShareRegistryProvider.
func (r *Impl) LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error) {
	data, err := r.store.Get(dbKeyForDKShare(sharedAddress))
	if err != nil {
		return nil, err
	}
	return tcrypto.DKShareFromBytes(data, r.suite)
}

func dbKeyForDKShare(sharedAddress ledgerstate.Address) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeDistributedKeyData, sharedAddress.Bytes())
}

// endregion //////////////////////////////////////////////////////////////

// region BlobCacheProvider ///////////////////////////////////////////////

// TODO blob cache cleanup

func dbKeyForBlob(h hashing.HashValue) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeBlobCache, h[:])
}

func dbKeyForBlobTTL(h hashing.HashValue) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeBlobCacheTTL)
}

const BlobCacheDefaultTTL = 1 * time.Hour

// PutBlob Writes data into the registry with the key of its hash
// Also stores TTL if provided
func (r *Impl) PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error) {
	h := hashing.HashData(data)
	err := r.store.Set(dbKeyForBlob(h), data)
	if err != nil {
		return hashing.NilHash, err
	}
	nowis := time.Now()
	cleanAfter := nowis.Add(BlobCacheDefaultTTL).UnixNano()
	if len(ttl) > 0 {
		cleanAfter = nowis.Add(ttl[0]).UnixNano()
	}
	if cleanAfter > 0 {
		err = r.store.Set(dbKeyForBlobTTL(h), codec.EncodeInt64(cleanAfter))
		if err != nil {
			return hashing.NilHash, err
		}
	}
	r.log.Infof("data blob has been stored. size: %d bytes, hash: %s", len(data), h)
	return h, nil
}

// Reads data from registry by hash. Returns existence flag
func (r *Impl) GetBlob(h hashing.HashValue) ([]byte, bool, error) {
	ret, err := r.store.Get(dbKeyForBlob(h))
	if err == kvstore.ErrKeyNotFound {
		return nil, false, nil
	}
	return ret, ret != nil && err == nil, err
}

func (r *Impl) HasBlob(h hashing.HashValue) (bool, error) {
	return r.store.Has(dbKeyForBlob(h))
}

// endregion /////////////////////////////////////////////////////////////

// region NodeIdentity //////////////////////////////////////////

// GetNodeIdentity implements NodeIdentityProvider.
func (r *Impl) GetNodeIdentity() (*key.Pair, error) {
	var err error
	var pair *key.Pair
	dbKey := dbKeyForNodeIdentity()
	var exists bool
	var data []byte
	exists, err = r.store.Has(dbKey)
	if !exists {
		pair = key.NewKeyPair(r.suite)
		if data, err = keyPairToBytes(pair); err != nil {
			return nil, err
		}
		r.store.Set(dbKey, data)
		r.log.Info("Node identity key pair generated.")
		return pair, nil
	}
	if data, err = r.store.Get(dbKey); err != nil {
		return nil, err
	}
	if pair, err = keyPairFromBytes(data, r.suite); err != nil {
		return nil, err
	}
	return pair, nil
}

// GetNodePublicKey implements NodeIdentityProvider.
func (r *Impl) GetNodePublicKey() (kyber.Point, error) {
	var err error
	var pair *key.Pair
	if pair, err = r.GetNodeIdentity(); err != nil {
		return nil, err
	}
	return pair.Public, nil
}

func dbKeyForNodeIdentity() []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeNodeIdentity)
}

func keyPairToBytes(pair *key.Pair) ([]byte, error) {
	var err error
	var w bytes.Buffer
	if err = util.WriteMarshaled(&w, pair.Private); err != nil {
		return nil, err
	}
	if err = util.WriteMarshaled(&w, pair.Public); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func keyPairFromBytes(buf []byte, suite kyber.Group) (*key.Pair, error) {
	var err error
	r := bytes.NewReader(buf)
	pair := key.Pair{
		Public:  suite.Point(),
		Private: suite.Scalar(),
	}
	if err = util.ReadMarshaled(r, pair.Private); err != nil {
		return nil, err
	}
	if err = util.ReadMarshaled(r, pair.Public); err != nil {
		return nil, err
	}
	return &pair, nil
}

// endregion ///////////////////////////////////////////////////
