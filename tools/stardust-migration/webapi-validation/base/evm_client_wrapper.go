package base

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type EvmClientWrapper struct {
	ValidationContext
	EvmClient
}

func NewEvmClientWrapper(s_clientUri string, r_clientUri string, ctx ValidationContext) (*EvmClientWrapper, error) {
	evmClient, err := NewEvmClient(s_clientUri, r_clientUri)
	if err != nil {
		return nil, err
	}
	return &EvmClientWrapper{
		ValidationContext: ctx,
		EvmClient:         *evmClient,
	}, nil
}

func (c *EvmClientWrapper) GetTxTraces(blockNumber uint64) ([]TxTrace, []TxTrace, error) {

	stardustTraces := make([]TxTrace, 0)
	rebasedTraces := make([]TxTrace, 0)
	var sErr error
	var rErr error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		stardustTraces, sErr = c.GetBlockTraces(c.Ctx, blockNumber, stardust)
	}()

	go func() {
		defer wg.Done()
		rebasedTraces, rErr = c.GetBlockTraces(c.Ctx, blockNumber, rebased)
	}()

	wg.Wait()

	if sErr != nil {
		return nil, nil, fmt.Errorf("failed to get stardust traces for block %d: %w", blockNumber, sErr)
	}

	if rErr != nil {
		return nil, nil, fmt.Errorf("failed to get rebased traces for block %d: %w", blockNumber, rErr)
	}

	return stardustTraces, rebasedTraces, nil
}

func (c *EvmClientWrapper) GetBalances(address common.Address, blockNumber uint64) (stardustBalance *big.Int, rebasedBalance *big.Int, err error) {
	stardustBalance, err = c.GetBalance(c.Ctx, address, blockNumber, stardust)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stardust balance for block %d: %w", blockNumber, err)
	}
	rebasedBalance, err = c.GetBalance(c.Ctx, address, blockNumber, rebased)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rebased balance for block %d: %w", blockNumber, err)
	}
	return stardustBalance, rebasedBalance, nil
}
