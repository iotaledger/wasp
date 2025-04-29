package base

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type evmEp string

const (
	stardust evmEp = "stardust"
	rebased  evmEp = "rebased"
)

type EvmClient struct {
	S_client *ethclient.Client
	R_client *ethclient.Client
}

func NewEvmClient(s_clientUri string, r_clientUri string) (*EvmClient, error) {
	s_rpc, err := rpc.Dial(s_clientUri)
	if err != nil {
		return nil, err
	}
	s_client := ethclient.NewClient(s_rpc)

	r_rpc, err := rpc.Dial(r_clientUri)
	if err != nil {
		return nil, err
	}
	r_client := ethclient.NewClient(r_rpc)

	return &EvmClient{
		S_client: s_client,
		R_client: r_client,
	}, nil
}

// func (c *EvmClient) BlockNumber(ctx context.Context) (uint64, error) {
// 	return c.R_client.BlockNumber(ctx)
// }

func (c *EvmClient) getClient(ep evmEp) *ethclient.Client {
	if ep == stardust {
		return c.S_client
	}
	return c.R_client
}

func (c *EvmClient) GetBlockByNumber(ctx context.Context, blockNumber uint64, ep evmEp) (*types.Block, error) {
	return c.getClient(ep).BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
}

func (c *EvmClient) TraceTransaction(ctx context.Context, txHash common.Hash, ep evmEp) (interface{}, error) {
	var result interface{}
	err := c.getClient(ep).Client().CallContext(ctx, &result, "debug_traceTransaction", txHash, map[string]interface{}{
		"tracer": "callTracer",
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type TxTrace struct {
	TxHash common.Hash
	Trace  any
}

func (c *EvmClient) GetBlockTraces(ctx context.Context, blockNumber uint64, ep evmEp) ([]TxTrace, error) {
	block, err := c.GetBlockByNumber(ctx, blockNumber, ep)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	traces := make([]TxTrace, 0)
	for _, tx := range block.Transactions() {
		traceResult, err := c.TraceTransaction(ctx, tx.Hash(), ep)
		if err != nil {
			return nil, fmt.Errorf("failed to trace transaction %s in block %s: %w", tx.Hash().Hex(), hexutil.EncodeUint64(blockNumber), err)
		}

		traces = append(traces, TxTrace{
			TxHash: tx.Hash(),
			Trace:  traceResult,
		})
	}

	return traces, nil
}

func (c *EvmClient) GetBalance(ctx context.Context, address common.Address, blockNumber uint64, ep evmEp) (*big.Int, error) {
	return c.getClient(ep).BalanceAt(ctx, address, big.NewInt(int64(blockNumber)))
}
