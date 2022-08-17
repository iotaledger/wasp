// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/cryptolib"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// region Registry /////////////////////////////////////////////////////////

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	log          *logger.Logger
	store        kvstore.KVStore
	nodeIdentity *cryptolib.KeyPair
}

var (
	_ NodeIdentityProvider        = &Impl{}
	_ DKShareRegistryProvider     = &Impl{}
	_ ChainRecordRegistryProvider = &Impl{}
	_ journal.Registry            = &Impl{}
)

type Config struct {
	UseText  bool
	Filename string
}

func DefaultConfig() *Config {
	return &Config{
		UseText:  false,
		Filename: "",
	}
}

// New creates new instance of the registry implementation.
func NewRegistry(log *logger.Logger, store kvstore.KVStore) *Impl {
	result := &Impl{
		log:   log.Named("registry"),
		store: store,
	}
	// Read or create node identity - private/public key pair
	dbKey := dbKeyForNodeIdentity()
	exists, _ := result.store.Has(dbKey)
	if !exists {
		result.nodeIdentity = cryptolib.NewKeyPair()
		data := result.nodeIdentity.GetPrivateKey().AsBytes()
		result.log.Infof("Node identity key pair generated. PublicKey: %s", result.nodeIdentity.GetPublicKey())
		if err := result.store.Set(dbKey, data); err != nil {
			result.log.Error("Generated node identity cannot be stored: %v", err)
		}
	}
	data, err := result.store.Get(dbKey)
	if err != nil {
		result.log.Panicf("Cannot read node identity: %v", err)
	}
	privateKey, err := cryptolib.NewPrivateKeyFromBytes(data)
	if err != nil {
		result.log.Panicf("Cannot read create private key from read node identity: %v", err)
	}
	result.nodeIdentity = cryptolib.NewKeyPairFromPrivateKey(privateKey)

	return result
}

// endregion ////////////////////////////////////////////////////////

// region ChainRecordProvider /////////////////////////////////////////////////////////

func MakeChainRecordDbKey(chainID *isc.ChainID) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, chainID.Bytes())
}

func (r *Impl) GetChainRecordByChainID(chainID *isc.ChainID) (*ChainRecord, error) {
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

func (r *Impl) UpdateChainRecord(chainID *isc.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error) {
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

func (r *Impl) ActivateChainRecord(chainID *isc.ChainID) (*ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func (r *Impl) DeactivateChainRecord(chainID *isc.ChainID) (*ChainRecord, error) {
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
func (r *Impl) SaveDKShare(dkShare tcrypto.DKShare) error {
	var err error
	var exists bool
	dbKey := dbKeyForDKShare(dkShare.GetAddress())

	r.log.Infof("Saving DKShare for address=%v as key %v", dkShare.GetAddress().String(), dbKey)

	if exists, err = r.store.Has(dbKey); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}
	return r.store.Set(dbKey, dkShare.Bytes())
}

// LoadDKShare implements dkg.DKShareRegistryProvider.
func (r *Impl) LoadDKShare(sharedAddress iotago.Address) (tcrypto.DKShare, error) {
	data, err := r.store.Get(dbKeyForDKShare(sharedAddress))
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, ErrDKShareNotFound
		}
		return nil, err
	}
	return tcrypto.DKShareFromBytes(data, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite(), r.nodeIdentity.GetPrivateKey())
}

func dbKeyForDKShare(sharedAddress iotago.Address) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeDistributedKeyData, isc.BytesFromAddress(sharedAddress))
}

// endregion //////////////////////////////////////////////////////////////

// region TrustedNetworkManager ////////////////////////////////////////////////////

// IsTrustedPeer implements TrustedNetworkManager interface.
func (r *Impl) IsTrustedPeer(pubKey *cryptolib.PublicKey) error {
	tp := &peering.TrustedPeer{PubKey: pubKey}
	tpKeyBytes, err := dbKeyForTrustedPeer(tp)
	if err != nil {
		return err
	}
	_, err = r.store.Get(tpKeyBytes)
	return err // Assume its trusted, if we can decode the entry.
}

// TrustPeer implements TrustedNetworkManager interface.
func (r *Impl) TrustPeer(pubKey *cryptolib.PublicKey, netID string) (*peering.TrustedPeer, error) {
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
func (r *Impl) DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error) {
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
	currentTime := time.Now()
	cleanAfter := currentTime.Add(BlobCacheDefaultTTL).UnixNano()
	if len(ttl) > 0 {
		cleanAfter = currentTime.Add(ttl[0]).UnixNano()
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
func (r *Impl) GetNodeIdentity() *cryptolib.KeyPair {
	return r.nodeIdentity
}

// GetNodePublicKey implements NodeIdentityProvider.
func (r *Impl) GetNodePublicKey() *cryptolib.PublicKey {
	return r.GetNodeIdentity().GetPublicKey()
}

func dbKeyForNodeIdentity() []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeNodeIdentity)
}

// endregion ///////////////////////////////////////////////////
