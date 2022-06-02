package nodeconn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/contextutils"
	serializer "github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/pow"
	"github.com/iotaledger/wasp/packages/parameters"
)

// taken from https://github.com/iotaledger/inx-spammer/blob/develop/pkg/pow/pow.go

const nonceBytes = 8 // len(uint64)

var ErrOperationAborted = errors.New("operation was aborted")

// RefreshTipsFunc refreshes tips of the block if PoW takes longer than a configured duration.
type RefreshTipsFunc = func() (tips iotago.BlockIDs, err error)

// doPoW does the proof-of-work required to hit the given target score.
// The given iota.Block's nonce is automatically updated.
func doPoW(ctx context.Context, block *iotago.Block, refreshTipsInterval time.Duration, refreshTipsFunc RefreshTipsFunc) error {
	if err := contextutils.ReturnErrIfCtxDone(ctx, ErrOperationAborted); err != nil {
		return err
	}
	targetScore := float64(parameters.L1.Protocol.MinPoWScore)

	getPoWData := func(block *iotago.Block) (powData []byte, err error) {
		blockData, err := block.Serialize(serializer.DeSeriModePerformValidation, parameters.L1.Protocol)
		if err != nil {
			return nil, fmt.Errorf("unable to perform PoW as block can't be serialized: %w", err)
		}

		return blockData[:len(blockData)-serializer.UInt64ByteSize], nil
	}

	powData, err := getPoWData(block)
	if err != nil {
		return err
	}

	doPow := func(ctx context.Context) (uint64, error) {
		powCtx, powCancel := context.WithCancel(ctx)
		defer powCancel()

		if refreshTipsFunc != nil {
			var powTimeoutCancel context.CancelFunc
			powCtx, powTimeoutCancel = context.WithTimeout(powCtx, refreshTipsInterval)
			defer powTimeoutCancel()
		}

		nonce, err := pow.New().Mine(powCtx, powData, targetScore)
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
				return ErrOperationAborted
			}

			if errors.Is(err, pow.ErrCancelled) {
				// redo the PoW with new tips
				continue
			}
			return err
		}

		block.Nonce = nonce
		return nil
	}
}
