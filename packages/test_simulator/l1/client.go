package l1

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
)

type FakeL1Client struct {
	iotagraphql.IotaClient

	iotaClient *FakeIotaClient
}

var _ clients.L1Client = (*FakeL1Client)(nil)

type Option func(*FakeL1Client)

func NewFakeL1Client(moveHandler MoveCallHandler, opts ...Option) *FakeL1Client {
	store := NewObjectStore()
	executor := NewExecutor(store, moveHandler)
	iotaClient := NewFakeIotaClient(store, executor)

	c := &FakeL1Client{
		IotaClient: iotaClient,
		iotaClient: iotaClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *FakeL1Client) Health(_ context.Context) error {
	return nil
}

func (c *FakeL1Client) L2() clients.L2Client {
	return iscmoveclient.NewClient(c.GetIotaClient())
}

func (c *FakeL1Client) GetIotaClient() iotagraphql.IotaClient {
	return c.iotaClient
}

// WaitForNextVersionForTesting is synchronous in the simulator
func (c *FakeL1Client) WaitForNextVersionForTesting(
	ctx context.Context,
	_ time.Duration,
	_ log.Logger,
	currentRef *iotago.ObjectRef,
	cb func(),
) (*iotago.ObjectRef, error) {
	if currentRef == nil {
		cb()
		return currentRef, nil
	}

	cb()

	return c.iotaClient.UpdateObjectRef(ctx, currentRef)
}

func (c *FakeL1Client) Store() *ObjectStore {
	return c.iotaClient.Store
}

func (c *FakeL1Client) UpdateMoveHandler(handler MoveCallHandler) {
	c.iotaClient.Executor.MoveHandler = handler
}

func WithPresetBalance(addr iotago.Address, amount uint64) Option {
	return func(c *FakeL1Client) {
		var counter uint64
		txDigest := ComputeDigest(addr[:])
		coinID := FreshID(txDigest, &counter)
		c.iotaClient.Store.PresetCoinObject(coinID, addr, IotaCoinTypeStr, amount, txDigest)
	}
}
