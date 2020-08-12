package sandbox

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type MockedSandbox struct {
	kv kv.KVStore
}

func NewMockedSandbox() vmtypes.Sandbox {
	return &MockedSandbox{
		kv: kv.NewMap(),
	}
}

func (m *MockedSandbox) IsOriginState() bool {
	panic("implement me")
}

func (m *MockedSandbox) GetSCAddress() *address.Address {
	panic("implement me")
}

func (m *MockedSandbox) GetOwnerAddress() *address.Address {
	panic("implement me")
}

func (m *MockedSandbox) GetTimestamp() int64 {
	panic("implement me")
}

func (m *MockedSandbox) GetEntropy() hashing.HashValue {
	panic("implement me")
}

func (m *MockedSandbox) Panic(v interface{}) {
	panic("implement me")
}

func (m *MockedSandbox) Rollback() {
	panic("implement me")
}

func (m *MockedSandbox) AccessRequest() vmtypes.RequestAccess {
	panic("implement me")
}

func (m *MockedSandbox) AccessState() kv.MustCodec {
	return kv.NewMustCodec(m.kv)
}

func (m *MockedSandbox) AccessSCAccount() vmtypes.AccountAccess {
	panic("implement me")
}

func (m *MockedSandbox) SendRequest(par vmtypes.NewRequestParams) bool {
	panic("implement me")
}

func (m *MockedSandbox) SendRequestToSelf(reqCode sctransaction.RequestCode, args kv.Map) bool {
	panic("implement me")
}

func (m *MockedSandbox) SendRequestToSelfWithDelay(reqCode sctransaction.RequestCode, args kv.Map, deferForSec uint32) bool {
	panic("implement me")
}

func (m *MockedSandbox) Publish(msg string) {
	fmt.Printf("MockedSandbox.Publish: %s\n", msg)
}

func (m *MockedSandbox) Publishf(format string, args ...interface{}) {
	fmt.Printf("MockedSandbox.Publish: "+format+"\n", args...)
}

func (m *MockedSandbox) GetWaspLog() *logger.Logger {
	panic("implement me")
}

func (m *MockedSandbox) DumpAccount() string {
	panic("implement me")
}
