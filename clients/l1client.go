package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
)

type L1Config struct {
	APIURL    string
	FaucetURL string
}

type L1Client interface {
	iotagraphql.IotaClient

	Health(ctx context.Context) error
	L2() L2Client
	GetIotaClient() iotagraphql.IotaClient
	WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error)
}

var _ L1Client = &l1Client{}

type l1Client struct {
	iotagraphql.IotaClient

	Config L1Config
}

func (c *l1Client) Health(ctx context.Context) error {
	_, err := c.GetLatestIotaSystemState(ctx)
	return err
}

func (c *l1Client) L2() L2Client {
	return iscmoveclient.NewClient(c.GetIotaClient())
}

func (c *l1Client) GetIotaClient() iotagraphql.IotaClient {
	return c
}

// WaitForNextVersionForTesting waits for an object to change its version.
// This tries to make sure that an object meant to be used multiple times, does not get referenced twice with the same ref.
// Handle with care. Only use it on objects that are expected to be used again, like a GasCoin/Generic coin/Requests
func (c *l1Client) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	// Some 'sugar' to make dynamic refs handling easier (where refs can be nil or set depending on state)
	if currentRef == nil {
		cb()
		return currentRef, nil
	}

	cb()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("WaitForNextVersionForTesting: context deadline exceeded while waiting for object version change: %v", currentRef)
		case <-ticker.C:
			// Poll for object update
			resp, err := c.GetObject(ctx, *currentRef.ObjectID)
			if err != nil {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: error getting object: %v, retrying...", err)
				}
				continue
			}

			if resp.Object.IsNotFound() || resp.Object.IsDeleted() {
				// The provided object got consumed and is gone. We can return.
				return currentRef, nil
			}

			ref, err := resp.Object.ObjectRef()
			if err != nil {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: error parsing object ref: %v, retrying...", err)
				}
				continue
			}

			if ref.Version > currentRef.Version {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: Found the updated version of %v, which is: %v", currentRef, ref)
				}
				return ref, nil
			}

			if logger != nil {
				logger.LogInfof("WaitForNextVersionForTesting: Getting the same version ref as before. Retrying. %v", currentRef)
			}
		}
	}
}

func NewL1Client(l1Config L1Config, waitUntilEffectsVisible *iotagraphql.WaitParams) L1Client {
	return &l1Client{
		IotaClient: iotagraphql.NewGraphQLClientWithWaitParams(l1Config.APIURL, l1Config.FaucetURL, waitUntilEffectsVisible),
		Config:     l1Config,
	}
}

// NewL1ClientFromIotaClient wraps an existing IotaClient as an L1Client.
func NewL1ClientFromIotaClient(iotaClient iotagraphql.IotaClient) L1Client {
	return &l1Client{
		IotaClient: iotaClient,
	}
}

func NewLocalnetClient(waitUntilEffectsVisible *iotagraphql.WaitParams) L1Client {
	return NewL1Client(L1Config{
		APIURL:    iotaconn.LocalnetEndpointURL,
		FaucetURL: iotaconn.LocalnetFaucetURL,
	}, waitUntilEffectsVisible)
}
