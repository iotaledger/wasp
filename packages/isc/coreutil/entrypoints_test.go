package coreutil_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

var _ isc.Sandbox = MockSandBox{}

type MockSandBox struct {
	MockParams isc.CallArguments
}

func (m MockSandBox) RequireCaller(agentID isc.AgentID) {
	panic("implement me")
}

func (m MockSandBox) RequireCallerAnyOf(agentID []isc.AgentID) {
	panic("implement me")
}

func (m MockSandBox) RequireCallerIsChainAdmin() {
	panic("implement me")
}

func (m MockSandBox) State() kv.KVStore {
	panic("implement me")
}

func (m MockSandBox) Request() isc.Calldata {
	panic("implement me")
}

func (m MockSandBox) Call(msg isc.Message, allowance *isc.Assets) isc.CallArguments {
	panic("implement me")
}

func (m MockSandBox) Event(topic string, payload []byte) {
	panic("implement me")
}

func (m MockSandBox) RegisterError(messageFormat string) *isc.VMErrorTemplate {
	panic("implement me")
}

func (m MockSandBox) GetEntropy() hashing.HashValue {
	panic("implement me")
}

func (m MockSandBox) TransferAllowedFunds(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	panic("implement me")
}

func (m MockSandBox) Send(metadata isc.RequestParameters) {
	panic("implement me")
}

func (m MockSandBox) StateIndex() uint32 {
	panic("implement me")
}

func (m MockSandBox) RequestIndex() uint16 {
	panic("implement me")
}

func (m MockSandBox) EVMTracer() *tracers.Tracer {
	panic("implement me")
}

func (m MockSandBox) TakeStateSnapshot() int {
	panic("implement me")
}

func (m MockSandBox) RevertToStateSnapshot(i int) {
	panic("implement me")
}

func (m MockSandBox) Privileged() isc.Privileged {
	panic("implement me")
}

func (m MockSandBox) Requiref(cond bool, format string, args ...any) {
	panic("implement me")
}

func (m MockSandBox) RequireNoError(err error, str ...string) {
	if err != nil {
		panic(str)
	}
}

func (m MockSandBox) BaseTokensBalance() (bts coin.Value, remainder *big.Int) {
	panic("implement me")
}

func (m MockSandBox) CoinBalance(coinType coin.Type) coin.Value {
	panic("implement me")
}

func (m MockSandBox) CoinBalances() isc.CoinBalances {
	panic("implement me")
}

func (m MockSandBox) OwnedObjects() []isc.IotaObject {
	panic("implement me")
}

func (m MockSandBox) HasInAccount(id isc.AgentID, assets *isc.Assets) bool {
	panic("implement me")
}

func (m MockSandBox) Params() isc.CallArguments {
	return m.MockParams
}

func (m MockSandBox) ChainID() isc.ChainID {
	panic("implement me")
}

func (m MockSandBox) ChainAdmin() isc.AgentID {
	panic("implement me")
}

func (m MockSandBox) ChainInfo() *isc.ChainInfo {
	panic("implement me")
}

func (m MockSandBox) Contract() isc.Hname {
	panic("implement me")
}

func (m MockSandBox) AccountID() isc.AgentID {
	panic("implement me")
}

func (m MockSandBox) Caller() isc.AgentID {
	panic("implement me")
}

func (m MockSandBox) Timestamp() time.Time {
	panic("implement me")
}

func (m MockSandBox) Log() isc.LogInterface {
	panic("implement me")
}

func (m MockSandBox) Utils() isc.Utils {
	panic("implement me")
}

func (m MockSandBox) Gas() isc.Gas {
	panic("implement me")
}

func (m MockSandBox) GetObjectBCS(id iotago.ObjectID) ([]byte, bool) {
	panic("implement me")
}

func (m MockSandBox) GetCoinInfo(coinType coin.Type) (*parameters.IotaCoinInfo, bool) {
	panic("implement me")
}

func (m MockSandBox) CallView(message isc.Message) isc.CallArguments {
	panic("implement me")
}

func (m MockSandBox) StateR() kv.KVStoreReader {
	panic("implement me")
}

func (m MockSandBox) SchemaVersion() isc.SchemaVersion {
	panic("implement me")
}

func (m MockSandBox) AllowanceAvailable() *isc.Assets {
	return nil
}

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts)

func TestEntryPointViewFunc(t *testing.T) {
	testViewFunc := coreutil.NewViewEP11(Contract, "", coreutil.Field[isc.AgentID](""), coreutil.Field[isc.AgentID](""))
	testViewFuncHandler := testViewFunc.WithHandler(func(view isc.SandboxView, id isc.AgentID) isc.AgentID {
		return id
	})
	testAgentIDInput := isctest.NewRandomAgentID()

	mock := MockSandBox{
		MockParams: isc.NewCallArguments(
			codec.Encode(testAgentIDInput)),
	}

	result := testViewFuncHandler.Call(mock)
	require.Len(t, result, 1)

	agentID, err := codec.Decode[isc.AgentID](result[0])
	require.NoError(t, err)
	require.EqualValues(t, testAgentIDInput, agentID)

	// Test auto decoding functionality
	agentID2, err := testViewFunc.DecodeOutput(result)
	require.NoError(t, err)
	require.EqualValues(t, testAgentIDInput, agentID2)
}

func TestEntryPointMutFunc11(t *testing.T) {
	testMutFunc := coreutil.NewEP11(Contract, "", coreutil.Field[uint32](""), coreutil.Field[uint32](""))
	testMutFuncHandler := testMutFunc.WithHandler(func(sandbox isc.Sandbox, u uint32) uint32 {
		return u * 2
	})

	testNumber := uint32(1024)
	mock := MockSandBox{
		MockParams: isc.NewCallArguments(
			codec.Encode(testNumber)),
	}

	result := testMutFuncHandler.Call(mock)
	require.Len(t, result, 1)

	testNumberResult, err := codec.Decode[uint32](result[0])
	require.NoError(t, err)
	require.EqualValues(t, testNumberResult, testNumber*2)

	// Test auto decoding functionality
	uint32Result, err := testMutFunc.DecodeOutput(result)
	require.NoError(t, err)
	require.EqualValues(t, uint32Result, testNumber*2)
}

func TestEntryPointMutFunc12(t *testing.T) {
	testMutFunc := coreutil.NewEP12(Contract, "", coreutil.Field[uint32](""), coreutil.Field[uint32](""), coreutil.Field[uint32](""))
	testMutFuncHandler := testMutFunc.WithHandler(func(sandbox isc.Sandbox, u uint32) (uint32, uint32) {
		return u * 2, u * 3
	})

	testNumber := uint32(1024)
	mock := MockSandBox{
		MockParams: isc.NewCallArguments(
			codec.Encode(testNumber)),
	}

	result := testMutFuncHandler.Call(mock)
	require.Len(t, result, 2)

	testNumberResult1, err := codec.Decode[uint32](result[0])
	require.NoError(t, err)
	require.EqualValues(t, testNumberResult1, testNumber*2)

	testNumberResult2, err := codec.Decode[uint32](result[1])
	require.NoError(t, err)
	require.EqualValues(t, testNumberResult2, testNumber*3)

	// Test auto decoding functionality
	uint32Result2, uint32Result3, err := testMutFunc.DecodeOutput(result)
	require.NoError(t, err)
	require.EqualValues(t, uint32Result2, testNumber*2)
	require.EqualValues(t, uint32Result3, testNumber*3)
}
