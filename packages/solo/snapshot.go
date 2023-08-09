package solo

import (
	"encoding/json"
	"os"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type soloSnapshot struct {
	UtxoDB *utxodb.UtxoDBState
	Chains []soloChainSnapshot
}

type soloChainSnapshot struct {
	Name                   string
	StateControllerKeyPair []byte
	ChainID                []byte
	OriginatorPrivateKey   []byte
	ValidatorFeeTarget     []byte
	DB                     [][]byte
}

// SaveSnapshot generates a snapshot of the Solo environment
func (env *Solo) SaveSnapshot(fname string) {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()

	snapshot := soloSnapshot{
		UtxoDB: env.utxoDB.State(),
	}

	for _, ch := range env.chains {
		chainSnapshot := soloChainSnapshot{
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

	b, err := json.Marshal(snapshot)
	require.NoError(env.T, err)
	err = os.WriteFile(fname, b, 0o600)
	require.NoError(env.T, err)
}

// LoadSnapshot restores the Solo environment from the given snapshot
func (env *Solo) LoadSnapshot(fname string) {
	env.glbMutex.Lock()
	defer env.glbMutex.Unlock()

	b, err := os.ReadFile(fname)
	require.NoError(env.T, err)
	var snapshot soloSnapshot
	err = json.Unmarshal(b, &snapshot)
	require.NoError(env.T, err)

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

		db, err := env.chainStateDatabaseManager.ChainStateKVStore(chainID)
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
		}
		env.addChain(chainData)
	}
}
