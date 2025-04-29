package solo

type Snapshot struct {
	// UtxoDB *utxodb.UtxoDBState
	Chains []*ChainSnapshot
}

type ChainSnapshot struct {
	Name                   string
	StateControllerKeyPair []byte
	ChainID                []byte
	OwnerPrivateKey        []byte
	ValidatorFeeTarget     []byte
	DB                     [][]byte
}

// TakeSnapshot generates a snapshot of the Solo environment
func (env *Solo) TakeSnapshot() *Snapshot {
	panic("TODO")
	// env.chainsMutex.Lock()
	// defer env.chainsMutex.Unlock()
	//
	// snapshot := &Snapshot{}
	//
	// for _, ch := range env.chains {
	// 	chainSnapshot := &ChainSnapshot{
	// 		Name:                   ch.Name,
	// 		StateControllerKeyPair: rwutil.WriteToBytes(ch.StateControllerKeyPair),
	// 		ChainID:                ch.ChainID.Bytes(),
	// 		OwnerPrivateKey:   rwutil.WriteToBytes(ch.OwnerPrivateKey),
	// 		ValidatorFeeTarget:     ch.ValidatorFeeTarget.Bytes(),
	// 	}
	//
	// 	err := ch.db.Iterate(kvstore.EmptyPrefix, func(k, v []byte) bool {
	// 		chainSnapshot.DB = append(chainSnapshot.DB, k, v)
	// 		return true
	// 	})
	// 	require.NoError(env.T, err)
	//
	// 	snapshot.Chains = append(snapshot.Chains, chainSnapshot)
	// }
	//
	// return snapshot
}

// RestoreSnapshot restores the Solo environment from the given snapshot
func (env *Solo) RestoreSnapshot(snapshot *Snapshot) {
	panic("TODO")
	// env.chainsMutex.Lock()
	// defer env.chainsMutex.Unlock()
	//
	// env.utxoDB.SetState(snapshot.UtxoDB)
	// for _, chainSnapshot := range snapshot.Chains {
	// 	sckp, err := rwutil.ReadFromBytes(chainSnapshot.StateControllerKeyPair, new(cryptolib.KeyPair))
	// 	require.NoError(env.T, err)
	//
	// 	chainID, err := isc.ChainIDFromBytes(chainSnapshot.ChainID)
	// 	require.NoError(env.T, err)
	//
	// 	okp, err := rwutil.ReadFromBytes(chainSnapshot.OwnerPrivateKey, new(cryptolib.KeyPair))
	// 	require.NoError(env.T, err)
	//
	// 	val, err := isc.AgentIDFromBytes(chainSnapshot.ValidatorFeeTarget)
	// 	require.NoError(env.T, err)
	//
	// 	db, _, err := env.chainStateDatabaseManager.ChainStateKVStore(chainID)
	// 	require.NoError(env.T, err)
	// 	for i := 0; i < len(chainSnapshot.DB); i += 2 {
	// 		err = db.Set(chainSnapshot.DB[i], chainSnapshot.DB[i+1])
	// 		require.NoError(env.T, err)
	// 	}
	//
	// 	chainData := chainData{
	// 		Name:                   chainSnapshot.Name,
	// 		StateControllerKeyPair: sckp,
	// 		ChainID:                chainID,
	// 		OwnerPrivateKey:   okp,
	// 		ValidatorFeeTarget:     val,
	// 		db:                     db,
	// 		migrationScheme:        &migrations.MigrationScheme{},
	// 	}
	// 	env.addChain(chainData)
	// }
}

// SaveSnapshot saves the given snapshot to a file
func (env *Solo) SaveSnapshot(snapshot *Snapshot, fname string) {
	panic("TODO")
	// b, err := json.Marshal(snapshot)
	// require.NoError(env.T, err)
	// err = os.WriteFile(fname, b, 0o600)
	// require.NoError(env.T, err)
}

// LoadSnapshot loads a snapshot previously saved with SaveSnapshot
func (env *Solo) LoadSnapshot(fname string) *Snapshot {
	panic("TODO")
	// b, err := os.ReadFile(fname)
	// require.NoError(env.T, err)
	// var snapshot Snapshot
	// err = json.Unmarshal(b, &snapshot)
	// require.NoError(env.T, err)
	// return &snapshot
}
