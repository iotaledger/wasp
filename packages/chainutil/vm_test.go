package chainutil_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/chainutil"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

var schemaVersion = allmigrations.DefaultScheme.LatestSchemaVersion()

// initChain initializes a new chain state on the given empty store, and returns a fake L1
// anchor with a random ObjectID and the corresponding StateMetadata.
func initChain(chainCreator *cryptolib.KeyPair, store state.Store) *isc.StateAnchor {
	// create the anchor for a new chain
	initParams := origin.NewInitParams(
		isc.NewAddressAgentID(chainCreator.Address()),
		evm.DefaultChainID,
		governance.DefaultBlockKeepAmount,
		false,
	).Encode()
	const originDeposit = 1 * isc.Million

	_, stateMetadata := origin.InitChain(
		schemaVersion,
		store,
		initParams,
		iotago.ObjectID{},
		originDeposit,
		parameterstest.L1Mock,
		nil, nil,
	)
	stateMetadataBytes := stateMetadata.Bytes()
	anchor := iscmove.Anchor{
		ID:            *iotatest.RandomAddress(),
		StateMetadata: stateMetadataBytes,
		StateIndex:    0,
		Assets: iscmove.Referent[iscmove.AssetsBag]{
			ID: *iotatest.RandomAddress(),
			Value: &iscmove.AssetsBag{
				ID:   *iotatest.RandomAddress(),
				Size: 1,
			},
		},
	}
	stateAnchor := isc.NewStateAnchor(
		&iscmove.AnchorWithRef{
			ObjectRef: iotago.ObjectRef{
				ObjectID: &anchor.ID,
				Version:  0,
			},
			Object: &anchor,
			Owner:  chainCreator.Address().AsIotaAddress(),
		},
		iotago.PackageID{},
	)
	return &stateAnchor
}

func TestEVMCall(t *testing.T) {
	chainCreator := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("chainCreator")))
	store := indexedstore.New(statetest.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()))
	anchor := initChain(chainCreator, store)

	magicContract := common.HexToAddress("1074000000000000000000000000000000000000")
	contractCall := "564b81ef" // getChainID()
	selector, err := hex.DecodeString(contractCall)
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return
	}

	msg := ethereum.CallMsg{
		From:      common.Address{},
		To:        &magicContract,
		Data:      selector,
		GasPrice:  big.NewInt(0),
		GasFeeCap: big.NewInt(0),
		GasTipCap: big.NewInt(0),
		Value:     big.NewInt(0),
	}

	logger := testlogger.NewLogger(t)

	result, err := chainutil.EVMCall(anchor, parameterstest.L1Mock, store, coreprocessors.NewConfig(), logger, msg)
	if err != nil {
		t.Fatalf("failed to call EVM: %v", err)
	}

	if !bytes.Equal(result, anchor.ChainID().Bytes()) {
		t.Fatalf("received wrong chain ID from evm. expected: %x, got: %x", anchor.ChainID().Bytes(), result[3:])
	}
}
