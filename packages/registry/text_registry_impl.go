package registry

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/database/textdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type textImpl struct {
	log          *logger.Logger
	store        kvstore.KVStore
	nodeIdentity *cryptolib.KeyPair
}

type AuxChainRecord struct {
	Active  bool
	ChainID string
}

var _ Registry = &textImpl{}

func NewTextRegistry(log *logger.Logger, filename string) Registry {
	result := &textImpl{
		log:   log.Named("registry"),
		store: textdb.NewTextKV(log.Named("textkv"), filename),
	}
	// Read or create node identity - private/public key pair
	dbKey, err := hexEncode(dbKeyForNodeIdentity())
	if err != nil {
		result.log.Panicf("Cannot create key for node identity")
	}
	exists, _ := result.store.Has(dbKey)
	if !exists {
		result.nodeIdentity = cryptolib.NewKeyPair()
		data, err := hexEncode(result.nodeIdentity.GetPrivateKey().AsBytes())
		if err != nil {
			result.log.Panicf("Generated node identity cannot be stored: %v")
		}
		result.log.Infof("Node identity key pair generated. PublicKey: %s", result.nodeIdentity.GetPublicKey())
		if err := result.store.Set(dbKey, data); err != nil {
			result.log.Error("Generated node identity cannot be stored: %v", err)
		}
	}
	data, err := result.store.Get(dbKey)
	if err != nil {
		result.log.Panicf("Cannot read node identity: %v", err)
	}
	data, err = hexDecode(data) // data will hold bytes for a json string. Needs to be decoded
	if err != nil {
		result.log.Panicf("Cannot create private key from node identity: %v", err)
	}
	privateKey, err := cryptolib.NewPrivateKeyFromBytes(data)
	if err != nil {
		result.log.Panicf("Canot create private key from read node identity: %v", err)
	}
	result.nodeIdentity = cryptolib.NewKeyPairFromPrivateKey(privateKey)
	return result
}

func hexEncode(data []byte) ([]byte, error) {
	return json.Marshal(hexutil.Encode(data))
}

func hexDecode(data []byte) ([]byte, error) {
	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return hexutil.Decode(v)
}

// NodeIdentityProvider implementation
func (r *textImpl) GetNodeIdentity() *cryptolib.KeyPair {
	return r.nodeIdentity
}

func (r *textImpl) GetNodePublicKey() *cryptolib.PublicKey {
	return r.GetNodeIdentity().GetPublicKey()
}

// ChainRecordRegistryProvider implementation
func (r *textImpl) GetChainRecordByChainID(chainID *isc.ChainID) (*ChainRecord, error) {
	key, err := json.Marshal(chainID.String())
	if err != nil {
		return nil, fmt.Errorf("Error encoding key for ChainRecord: %w", err)
	}
	data, err := r.store.Get(key)
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var auxRec AuxChainRecord
	err = json.Unmarshal(data, &auxRec)
	if err != nil {
		return nil, fmt.Errorf("Error decoding ChainRecord from registry: %w", err)
	}
	id, err := isc.ChainIDFromString(auxRec.ChainID)
	if err != nil {
		return nil, fmt.Errorf("Error reading ChainID from registry: %w", err)
	}
	return &ChainRecord{Active: auxRec.Active, ChainID: *id}, nil
}

func (r *textImpl) GetChainRecords() ([]*ChainRecord, error) {
	ret := make([]*ChainRecord, 0)

	// An empty prefix will match all records. Less efficient but keys are hex encoded so providing a prefix risks skipping data
	err := r.store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		var auxRec AuxChainRecord
		if err1 := json.Unmarshal(value, &auxRec); err1 == nil {
			chainID, err := isc.ChainIDFromString(auxRec.ChainID)
			if err != nil {
				return false
			}
			ret = append(ret, &ChainRecord{Active: auxRec.Active, ChainID: *chainID})
		}
		return true
	})
	return ret, err
}

func (r *textImpl) UpdateChainRecord(chainID *isc.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error) {
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

func (r *textImpl) ActivateChainRecord(chainID *isc.ChainID) (*ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func (r *textImpl) DeactivateChainRecord(chainID *isc.ChainID) (*ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func (r *textImpl) SaveChainRecord(rec *ChainRecord) error {
	key, err := json.Marshal(rec.ChainID.String())
	if err != nil {
		return fmt.Errorf("Error encoding key for ChainRecord: %w", err)
	}
	data, err := json.Marshal(&AuxChainRecord{
		Active:  rec.Active,
		ChainID: rec.ChainID.String(),
	})
	if err != nil {
		return fmt.Errorf("Error encoding ChainRecord: %w", err)
	}
	return r.store.Set(key, data)
}

// DKShareRegistryProvider implementation
func (r *textImpl) SaveDKShare(dkShare tcrypto.DKShare) error {
	dbKey, err := json.Marshal(dkShare.GetAddress().String())
	if err != nil {
		return fmt.Errorf("error creating key for DKShare: %w", err)
	}

	r.log.Infof("Saving DKShare for address=%v as key %v", dkShare.GetAddress().String(), dbKey)

	var exists bool
	if exists, err = r.store.Has(dbKey); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}
	data, err := hexEncode(dkShare.Bytes())
	if err != nil {
		return fmt.Errorf("Error encoding DKShare: %w", err)
	}
	return r.store.Set(dbKey, data)
}

func (r *textImpl) LoadDKShare(sharedAddress iotago.Address) (tcrypto.DKShare, error) {
	key, err := json.Marshal(sharedAddress.String())
	if err != nil {
		return nil, fmt.Errorf("Error encoding DKShare key: %w", err)
	}
	data, err := r.store.Get(key)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, ErrDKShareNotFound
		}
		return nil, err
	}
	data, err = hexDecode(data)
	if err != nil {
		return nil, fmt.Errorf("Error decoding DKShare: %w", err)
	}
	return tcrypto.DKShareFromBytes(data, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite(), r.nodeIdentity.GetPrivateKey())
}

