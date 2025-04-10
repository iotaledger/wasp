package base

import (
	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"
	"github.com/stretchr/testify/require"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
)

type BlockLogClientWrapper struct {
	ValidationContext
}

// BlocklogGetBlockInfo wraps both API calls for getting block info
func (c *BlockLogClientWrapper) BlocklogGetBlockInfo(blockIndex uint32) (*stardust_client.BlockInfoResponse, *rebased_client.BlockInfoResponse) {
	blockStr := Uint32ToString(blockIndex)

	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetBlockInfo(c.Ctx, MainnetChainID, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetBlockInfo(c.Ctx, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetControlAddresses wraps both API calls for getting control addresses
func (c *BlockLogClientWrapper) BlocklogGetControlAddresses() (*stardust_client.ControlAddressesResponse, *rebased_client.ControlAddressesResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetControlAddresses(c.Ctx, MainnetChainID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetControlAddresses(c.Ctx).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetEventsOfBlock wraps both API calls for getting events of a block
func (c *BlockLogClientWrapper) BlocklogGetEventsOfBlock(blockIndex uint32) (*stardust_client.EventsResponse, *rebased_client.EventsResponse) {
	blockStr := Uint32ToString(blockIndex)

	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetEventsOfBlock(c.Ctx, MainnetChainID, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetEventsOfBlock(c.Ctx, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetEventsOfLatestBlock wraps both API calls for getting events of the latest block
func (c *BlockLogClientWrapper) BlocklogGetEventsOfLatestBlock() (*stardust_client.EventsResponse, *rebased_client.EventsResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetEventsOfLatestBlock(c.Ctx, MainnetChainID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetEventsOfLatestBlock(c.Ctx).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetEventsOfRequest wraps both API calls for getting events of a request
func (c *BlockLogClientWrapper) BlocklogGetEventsOfRequest(requestID string) (*stardust_client.EventsResponse, *rebased_client.EventsResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetEventsOfRequest(c.Ctx, MainnetChainID, requestID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetEventsOfRequest(c.Ctx, requestID).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetLatestBlockInfo wraps both API calls for getting latest block info
func (c *BlockLogClientWrapper) BlocklogGetLatestBlockInfo() (*stardust_client.BlockInfoResponse, *rebased_client.BlockInfoResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetLatestBlockInfo(c.Ctx, MainnetChainID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetLatestBlockInfo(c.Ctx).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetRequestIDsForBlock wraps both API calls for getting request IDs for a block
func (c *BlockLogClientWrapper) BlocklogGetRequestIDsForBlock(blockIndex uint32) (*stardust_client.RequestIDsResponse, *rebased_client.RequestIDsResponse) {
	blockStr := Uint32ToString(blockIndex)

	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetRequestIDsForBlock(c.Ctx, MainnetChainID, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetRequestIDsForBlock(c.Ctx, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetRequestIDsForLatestBlock wraps both API calls for getting request IDs for the latest block
func (c *BlockLogClientWrapper) BlocklogGetRequestIDsForLatestBlock() (*stardust_client.RequestIDsResponse, *rebased_client.RequestIDsResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetRequestIDsForLatestBlock(c.Ctx, MainnetChainID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetRequestIDsForLatestBlock(c.Ctx).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetRequestIsProcessed wraps both API calls for checking if a request is processed
func (c *BlockLogClientWrapper) BlocklogGetRequestIsProcessed(requestID string) (*stardust_client.RequestProcessedResponse, *rebased_client.RequestProcessedResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetRequestIsProcessed(c.Ctx, MainnetChainID, requestID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetRequestIsProcessed(c.Ctx, requestID).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetRequestReceipt wraps both API calls for getting a request receipt
func (c *BlockLogClientWrapper) BlocklogGetRequestReceipt(requestID string) (*stardust_client.ReceiptResponse, *rebased_client.ReceiptResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetRequestReceipt(c.Ctx, MainnetChainID, requestID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetRequestReceipt(c.Ctx, requestID).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetRequestReceiptsOfBlock wraps both API calls for getting request receipts of a block
func (c *BlockLogClientWrapper) BlocklogGetRequestReceiptsOfBlock(blockIndex uint32) ([]stardust_client.ReceiptResponse, []rebased_client.ReceiptResponse) {
	blockStr := Uint32ToString(blockIndex)

	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetRequestReceiptsOfBlock(c.Ctx, MainnetChainID, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetRequestReceiptsOfBlock(c.Ctx, blockIndex).Block(blockStr).Execute()
	require.NoError(T, err)

	return sRes, rRes
}

// BlocklogGetRequestReceiptsOfLatestBlock wraps both API calls for getting request receipts of the latest block
func (c *BlockLogClientWrapper) BlocklogGetRequestReceiptsOfLatestBlock() ([]stardust_client.ReceiptResponse, []rebased_client.ReceiptResponse) {
	sRes, _, err := c.SClient.CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock(c.Ctx, MainnetChainID).Execute()
	require.NoError(T, err)

	rRes, _, err := c.RClient.CorecontractsAPI.BlocklogGetRequestReceiptsOfLatestBlock(c.Ctx).Execute()
	require.NoError(T, err)

	return sRes, rRes
}