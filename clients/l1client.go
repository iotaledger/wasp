package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type L1Config struct {
	APIURL    string
	FaucetURL string
}

type L1Client interface {
	iotagraphql.IotaClient

	RequestFunds(ctx context.Context, address cryptolib.Address) error
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

func (c *l1Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	faucetURL := c.Config.FaucetURL
	if faucetURL == "" {
		faucetURL = iotaconn.FaucetURL(c.Config.APIURL)
	}
	return iotaclient.RequestFundsFromFaucet(ctx, address.AsIotaAddress(), faucetURL)
}

func (c *l1Client) Health(ctx context.Context) error {
	_, err := c.GetLatestIotaSystemState(ctx)
	return err
}

func (c *l1Client) L2() L2Client {
	return iscmoveclient.NewClient(c.GetIotaClient(), c.Config.FaucetURL)
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

	// Create a ticker for polling
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	// Add timeout to context if not already set
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("WaitForNextVersionForTesting: context deadline exceeded while waiting for object version change: %v", currentRef)
		case <-ticker.C:
			// Poll for object update
			newRef, err := c.GetObject(ctx, iotagraphql.GetObjectRequest{ObjectID: currentRef.ObjectID})
			if err != nil {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: error getting object: %v, retrying...", err)
				}
				continue
			}

			if newRef.Error != nil {
				// The provided object got consumed and is gone. We can return.
				if newRef.Error.Data.Deleted != nil || newRef.Error.Data.NotExists != nil {
					return currentRef, nil
				}

				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: object error: %v, retrying...", newRef.Error)
				}
				continue
			}

			if newRef.Data.Ref().Version > currentRef.Version {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: Found the updated version of %v, which is: %v", currentRef, newRef.Data.Ref())
				}

				ref := newRef.Data.Ref()
				return &ref, nil
			}

			if logger != nil {
				logger.LogInfof("WaitForNextVersionForTesting: Getting the same version ref as before. Retrying. %v", currentRef)
			}
		}
	}
}

func NewL1Client(l1Config L1Config, waitUntilEffectsVisible *iotagraphql.WaitParams) L1Client {
	return &l1Client{
		IotaClient: iotagraphql.NewGraphQLClientWithWaitParams(l1Config.APIURL, waitUntilEffectsVisible),
		Config:     l1Config,
	}
}

func NewLocalnetClient(waitUntilEffectsVisible *iotagraphql.WaitParams) L1Client {
	return NewL1Client(L1Config{
		APIURL:    iotaconn.LocalnetEndpointURL,
		FaucetURL: iotaconn.LocalnetFaucetURL,
	}, waitUntilEffectsVisible)
}
