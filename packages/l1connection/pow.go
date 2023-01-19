package l1connection

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	inxpow "github.com/iotaledger/inx-app/pkg/pow"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/parameters"
)

const (
	refreshTipsInterval = 5 * time.Second
	parallelism         = 1
)

// doBlockPow will update the parents during PoW and add initial parents if no parents were given.
func doBlockPow(ctx context.Context, block *iotago.Block, nodeClient *nodeclient.Client) error {
	var refreshTipsFunc inxpow.RefreshTipsFunc

	// only allow to update tips during proof of work if no parents were given
	if len(block.Parents) == 0 {
		refreshTipsFunc = func() (tips iotago.BlockIDs, err error) {
			// refresh tips if PoW takes longer than a configured duration.
			resp, err := nodeClient.Tips(ctx)
			if err != nil {
				return nil, err
			}
			return resp.Tips()
		}
	}

	// DoPoW does the proof-of-work required to hit the target score configured on this Handler.
	// The given iota.Block's nonce is automatically updated.
	_, err := inxpow.DoPoW(ctx, block, serializer.DeSeriModePerformValidation|serializer.DeSeriModePerformLexicalOrdering, parameters.L1().Protocol, parallelism, refreshTipsInterval, refreshTipsFunc)
	return err
}
