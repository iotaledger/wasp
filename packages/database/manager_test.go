package database

import (
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/v2/packages/kvstore"

	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/state"
)

func TestNewChainStateDatabaseManager(t *testing.T) {
	chainRecordRegistry, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainStateDatabaseManager, err := NewChainStateDatabaseManager(chainRecordRegistry, WithEngine(hivedb.EngineMapDB))
	require.NoError(t, err)

	require.Empty(t, chainStateDatabaseManager.databases)
}

func TestCreateChainStateDatabase(t *testing.T) {
	chainRecordRegistry, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainStateDatabaseManager, err := NewChainStateDatabaseManager(chainRecordRegistry, WithEngine(hivedb.EngineMapDB))
	require.NoError(t, err)

	chainID := isctest.RandomChainID()
	store, _ := chainStateDatabaseManager.chainStateKVStore(chainID)
	require.Nil(t, store)
	store, _, err = chainStateDatabaseManager.ChainStateKVStore(chainID)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Len(t, chainStateDatabaseManager.databases, 1)
}

// go test -tags rocksdb ./packages/database/ --run TestWriteAmplification -v --count=1 --timeout=30m
// See <https://github.com/EighteenZi/rocksdb_wiki/blob/master/RocksDB-Tuning-Guide.md>.
// On compaction: <https://vinodhinic.medium.com/lets-rock-3a73fbc6ea79>
// Misc options: <https://github.com/facebook/rocksdb/blob/master/include/rocksdb/options.h>
func TestWriteAmplification(t *testing.T) {
	t.Skip("that's heavy, enable only if needed, only works on linux.")
	printIoStats("IO-Start", t)

	chainID := isctest.RandomChainID()

	chainRecordRegistry, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)
	chainRecordRegistry.ActivateChainRecord(chainID)

	tempDir := fmt.Sprintf("/tmp/TestWriteAmplification-%v", time.Now().UnixMilli())

	chainStateDatabaseManager, err := NewChainStateDatabaseManager(
		chainRecordRegistry,
		WithEngine(hivedb.EngineRocksDB),
		WithPath(tempDir),
	)
	require.NoError(t, err)

	chainKVStore, writeMutex, err := chainStateDatabaseManager.ChainStateKVStore(chainID)
	require.NoError(t, err)
	countKVStore := newCountingKVStore(chainKVStore)
	chainStore := state.NewStore(countKVStore, writeMutex)
	require.NotNil(t, chainStore)

	originSD := chainStore.NewOriginStateDraft()
	originSD.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode[uint32](uint32(0)))
	originSD.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Encode[time.Time](time.Unix(0, 0)))
	originBlock := chainStore.Commit(originSD)
	require.NotNil(t, originBlock)

	b := originBlock
	allBytes := 0
	for i := 0; i < 10_000; i++ {
		sd, sdErr := chainStore.NewStateDraft(time.Now(), b.L1Commitment())
		require.NoError(t, sdErr)
		for j := 0; j < 100; j++ {
			k := kv.Key(fmt.Sprintf("key-%d-%d", 0, j)) // NOTE: independent of i.
			v := []byte(fmt.Sprintf("v v v a a a a a a l l l l u u u e e e e e-%d-%d", i, j))
			allBytes += len(k) + len(v)
			sd.Set(k, v)
		}
		b = chainStore.Commit(sd)
	}

	printDu("DU-End", tempDir, t)
	printIoStats("IO-End", t)
	t.Logf("AllBytes=%v, kvStore=%v\n", allBytes, countKVStore)
}

func printDu(name, tempDir string, t *testing.T) {
	out, err := exec.Command("du", "-bs", tempDir).CombinedOutput()
	require.NoError(t, err)
	t.Logf("%s:\n%s", name, out)
}

func printIoStats(name string, t *testing.T) {
	out, err := exec.Command("cat", fmt.Sprintf("/proc/%d/io", os.Getpid())).CombinedOutput()
	require.NoError(t, err)
	t.Logf("%s:\n%s", name, out)
}

type countingKVStore struct {
	nested kvstore.KVStore
	wBytes *atomic.Uint64
	wCount *atomic.Uint64
	wBatch *atomic.Uint64
	rCount *atomic.Uint64
	rScan  *atomic.Uint64
	fCount *atomic.Uint64
}

