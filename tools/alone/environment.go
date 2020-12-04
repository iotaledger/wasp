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
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type Environment struct {
	T                   *testing.T
	ChainSigscheme      signaturescheme.SignatureScheme
	OriginatorSigscheme signaturescheme.SignatureScheme
	ChainID             coretypes.ChainID
	ChainAddress        address.Address
	ChainColor          balance.Color
	OriginatorAddress   address.Address
	UtxoDB              *utxodb.UtxoDB
	State               state.VirtualState
	Proc                *processors.ProcessorCache
	Log                 *logger.Logger
}

func NewEnvironment(t *testing.T) *Environment {
	ch := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	or := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	chainID := coretypes.NewRandomChainID()
	ret := &Environment{
		T:                   t,
		ChainSigscheme:      ch,
		OriginatorSigscheme: or,
		ChainAddress:        ch.Address(),
		OriginatorAddress:   or.Address(),
		ChainID:             coretypes.ChainID(ch.Address()),
		UtxoDB:              utxodb.New(),
		State:               state.NewVirtualState(mapdb.NewMapDB(), &chainID),
		Proc:                processors.MustNew(),
		Log:                 testutil.NewLogger(t),
	}
	_, err := ret.UtxoDB.RequestFunds(ret.OriginatorAddress)
	require.NoError(t, err)
	ret.CheckBalance(ret.OriginatorAddress, balance.ColorIOTA, testutil.RequestFundsAmount)

	origTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		OriginAddress:             ret.ChainAddress,
		OriginatorSignatureScheme: ret.OriginatorSigscheme,
		AllInputs:                 ret.UtxoDB.GetAddressOutputs(ret.OriginatorAddress),
	})
	require.NoError(t, err)
	require.NotNil(t, origTx)
	err = ret.UtxoDB.AddTransaction(origTx.Transaction)
	require.NoError(t, err)
	ret.ChainColor = balance.Color(origTx.ID())

	originBlock := state.MustNewOriginBlock(&ret.ChainColor)
	err = ret.State.ApplyBlock(originBlock)
	require.NoError(t, err)
	err = ret.State.CommitToDb(originBlock)
	require.NoError(t, err)

	ret.Infof("Origin transaction: %s", origTx.ID().String())

	initTx, err := origin.NewRootInitRequestTransaction(origin.NewRootInitRequestTransactionParams{
		ChainID:              chainID,
		Description:          "'alone' testing chain",
		OwnerSignatureScheme: ret.OriginatorSigscheme,
		AllInputs:            ret.UtxoDB.GetAddressOutputs(ret.OriginatorAddress),
	})
	require.NoError(t, err)
	require.NotNil(t, initTx)
	err = ret.UtxoDB.AddTransaction(initTx.Transaction)
	require.NoError(t, err)
	ret.Infof("Init transaction: %s", initTx.ID().String())

	task := &vm.VMTask{
		Processors:   ret.Proc,
		ChainID:      ret.ChainID,
		Color:        ret.ChainColor,
		Entropy:      *hashing.RandomHash(nil),
		Balances:     waspconn.OutputsToBalances(ret.UtxoDB.GetAddressOutputs(ret.ChainAddress)),
		Requests:     []sctransaction.RequestRef{{Tx: initTx}},
		Timestamp:    origTx.MustState().Timestamp(),
		VirtualState: ret.State,
		Log:          ret.Log,
	}

	var wg sync.WaitGroup
	task.OnFinish = func(err error) {
		require.NoError(t, err)

		ret.Infof("root.init:\nResult tx: %s Result block essence hash: %s",
			task.ResultTransaction.ID().String(), task.ResultBlock.EssenceHash().String())
		wg.Done()
	}

	wg.Add(1)
	err = runvm.RunComputationsAsync(task)
	require.NoError(t, err)
	wg.Wait()

	return ret
}

//goland:noinspection ALL
func (e *Environment) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", e.ChainID.String())
	fmt.Fprintf(&buf, "Chain address: %s\n", e.ChainAddress.String())
	fmt.Fprintf(&buf, "State hash: %s\n", e.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", e.UtxoDB.GetGenesisAddress().String())
	return string(buf.Bytes())
}

func (e *Environment) Infof(format string, args ...interface{}) {
	e.Log.Infof(format, args...)
}

func (e *Environment) CheckBalance(addr address.Address, col balance.Color, expected int64) {
	require.EqualValues(e.T, expected, e.GetBalance(addr, col))
}

func (e *Environment) GetBalance(addr address.Address, col balance.Color) int64 {
	bals := e.GetColoredBalances(addr)
	ret, _ := bals[col]
	return ret
}

func (e *Environment) GetColoredBalances(addr address.Address) map[balance.Color]int64 {
	outs := e.UtxoDB.GetAddressOutputs(addr)
	ret, _ := waspconn.OutputBalancesByColor(outs)
	return ret
}
