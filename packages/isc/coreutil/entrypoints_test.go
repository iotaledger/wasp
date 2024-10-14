package coreutil

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var _ isc.SandboxBase = MockSandBox{}

type MockSandBox struct {
	MockParams isc.CallArguments
}

func (m MockSandBox) RequireCaller(agentID isc.AgentID) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) RequireCallerAnyOf(agentID []isc.AgentID) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) RequireCallerIsChainOwner() {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) State() kv.KVStore {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Request() isc.Calldata {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Call(msg isc.Message, allowance *isc.Assets) dict.Dict {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Event(topic string, payload []byte) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) RegisterError(messageFormat string) *isc.VMErrorTemplate {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) GetEntropy() hashing.HashValue {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) TransferAllowedFunds(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Send(metadata isc.RequestParameters) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) EstimateRequiredStorageDeposit(r isc.RequestParameters) uint64 {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) StateAnchor() *isc.StateAnchor {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) RequestIndex() uint16 {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) EVMTracer() *isc.EVMTracer {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) TakeStateSnapshot() int {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) RevertToStateSnapshot(i int) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Privileged() isc.Privileged {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Requiref(cond bool, format string, args ...interface{}) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) RequireNoError(err error, str ...string) {
	if err != nil {
		panic(str)
	}
}

func (m MockSandBox) BaseTokensBalance() (bts coin.Value, remainder *big.Int) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) CoinBalance(coinType coin.Type) coin.Value {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) CoinBalances() isc.CoinBalances {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) OwnedObjects() []iotago.ObjectID {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) HasInAccount(id isc.AgentID, assets *isc.Assets) bool {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Params() isc.CallArguments {
	return m.MockParams
}

func (m MockSandBox) ChainID() isc.ChainID {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) ChainOwnerID() isc.AgentID {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) ChainInfo() *isc.ChainInfo {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Contract() isc.Hname {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) AccountID() isc.AgentID {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Caller() isc.AgentID {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Timestamp() time.Time {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Log() isc.LogInterface {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Utils() isc.Utils {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) Gas() isc.Gas {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) GetObjectBCS(id iotago.ObjectID) ([]byte, bool) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) GetCoinInfo(coinType coin.Type) (*isc.SuiCoinInfo, bool) {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) CallView(message isc.Message) isc.CallArguments {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) StateR() kv.KVStoreReader {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) SchemaVersion() isc.SchemaVersion {
	// TODO implement me
	panic("implement me")
}

func (m MockSandBox) AllowanceAvailable() *isc.Assets {
	return nil
}

var Contract = NewContract(CoreContractAccounts)

func TestEntryPointViewFunc(t *testing.T) {
	testViewFunc := NewViewEP11(Contract, "", Field[isc.AgentID](), Field[isc.AgentID]())
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
	testMutFunc := NewEP11(Contract, "", Field[uint32](), Field[uint32]())
	testMutFuncHandler := testMutFunc.WithHandler(func(sandbox isc.Sandbox, u uint32) uint32 {
		return u * 2
	})

	testNumber := uint32(1024)
	mock := MockSandBox{
		MockParams: isc.NewCallArguments(
			codec.Encode[uint32](testNumber)),
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
	testMutFunc := NewEP12(Contract, "", Field[uint32](), Field[uint32](), Field[uint32]())
	testMutFuncHandler := testMutFunc.WithHandler(func(sandbox isc.Sandbox, u uint32) (uint32, uint32) {
		return u * 2, u * 3
	})

	testNumber := uint32(1024)
	mock := MockSandBox{
		MockParams: isc.NewCallArguments(
			codec.Encode[uint32](testNumber)),
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
