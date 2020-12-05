package alone

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
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
	UtxoDB              *utxodb.UtxoDB
	StateTx             *sctransaction.Transaction
	State               state.VirtualState
	Proc                *processors.ProcessorCache
	Log                 *logger.Logger
}

func New(t *testing.T, debug bool) *aloneEnvironment {
	chSig := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	orSig := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	chainID := coretypes.ChainID(chSig.Address())
	log := testutil.NewLogger(t)
	if !debug {
		log = testutil.WithLevel(log, zapcore.InfoLevel)
	}
	env := &aloneEnvironment{
		T:                   t,
		ChainSigscheme:      chSig,
		OriginatorSigscheme: orSig,
		ChainAddress:        chSig.Address(),
		OriginatorAddress:   orSig.Address(),
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

	_, _ = env.runRequest(initTx)
	return env
}

//goland:noinspection ALL
func (e *aloneEnvironment) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", e.ChainID.String())
	fmt.Fprintf(&buf, "Chain address: %s\n", e.ChainAddress.String())
	fmt.Fprintf(&buf, "State hash: %s\n", e.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", e.UtxoDB.GetGenesisAddress().String())
	return string(buf.Bytes())
}

func (e *aloneEnvironment) Infof(format string, args ...interface{}) {
	e.Log.Infof(format, args...)
}

func (e *aloneEnvironment) CheckBalance(addr address.Address, col balance.Color, expected int64) {
	require.EqualValues(e.T, expected, e.GetBalance(addr, col))
}

func (e *aloneEnvironment) GetBalance(addr address.Address, col balance.Color) int64 {
	bals := e.GetColoredBalances(addr)
	ret, _ := bals[col]
	return ret
}

func (e *aloneEnvironment) GetColoredBalances(addr address.Address) map[balance.Color]int64 {
	outs := e.UtxoDB.GetAddressOutputs(addr)
	ret, _ := waspconn.OutputBalancesByColor(outs)
	return ret
}

func (e *aloneEnvironment) CheckBase() {
	req := NewCall(root.Interface.Name, root.FuncGetInfo)
	res1, err := e.PostRequest(req, e.OriginatorSigscheme)
	require.NoError(e.T, err)

	res2, err := e.CallView(req)
	require.NoError(e.T, err)

	require.EqualValues(e.T, res1.Hash(), res2.Hash())
}
