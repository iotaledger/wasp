package vmimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func TestNFTDepositNoIssuer(t *testing.T) {
	metadata := isc.RequestMetadata{Message: accounts.FuncDeposit.Message()}
	o := &iotago.NFTOutput{
		Amount:       100 * isc.Million,
		NativeTokens: []*isc.NativeToken{},
		NFTID:        isc.NFTID{0x1},
		Conditions:   []iotago.UnlockCondition{},
		Features: []iotago.Feature{
			&iotago.MetadataFeature{
				Data: metadata.Bytes(),
			},
			&iotago.SenderFeature{
				Address: tpkg.RandEd25519Address(),
			},
		},
		ImmutableFeatures: []iotago.Feature{
			&iotago.MetadataFeature{
				Data: []byte("foobar"),
			},
		},
	}

	res := simulateRunOutput(t, o)
	require.Len(t, res.RequestResults, 1)
	require.Nil(t, res.RequestResults[0].Receipt.Error)
}

func simulateRunOutput(
	t *testing.T,
	request *iscmove.RefWithObject[iscmove.Request],
	anchorAddress *cryptolib.Address,
) *vm.VMTaskResult {
	store := indexedstore.New(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))

	req, err := isc.OnLedgerFromRequest(request, anchorAddress)
	require.NoError(t, err)

	// create the AO for a new chain
	chainCreator := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("foobar")))
	originBlock := origin.InitChain(0, store, nil)
	_, chainAO, _, err := origin.NewChainOriginTransaction(
		chainCreator,
		chainCreator.Address(),
		chainCreator.Address(),
		10*isc.Million,
		nil,
		iotago.OutputSet{
			iotago.OutputID{}: &iotago.BasicOutput{
				Amount:       1000 * isc.Million,
				NativeTokens: []*isc.NativeToken{},
				Conditions:   []iotago.UnlockCondition{},
				Features:     []iotago.Feature{},
			},
		},
		iotago.OutputIDs{{}},
		0,
	)
	require.NoError(t, err)

	// create task and run it
	task := &vm.VMTask{
		Processors: processors.MustNew(coreprocessors.NewConfigWithCoreContracts()),
		Anchor: &isc.StateAnchor{
			Ref:   &iscmove.RefWithObject[iscmove.Anchor]{},
			Owner: chainCreator.Address(),
		},
		Store:                store,
		Requests:             []isc.Request{req},
		Timestamp:            time.Time{},
		Entropy:              [32]byte{},
		ValidatorFeeTarget:   nil,
		EstimateGasMode:      false,
		EVMTracer:            &isc.EVMTracer{},
		EnableGasBurnLogging: false,
		Migrations:           &migrations.MigrationScheme{},
		Log:                  testlogger.NewLogger(t),
	}

	chainAOWithID := isc.NewAliasOutputWithID(chainAO, chainAOID)
	origin.InitChainByAliasOutput(task.Store, chainAOWithID)

	return runTask(task)
}
