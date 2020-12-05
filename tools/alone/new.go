package alone

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/processors"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
)

type aloneEnvironment struct {
	T                   *testing.T
	ChainSigscheme      signaturescheme.SignatureScheme
	OriginatorSigscheme signaturescheme.SignatureScheme
	ChainID             coretypes.ChainID
	ChainAddress        address.Address
	ChainColor          balance.Color
	OriginatorAddress   address.Address
	OriginatorAgentID   coretypes.AgentID
	UtxoDB              *utxodb.UtxoDB
	StateTx             *sctransaction.Transaction
	State               state.VirtualState
	Proc                *processors.ProcessorCache
	Log                 *logger.Logger
}

var regOnce sync.Once

func New(t *testing.T, debug bool) *aloneEnvironment {
	chSig := signaturescheme.ED25519(ed25519.GenerateKeyPair()) // chain address will be ED25519, not BLS
	orSig := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	chainID := coretypes.ChainID(chSig.Address())
	log := testutil.NewLogger(t)
	if !debug {
		log = testutil.WithLevel(log, zapcore.InfoLevel)
	}
	regOnce.Do(func() {
		err := processors.RegisterVMType(wasmtimevm.VMType, wasmhost.GetProcessor)
		if err != nil {
			log.Panicf("%v: %v", wasmtimevm.VMType, err)
		}
	})

	env := &aloneEnvironment{
		T:                   t,
		ChainSigscheme:      chSig,
		OriginatorSigscheme: orSig,
		ChainAddress:        chSig.Address(),
		OriginatorAddress:   orSig.Address(),
		OriginatorAgentID:   coretypes.NewAgentIDFromAddress(orSig.Address()),
		ChainID:             chainID,
		UtxoDB:              utxodb.New(),
		State:               state.NewVirtualState(mapdb.NewMapDB(), &chainID),
		Proc:                processors.MustNew(),
		Log:                 log,
	}
	_, err := env.UtxoDB.RequestFunds(env.OriginatorAddress)
	require.NoError(t, err)
	env.CheckBalance(env.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	env.StateTx, err = origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		OriginAddress:             env.ChainAddress,
		OriginatorSignatureScheme: env.OriginatorSigscheme,
		AllInputs:                 env.UtxoDB.GetAddressOutputs(env.OriginatorAddress),
	})
	require.NoError(t, err)
	require.NotNil(t, env.StateTx)
	err = env.UtxoDB.AddTransaction(env.StateTx.Transaction)
	require.NoError(t, err)

	env.ChainColor = balance.Color(env.StateTx.ID())

	originBlock := state.MustNewOriginBlock(&env.ChainColor)
	err = env.State.ApplyBlock(originBlock)
	require.NoError(t, err)
	err = env.State.CommitToDb(originBlock)
	require.NoError(t, err)

	initTx, err := origin.NewRootInitRequestTransaction(origin.NewRootInitRequestTransactionParams{
		ChainID:              chainID,
		Description:          "'alone' testing chain",
		OwnerSignatureScheme: env.OriginatorSigscheme,
		AllInputs:            env.UtxoDB.GetAddressOutputs(env.OriginatorAddress),
	})
	require.NoError(t, err)
	require.NotNil(t, initTx)

	_, err = env.runRequest(initTx)
	require.NoError(t, err)
	return env
}