var _ kvstore.KVStore = &countingKVStore{}

func newCountingKVStore(nested kvstore.KVStore) *countingKVStore {
	return &countingKVStore{
		nested: nested,
		wBytes: &atomic.Uint64{},
		wCount: &atomic.Uint64{},
		wBatch: &atomic.Uint64{},
		rCount: &atomic.Uint64{},
		rScan:  &atomic.Uint64{},
		fCount: &atomic.Uint64{},
	}
}

func (ckv *countingKVStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	kvs, err := ckv.nested.WithRealm(realm)
	if err != nil {
		return kvs, err
	}
	return &countingKVStore{nested: kvs, wBytes: ckv.wBytes, wCount: ckv.wCount, wBatch: ckv.wBatch, rCount: ckv.rCount, rScan: ckv.rScan, fCount: ckv.fCount}, nil
}

func (ckv *countingKVStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	kvs, err := ckv.nested.WithExtendedRealm(realm)
	if err != nil {
		return kvs, err
	}
	return &countingKVStore{nested: kvs, wBytes: ckv.wBytes, wCount: ckv.wCount, wBatch: ckv.wBatch, rCount: ckv.rCount, rScan: ckv.rScan, fCount: ckv.fCount}, nil
}

func (ckv *countingKVStore) Realm() kvstore.Realm {
	return ckv.nested.Realm()
}

func (ckv *countingKVStore) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {
	ckv.rScan.Add(1)
	return ckv.nested.Iterate(prefix, kvConsumerFunc, direction...)
}

func (ckv *countingKVStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {
	ckv.rScan.Add(1)
	return ckv.nested.IterateKeys(prefix, consumerFunc, direction...)
}

func (ckv *countingKVStore) Clear() error {
	return ckv.nested.Clear()
}

func (ckv *countingKVStore) Get(key kvstore.Key) (value kvstore.Value, err error) {
	ckv.rCount.Add(1)
	return ckv.nested.Get(key)
}

func (ckv *countingKVStore) MultiGet(keys []kvstore.Key) (values []kvstore.Value, err error) {
	ckv.rCount.Add(1)
	return ckv.nested.MultiGet(keys)
}

func (ckv *countingKVStore) Set(key kvstore.Key, value kvstore.Value) error {
	ckv.wBytes.Add(uint64(len(key) + len(value)))
	ckv.wCount.Add(1)
	return ckv.nested.Set(key, value)
}

func (ckv *countingKVStore) Has(key kvstore.Key) (bool, error) {
	ckv.rCount.Add(1)
	return ckv.nested.Has(key)
}

func (ckv *countingKVStore) Delete(key kvstore.Key) error {
	return ckv.nested.Delete(key)
}

func (ckv *countingKVStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	return ckv.nested.DeletePrefix(prefix)
}

func (ckv *countingKVStore) Flush() error {
	ckv.fCount.Add(1)
	return ckv.nested.Flush()
}

func (ckv *countingKVStore) Close() error {
	return ckv.nested.Close()
}

func (ckv *countingKVStore) Batched() (kvstore.BatchedMutations, error) {
	kvb, err := ckv.nested.Batched()
	if err != nil {
		return kvb, err
	}
	return &countingKVBatch{nested: kvb, store: ckv}, nil
}

func (ckv *countingKVStore) String() string {
	return fmt.Sprintf(
		"{wBytes=%v, wCount=%v, wBatch=%v, rCount=%v, rScan=%v, fCount=%v}",
		ckv.wBytes.Load(), ckv.wCount.Load(), ckv.wBatch.Load(), ckv.rCount.Load(), ckv.rScan.Load(), ckv.fCount.Load(),
	)
}

type countingKVBatch struct {
	nested kvstore.BatchedMutations
	store  *countingKVStore
}

func (ckv *countingKVBatch) Set(key kvstore.Key, value kvstore.Value) error {
	ckv.store.wBytes.Add(uint64(len(key) + len(value)))
	ckv.store.wBatch.Add(1)
	return ckv.nested.Set(key, value)
}

func (ckv *countingKVBatch) Delete(key kvstore.Key) error {
	return ckv.nested.Delete(key)
}

func (ckv *countingKVBatch) Cancel() {
	ckv.nested.Cancel()
}

func (ckv *countingKVBatch) Commit() error {
	return ckv.nested.Commit()
}
