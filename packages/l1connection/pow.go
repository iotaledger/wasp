package l1connection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/contextutils"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/pow"
	"github.com/iotaledger/wasp/packages/parameters"
)

type submitBlockFn = func(ctx context.Context, block *iotago.Block) error

const (
	refreshTipsDuringPoWInterval = 5 * time.Second
	parallelWorkers              = 1
)

func doBlockPow(ctx context.Context, block *iotago.Block, useRemotePow bool, submitBlock submitBlockFn, nodeClient *nodeclient.Client) error {
	if useRemotePow {
		// remote PoW: Take the Block, clear parents, clear nonce, send to node
		block.Parents = nil
		block.Nonce = 0
		err := submitBlock(ctx, block)
		return err
	}
	// do the PoW
	refreshTipsFn := func() (tips iotago.BlockIDs, err error) {
		// refresh tips if PoW takes longer than a configured duration.
		resp, err := nodeClient.Tips(ctx)
		if err != nil {
			return nil, err
		}
		return resp.Tips()
	}

	targetScore := float64(parameters.L1().Protocol.MinPoWScore)

	_, err := doPoW(
		ctx,
		block,
		targetScore,
		parallelWorkers,
		refreshTipsDuringPoWInterval,
		refreshTipsFn,
	)

	return err
}

// taken from https://github.com/iotaledger/inx-app/blob/ea47c776549669b81a5105a0ce78c7074694f44e/pow/pow.go
const (
	nonceBytes = 8 // len(uint64)
)

// ErrOperationAborted is returned when the operation was aborted e.g. by a shutdown signal.
var ErrOperationAborted = errors.New("operation was aborted")

// RefreshTipsFunc refreshes tips of the block if PoW takes longer than a configured duration.
type RefreshTipsFunc = func() (tips iotago.BlockIDs, err error)

// DoPoW does the proof-of-work required to hit the given target score.
// The given iota.Block's nonce is automatically updated.
//

//nolint:gocyclo
func doPoW(ctx context.Context, block *iotago.Block, targetScore float64, parallelism int, refreshTipsInterval time.Duration, refreshTipsFunc RefreshTipsFunc) (blockSize int, err error) {
	if targetScore == 0 {
		block.Nonce = 0
		return 0, nil
	}

	if err := contextutils.ReturnErrIfCtxDone(ctx, ErrOperationAborted); err != nil {
		return 0, err
	}

	// enforce milestone block nonce == 0
	if _, isMilestone := block.Payload.(*iotago.Milestone); isMilestone {
		block.Nonce = 0
		return 0, nil
	}

	getPoWData := func(block *iotago.Block) (powData []byte, err error) {
		blockData, err := block.Serialize(serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to perform PoW as block can't be serialized: %w", err)
		}

		return blockData[:len(blockData)-nonceBytes], nil
	}

	powData, err := getPoWData(block)
	if err != nil {
		return 0, err
	}

	doPow := func(ctx context.Context) (uint64, error) {
		powCtx, powCancel := context.WithCancel(ctx)
		defer powCancel()

		if refreshTipsFunc != nil {
			var powTimeoutCancel context.CancelFunc
			powCtx, powTimeoutCancel = context.WithTimeout(powCtx, refreshTipsInterval)
			defer powTimeoutCancel()
		}

		nonce, err := pow.New(parallelism).Mine(powCtx, powData, targetScore)
		if err != nil {
			if errors.Is(err, pow.ErrCancelled) && refreshTipsFunc != nil {
				// context was canceled and tips can be refreshed
				tips, err := refreshTipsFunc()
				if err != nil {
					return 0, err
				}
				block.Parents = tips

				// replace the powData to update the new tips
				powData, err = getPoWData(block)
				if err != nil {
					return 0, err
				}

				return 0, pow.ErrCancelled
			}
			return 0, err
		}

		return nonce, nil
	}

	for {
		nonce, err := doPow(ctx)
		if err != nil {
			// check if the external context got canceled.
			if ctx.Err() != nil {
				return 0, ErrOperationAborted
			}

			if errors.Is(err, pow.ErrCancelled) {
				// redo the PoW with new tips
				continue
			}
			return 0, err
		}

		block.Nonce = nonce
		return len(powData) + nonceBytes, nil
	}
}
