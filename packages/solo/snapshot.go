package solo

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Snapshot struct {
	UtxoDB *utxodb.UtxoDBState
	Chains []*ChainSnapshot
}

type ChainSnapshot struct {
	Name                   string
	StateControllerKeyPair []byte
	ChainID                []byte
	OriginatorPrivateKey   []byte
	ValidatorFeeTarget     []byte
	DB                     [][]byte
}

// SaveSnapshot generates a snapshot of the Solo environment
func (env *Solo) TakeSnapshot() *Snapshot {
	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()

	snapshot := &Snapshot{
		UtxoDB: env.utxoDB.State(),
	}

	for _, ch := range env.chains {
		chainSnapshot := &ChainSnapshot{
			Name:                   ch.Name,
			StateControllerKeyPair: rwutil.WriteToBytes(ch.StateControllerKeyPair),
			ChainID:                ch.ChainID.Bytes(),
			OriginatorPrivateKey:   rwutil.WriteToBytes(ch.OriginatorPrivateKey),
			ValidatorFeeTarget:     ch.ValidatorFeeTarget.Bytes(),
		}

		err := ch.db.Iterate(kvstore.EmptyPrefix, func(k, v []byte) bool {
			chainSnapshot.DB = append(chainSnapshot.DB, k, v)
			return true
		})
		require.NoError(env.T, err)

		snapshot.Chains = append(snapshot.Chains, chainSnapshot)
	}

	return snapshot
}

// LoadSnapshot restores the Solo environment from the given snapshot
func (env *Solo) RestoreSnapshot(snapshot *Snapshot) {
	env.chainsMutex.Lock()
	defer env.chainsMutex.Unlock()

	env.utxoDB.SetState(snapshot.UtxoDB)
	for _, chainSnapshot := range snapshot.Chains {
		sckp, err := rwutil.ReadFromBytes(chainSnapshot.StateControllerKeyPair, new(cryptolib.KeyPair))
		require.NoError(env.T, err)

		chainID, err := isc.ChainIDFromBytes(chainSnapshot.ChainID)
		require.NoError(env.T, err)

		okp, err := rwutil.ReadFromBytes(chainSnapshot.OriginatorPrivateKey, new(cryptolib.KeyPair))
		require.NoError(env.T, err)

		val, err := isc.AgentIDFromBytes(chainSnapshot.ValidatorFeeTarget)
		require.NoError(env.T, err)

		db, _, err := env.chainStateDatabaseManager.ChainStateKVStore(chainID)
		require.NoError(env.T, err)
		for i := 0; i < len(chainSnapshot.DB); i += 2 {
			err = db.Set(chainSnapshot.DB[i], chainSnapshot.DB[i+1])
			require.NoError(env.T, err)
		}

		chainData := chainData{
			Name:                   chainSnapshot.Name,
			StateControllerKeyPair: sckp,
			ChainID:                chainID,
			OriginatorPrivateKey:   okp,
			ValidatorFeeTarget:     val,
			db:                     db,
			writeMutex:             &sync.Mutex{},
		}
		env.addChain(chainData)
	}
}

// SaveSnapshot saves the given snapshot to a file
func (env *Solo) SaveSnapshot(snapshot *Snapshot, fname string) {
	b, err := json.Marshal(snapshot)
	require.NoError(env.T, err)
	err = os.WriteFile(fname, b, 0o600)
	require.NoError(env.T, err)
}

// LoadSnapshot loads a snapshot previously saved with SaveSnapshot
func (env *Solo) LoadSnapshot(fname string) *Snapshot {
	b, err := os.ReadFile(fname)
	require.NoError(env.T, err)
	var snapshot Snapshot
	err = json.Unmarshal(b, &snapshot)
	require.NoError(env.T, err)
	return &snapshot
}
