// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// region Registry /////////////////////////////////////////////////////////

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	log   *logger.Logger
	store kvstore.KVStore
}

// New creates new instance of the registry implementation.
func NewRegistry(log *logger.Logger, store kvstore.KVStore) *Impl {
	return &Impl{
		log:   log.Named("registry"),
		store: store,
	}
}

// endregion ////////////////////////////////////////////////////////

// region ChainRecordProvider /////////////////////////////////////////////////////////

func MakeChainRecordDbKey(chainID *iscp.ChainID) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, chainID.Bytes())
}

func (r *Impl) GetChainRecordByChainID(chainID *iscp.ChainID) (*ChainRecord, error) {
	data, err := r.store.Get(MakeChainRecordDbKey(chainID))
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ChainRecordFromBytes(data)
}

func (r *Impl) GetChainRecords() ([]*ChainRecord, error) {
	ret := make([]*ChainRecord, 0)

	err := r.store.Iterate([]byte{dbkeys.ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		if rec, err1 := ChainRecordFromBytes(value); err1 == nil {
			ret = append(ret, rec)
		}
		return true
	})
	return ret, err
}

func (r *Impl) UpdateChainRecord(chainID *iscp.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error) {
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

func (r *Impl) ActivateChainRecord(chainID *iscp.ChainID) (*ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func (r *Impl) DeactivateChainRecord(chainID *iscp.ChainID) (*ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func (r *Impl) SaveChainRecord(rec *ChainRecord) error {
	k := dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, rec.ChainID.Bytes())
	return r.store.Set(k, rec.Bytes())
}

// endregion ///////////////////////////////////////////////////////////////

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
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, ErrDKShareNotFound
		}
		return nil, err
	}
	return tcrypto.DKShareFromBytes(data, tcrypto.DefaultSuite())
}

func dbKeyForDKShare(sharedAddress ledgerstate.Address) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeDistributedKeyData, sharedAddress.Bytes())
}

// endregion //////////////////////////////////////////////////////////////

// region TrustedNetworkManager ////////////////////////////////////////////////////

// IsTrustedPeer implements TrustedNetworkManager interface.
func (r *Impl) IsTrustedPeer(pubKey ed25519.PublicKey) error {
	tp := &peering.TrustedPeer{PubKey: pubKey}
	tpKeyBytes, err := dbKeyForTrustedPeer(tp)
	if err != nil {
		return err
	}
	_, err = r.store.Get(tpKeyBytes)
	return err // Assume its trusted, if we can decode the entry.
}

// TrustPeer implements TrustedNetworkManager interface.
func (r *Impl) TrustPeer(pubKey ed25519.PublicKey, netID string) (*peering.TrustedPeer, error) {
	tp := &peering.TrustedPeer{PubKey: pubKey, NetID: netID}
	tpKeyBytes, err := dbKeyForTrustedPeer(tp)
	if err != nil {
		return nil, err
	}
	tpBinary, err := tp.Bytes()
	if err != nil {
		return nil, err
	}
	err = r.store.Set(tpKeyBytes, tpBinary)
	if err != nil {
		return nil, err
	}
	return tp, nil
}

// DistrustPeer implements TrustedNetworkManager interface.
// Get is kind of optional, so we ignore errors related to it.
func (r *Impl) DistrustPeer(pubKey ed25519.PublicKey) (*peering.TrustedPeer, error) {
	tp := &peering.TrustedPeer{PubKey: pubKey}
	tpKeyBytes, err := dbKeyForTrustedPeer(tp)
	if err != nil {
		return nil, err
	}
	tpBinary, getErr := r.store.Get(tpKeyBytes)
	delErr := r.store.Delete(tpKeyBytes)
	if delErr != nil {
		return nil, delErr
	}
	if getErr != nil {
		return nil, nil
	}
	tp, err = peering.TrustedPeerFromBytes(tpBinary)
	if err != nil {
		return nil, nil
	}
	return tp, nil
}

// TrustedPeers implements TrustedNetworkManager interface.
func (r *Impl) TrustedPeers() ([]*peering.TrustedPeer, error) {
	ret := make([]*peering.TrustedPeer, 0)
	err := r.store.Iterate([]byte{dbkeys.ObjectTypeTrustedPeer}, func(key kvstore.Key, value kvstore.Value) bool {
		if tp, recErr := peering.TrustedPeerFromBytes(value); recErr == nil {
			ret = append(ret, tp)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func dbKeyForTrustedPeer(tp *peering.TrustedPeer) ([]byte, error) {
	buf, err := tp.PubKeyBytes()
	if err != nil {
		return nil, err
	}
	return dbkeys.MakeKey(dbkeys.ObjectTypeTrustedPeer, buf), nil
}

// endregion  //////////////////////////////////////////////////////////////////////

// region BlobCacheProvider ///////////////////////////////////////////////

// TODO blob cache cleanup

func dbKeyForBlob(h hashing.HashValue) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeBlobCache, h[:])
}

func dbKeyForBlobTTL(h hashing.HashValue) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeBlobCacheTTL, h[:])
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
	if errors.Is(err, kvstore.ErrKeyNotFound) {
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
func (r *Impl) GetNodeIdentity() (*ed25519.KeyPair, error) {
	var err error
	var pair ed25519.KeyPair
	dbKey := dbKeyForNodeIdentity()
	var exists bool
	var data []byte
	exists, _ = r.store.Has(dbKey)
	if !exists {
		pair = ed25519.GenerateKeyPair()
		data = pair.PrivateKey.Bytes()
		if err := r.store.Set(dbKey, data); err != nil {
			return nil, err
		}
		r.log.Info("Node identity key pair generated.")
		return &pair, nil
	}
	if data, err = r.store.Get(dbKey); err != nil {
		return nil, err
	}
	if pair.PrivateKey, err, _ = ed25519.PrivateKeyFromBytes(data); err != nil {
		return nil, err
	}
	pair.PublicKey = pair.PrivateKey.Public()
	return &pair, nil
}

// GetNodePublicKey implements NodeIdentityProvider.
func (r *Impl) GetNodePublicKey() (*ed25519.PublicKey, error) {
	var err error
	var pair *ed25519.KeyPair
	if pair, err = r.GetNodeIdentity(); err != nil {
		return nil, err
	}
	return &pair.PublicKey, nil
}

func dbKeyForNodeIdentity() []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeNodeIdentity)
}

// endregion ///////////////////////////////////////////////////
