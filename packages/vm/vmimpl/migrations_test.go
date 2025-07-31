package vmimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

type migrationsTestEnv struct {
	t     *testing.T
	db    kvstore.KVStore
	cs    state.Store
	vmctx *vmContext

	counter    int
	incCounter migrations.Migration
	panic      migrations.Migration
}

func (e *migrationsTestEnv) getSchemaVersion() (ret isc.SchemaVersion) {
	e.vmctx.withStateUpdate(func(chainState kv.KVStore) {
		ret = root.NewStateReaderFromChainState(chainState).GetSchemaVersion()
	})
	return
}

func (e *migrationsTestEnv) setSchemaVersion(v isc.SchemaVersion) {
	e.vmctx.withStateUpdate(func(chainState kv.KVStore) {
		root.NewStateWriter(root.Contract.StateSubrealm(chainState)).SetSchemaVersion(v)
	})
}

func newMigrationsTest(t *testing.T) *migrationsTestEnv {
	db := mapdb.NewMapDB()
	cs := statetest.NewStoreWithUniqueWriteMutex(db)
	chainCreator := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("chainCreator")))
	anchor := initChain(chainCreator, cs)
	latest, err := cs.LatestBlock()
	require.NoError(t, err)
	stateDraft, err := cs.NewStateDraft(time.Now(), latest.L1Commitment())
	require.NoError(t, err)
	task := &vm.VMTask{
		Anchor: anchor,
	}
	vmctx := &vmContext{
		task:       task,
		stateDraft: stateDraft,
	}
	vmctx.loadChainConfig()

	env := &migrationsTestEnv{
		t:     t,
		db:    db,
		cs:    cs,
		vmctx: vmctx,
	}

	env.incCounter = migrations.Migration{
		Contract: governance.Contract,
		Apply: func(state kv.KVStore, log log.Logger) error {
			env.counter++
			return nil
		},
	}

	env.panic = migrations.Migration{
		Contract: governance.Contract,
		Apply: func(state kv.KVStore, log log.Logger) error {
			panic("should not be called")
		},
	}

	return env
}

func TestMigrations(t *testing.T) {
	env := newMigrationsTest(t)

	require.EqualValues(t, allmigrations.LatestSchemaVersion, env.getSchemaVersion())

	env.vmctx.withStateUpdate(func(chainState kv.KVStore) {
		env.vmctx.runMigrations(chainState, &migrations.MigrationScheme{
			BaseSchemaVersion: allmigrations.LatestSchemaVersion,
			Migrations:        []migrations.Migration{env.incCounter, env.incCounter, env.incCounter},
		})
	})

	require.EqualValues(t, 3, env.counter)
	require.EqualValues(t, allmigrations.LatestSchemaVersion+3, env.getSchemaVersion())
}

func TestMigrationsCurrent1(t *testing.T) {
	env := newMigrationsTest(t)

	env.setSchemaVersion(1)

	env.vmctx.withStateUpdate(func(chainState kv.KVStore) {
		env.vmctx.runMigrations(chainState, &migrations.MigrationScheme{
			BaseSchemaVersion: 0,
			Migrations:        []migrations.Migration{env.panic, env.incCounter, env.incCounter},
		})
	})

	require.EqualValues(t, 2, env.counter)
	require.EqualValues(t, 3, env.getSchemaVersion())
}

func TestMigrationsCurrent2Base1(t *testing.T) {
	env := newMigrationsTest(t)

	env.setSchemaVersion(2)

	env.vmctx.withStateUpdate(func(chainState kv.KVStore) {
		env.vmctx.runMigrations(chainState, &migrations.MigrationScheme{
			BaseSchemaVersion: 1,
			Migrations:        []migrations.Migration{env.panic, env.incCounter, env.incCounter},
		})
	})

	require.EqualValues(t, 2, env.counter)
	require.EqualValues(t, 4, env.getSchemaVersion())
}
