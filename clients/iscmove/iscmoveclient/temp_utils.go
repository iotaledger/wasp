// Package iscmoveclient implements client functionality for ISC move operations.
package iscmoveclient

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
)

// TODO this is a 1:1 copy of l1client.WaitForNextVersionForTesting
// As we separate both libraries, we can't simply reference this function
// For now it's placed here, but should be removed soon.

func (c *Client) MustWaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) *iotago.ObjectRef {
	return lo.Must(c.WaitForNextVersionForTesting(ctx, timeout, logger, currentRef, cb))
}

func (c *Client) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
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
			newRef, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: currentRef.ObjectID})
			if err != nil {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: error getting object: %v, retrying...", err)
				} else {
					fmt.Printf("WaitForNextVersionForTesting: error getting object: %v, retrying...", err)
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
				} else {
					fmt.Printf("WaitForNextVersionForTesting: object error: %v, retrying...", newRef.Error)
				}
				continue
			}

			if newRef.Data.Ref().Version > currentRef.Version {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: Found the updated version of %v, which is: %v", currentRef, newRef.Data.Ref())
				} else {
					fmt.Printf("WaitForNextVersionForTesting: Found the updated version of %v, which is: %v", currentRef, newRef.Data.Ref())
				}

				ref := newRef.Data.Ref()
				return &ref, nil
			}

			if logger != nil {
				logger.LogInfof("WaitForNextVersionForTesting: Getting the same version ref as before. Retrying. %v", currentRef)
			} else {
				fmt.Printf("WaitForNextVersionForTesting: Getting the same version ref as before. Retrying. %v", currentRef)
			}
		}
	}
}
