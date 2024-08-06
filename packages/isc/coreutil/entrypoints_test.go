package coreutil

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type MockSandBox struct {
	MockParams isc.CallArguments
}

func (m MockSandBox) Requiref(cond bool, format string, args ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) RequireNoError(err error, str ...string) {
	if err != nil {
		panic(str)
	}
}

func (m MockSandBox) BalanceBaseTokens() (bts uint64, remainder *big.Int) {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) BalanceNativeToken(id iotago.NativeTokenID) *big.Int {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) BalanceNativeTokens() iotago.NativeTokens {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) OwnedNFTs() []iotago.NFTID {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) HasInAccount(id isc.AgentID, assets *isc.Assets) bool {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Params() *isc.Params {
	return &isc.Params{
		Args: m.MockParams,
	}
}

func (m MockSandBox) ChainID() isc.ChainID {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) ChainOwnerID() isc.AgentID {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) ChainInfo() *isc.ChainInfo {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Contract() isc.Hname {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) AccountID() isc.AgentID {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Caller() isc.AgentID {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Timestamp() time.Time {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Log() isc.LogInterface {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Utils() isc.Utils {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) Gas() isc.Gas {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) GetNFTData(nftID iotago.NFTID) *isc.NFT {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) CallView(message isc.Message) dict.Dict {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) StateR() kv.KVStoreReader {
	//TODO implement me
	panic("implement me")
}

func (m MockSandBox) SchemaVersion() isc.SchemaVersion {
	//TODO implement me
	panic("implement me")
}

var TestAgentIDInput = isc.NewRandomAgentID()
var Contract = NewContract(CoreContractAccounts)

var TestFunc = NewViewEP11(Contract, "getEVMGasRatio", FieldWithCodec("", codec.AgentID), FieldWithCodec("", codec.AgentID)).
	WithHandler(func(view isc.SandboxView, id isc.AgentID) isc.AgentID {
		return id
	})

func TestTestFunc(t *testing.T) {
	mock := MockSandBox{
		MockParams: isc.NewCallArguments(
			codec.AgentID.Encode(TestAgentIDInput)),
	}

	result := TestFunc.Call(mock)
	require.Len(t, result, 1)

	agentID, err := codec.AgentID.Decode(result[0])
	require.NoError(t, err)
	require.EqualValues(t, TestAgentIDInput, agentID)
}
