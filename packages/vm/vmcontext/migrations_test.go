package vmcontext

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

type migrationsTestEnv struct {
	t     *testing.T
	db    kvstore.KVStore
	cs    state.Store
	vmctx *VMContext

	counter    int
	incCounter migrations.Migration
	panic      migrations.Migration
}

func (e *migrationsTestEnv) getSchemaVersion() (ret uint32) {
	e.vmctx.withStateUpdate(func() {
		e.vmctx.callCore(root.Contract, func(s kv.KVStore) {
			ret = root.GetSchemaVersion(s)
		})
	})
	return
}

func (e *migrationsTestEnv) setSchemaVersion(v uint32) {
	e.vmctx.withStateUpdate(func() {
		e.vmctx.callCore(root.Contract, func(s kv.KVStore) {
			root.SetSchemaVersion(s, v)
		})
	})
}

func newMigrationsTest(t *testing.T, stateIndex uint32) *migrationsTestEnv {
	db := mapdb.NewMapDB()
	cs := state.InitChainStore(db)
	latest, err := cs.LatestBlock()
	assert.NoError(t, err)
	stateDraft, err := cs.NewStateDraft(time.Now(), latest.L1Commitment())
	assert.NoError(t, err)
	vmctx := &VMContext{
		task: &vm.VMTask{
			StateDraft: stateDraft,
			AnchorOutput: &iotago.AliasOutput{
				StateIndex: stateIndex,
			},
		},
	}

	env := &migrationsTestEnv{
		t:     t,
		db:    db,
		cs:    cs,
		vmctx: vmctx,
	}

	env.incCounter = migrations.Migration{
		Contract: governance.Contract,
		Apply: func(state kv.KVStore, log *zap.SugaredLogger) error {
			env.counter++
			return nil
		},
	}

	env.panic = migrations.Migration{
		Contract: governance.Contract,
		Apply: func(state kv.KVStore, log *zap.SugaredLogger) error {
			panic("should not be called")
		},
	}

	return env
}

func TestMigrationsStateIndex0(t *testing.T) {
	env := newMigrationsTest(t, 0)

	require.EqualValues(t, 0, env.getSchemaVersion())

	env.vmctx.withStateUpdate(func() {
		env.vmctx.runMigrations(0, []migrations.Migration{env.panic, env.panic, env.panic})
	})

	require.EqualValues(t, 3, env.getSchemaVersion())
}

func TestMigrationsStateIndex1(t *testing.T) {
	env := newMigrationsTest(t, 1)

	require.EqualValues(t, 0, env.getSchemaVersion())

	env.vmctx.withStateUpdate(func() {
		env.vmctx.runMigrations(0, []migrations.Migration{env.incCounter, env.incCounter, env.incCounter})
	})

	require.EqualValues(t, 3, env.counter)
	require.EqualValues(t, 3, env.getSchemaVersion())
}

func TestMigrationsStateIndex1Current1(t *testing.T) {
	env := newMigrationsTest(t, 1)

	env.setSchemaVersion(1)

	env.vmctx.withStateUpdate(func() {
		env.vmctx.runMigrations(0, []migrations.Migration{env.panic, env.incCounter, env.incCounter})
	})

	require.EqualValues(t, 2, env.counter)
	require.EqualValues(t, 3, env.getSchemaVersion())
}

func TestMigrationsStateIndex1Current2Base1(t *testing.T) {
	env := newMigrationsTest(t, 1)

	env.setSchemaVersion(2)

	env.vmctx.withStateUpdate(func() {
		env.vmctx.runMigrations(1, []migrations.Migration{env.panic, env.incCounter, env.incCounter})
	})

	require.EqualValues(t, 2, env.counter)
	require.EqualValues(t, 4, env.getSchemaVersion())
}