// TrustedNetworkManager implementation
func (r *textImpl) IsTrustedPeer(pubKey *cryptolib.PublicKey) error {
	tp := &peering.TrustedPeer{PubKey: pubKey}
	tpKeyBytes, err := json.Marshal(tp.PubKey.String())
	if err != nil {
		return err
	}
	_, err = r.store.Get(tpKeyBytes)
	return err // Assume its trusted, if we can decode the entry.
}

func (r *textImpl) TrustPeer(pubKey *cryptolib.PublicKey, netID string) (*peering.TrustedPeer, error) {
	tp := &peering.TrustedPeer{PubKey: pubKey, NetID: netID}
	tpKeyBytes, err := json.Marshal(tp.PubKey.String())
	if err != nil {
		return nil, err
	}

	tpBinary, err := json.Marshal(tp)
	if err != nil {
		return nil, err
	}
	err = r.store.Set(tpKeyBytes, tpBinary)
	if err != nil {
		return nil, err
	}
	return tp, nil
}

func (r *textImpl) DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error) {
	tp := &peering.TrustedPeer{PubKey: pubKey}
	tpKeyBytes, err := json.Marshal(tp.PubKey.String())
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
	err = json.Unmarshal(tpBinary, &tp)
	if err != nil {
		return nil, nil
	}
	return tp, nil
}

func (r *textImpl) TrustedPeers() ([]*peering.TrustedPeer, error) {
	ret := make([]*peering.TrustedPeer, 0)
	err := r.store.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		var tp peering.TrustedPeer
		if err := json.Unmarshal(value, &tp); err == nil {
			ret = append(ret, &tp)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// journal.Registry implementation
func (r *textImpl) LoadConsensusJournal(id journal.ID) (journal.LogIndex, journal.LocalView, error) {
	key, err := hexEncode(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLogIndex))
	if err != nil {
		return 0, nil, err
	}
	liBytes, err := r.store.Get(key)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return 0, nil, journal.ErrConsensusJournalNotFound
		}
		return 0, nil, err
	}
	liBytes, err = hexDecode(liBytes)
	if err != nil {
		return 0, nil, err
	}
	li := journal.LogIndex(binary.BigEndian.Uint32(liBytes))

	key, err = hexEncode(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLocalView))
	if err != nil {
		return 0, nil, err
	}
	lvBytes, err := r.store.Get(key)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return 0, nil, journal.ErrConsensusJournalNotFound
		}
		return 0, nil, err
	}
	lvBytes, err = hexDecode(lvBytes)
	if err != nil {
		return 0, nil, err
	}
	lv, err := journal.NewLocalViewFromBytes(lvBytes)
	if err != nil {
		return 0, nil, fmt.Errorf("cannot deserialize LocalView")
	}

	return li, lv, nil
}

func (r *textImpl) SaveConsensusJournalLogIndex(id journal.ID, logIndex journal.LogIndex) error {
	liBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(liBytes, logIndex.AsUint32())
	liBytes, err := hexEncode(liBytes)
	if err != nil {
		return fmt.Errorf("cannot serialize logIndex: %w", err)
	}
	key, err := hexEncode(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLogIndex))
	if err != nil {
		return fmt.Errorf("cannot serialize key for logIndex: %w", err)
	}
	err = r.store.Set(key, liBytes)
	if err != nil {
		return err
	}
	return nil
}

func (r *textImpl) SaveConsensusJournalLocalView(id journal.ID, localView journal.LocalView) error {
	lvBytes, err := localView.AsBytes()
	if err != nil {
		return fmt.Errorf("cannot serialize localView: %w", err)
	}
	lvBytes, err = hexEncode(lvBytes)
	if err != nil {
		return fmt.Errorf("cannot serialize localView: %w", err)
	}

	key, err := hexEncode(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLocalView))
	if err != nil {
		return fmt.Errorf("cannot serialize key for localView: %w", err)
	}
	if err := r.store.Set(key, lvBytes); err != nil {
		return err
	}
	return nil
}
