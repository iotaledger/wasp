package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) GetChainIdentifier(ctx context.Context) (string, error) {
	var resp string
	return resp, c.transport.Call(ctx, &resp, getChainIdentifier)
}

func (c *Client) GetCheckpoint(ctx context.Context, checkpointId *suijsonrpc.BigInt) (*suijsonrpc.Checkpoint, error) {
	var resp suijsonrpc.Checkpoint
	return &resp, c.transport.Call(ctx, &resp, getCheckpoint, checkpointId)
}

type GetCheckpointsRequest struct {
	Cursor          *suijsonrpc.BigInt // optional
	Limit           *uint64            // optional
	DescendingOrder bool
}

func (c *Client) GetCheckpoints(ctx context.Context, req GetCheckpointsRequest) (*suijsonrpc.CheckpointPage, error) {
	var resp suijsonrpc.CheckpointPage
	return &resp, c.transport.Call(ctx, &resp, getCheckpoints, req.Cursor, req.Limit, req.DescendingOrder)
}

func (c *Client) GetEvents(ctx context.Context, digest *sui.TransactionDigest) ([]*suijsonrpc.SuiEvent, error) {
	var resp []*suijsonrpc.SuiEvent
	return resp, c.transport.Call(ctx, &resp, getEvents, digest)
}

func (c *Client) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	var resp string
	return resp, c.transport.Call(ctx, &resp, getLatestCheckpointSequenceNumber)
}

// TODO getLoadedChildObjects

type GetObjectRequest struct {
	ObjectID *sui.ObjectID
	Options  *suijsonrpc.SuiObjectDataOptions // optional
}

func (c *Client) GetObject(ctx context.Context, req GetObjectRequest) (*suijsonrpc.SuiObjectResponse, error) {
	var resp suijsonrpc.SuiObjectResponse
	return &resp, c.transport.Call(ctx, &resp, getObject, req.ObjectID, req.Options)
}

func (c *Client) GetProtocolConfig(
	ctx context.Context,
	version *suijsonrpc.BigInt, // optional
) (*suijsonrpc.ProtocolConfig, error) {
	var resp suijsonrpc.ProtocolConfig
	return &resp, c.transport.Call(ctx, &resp, getProtocolConfig, version)
}

func (c *Client) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	var resp string
	return resp, c.transport.Call(ctx, &resp, getTotalTransactionBlocks)
}

type GetTransactionBlockRequest struct {
	Digest  *sui.TransactionDigest
	Options *suijsonrpc.SuiTransactionBlockResponseOptions // optional
}

func (c *Client) GetTransactionBlock(ctx context.Context, req GetTransactionBlockRequest) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	resp := suijsonrpc.SuiTransactionBlockResponse{}
	return &resp, c.transport.Call(ctx, &resp, getTransactionBlock, req.Digest, req.Options)
}

type MultiGetObjectsRequest struct {
	ObjectIDs []*sui.ObjectID
	Options   *suijsonrpc.SuiObjectDataOptions // optional
}

func (c *Client) MultiGetObjects(ctx context.Context, req MultiGetObjectsRequest) ([]suijsonrpc.SuiObjectResponse, error) {
	var resp []suijsonrpc.SuiObjectResponse
	return resp, c.transport.Call(ctx, &resp, multiGetObjects, req.ObjectIDs, req.Options)
}

type MultiGetTransactionBlocksRequest struct {
	Digests []*sui.Digest
	Options *suijsonrpc.SuiTransactionBlockResponseOptions // optional
}

func (c *Client) MultiGetTransactionBlocks(
	ctx context.Context,
	req MultiGetTransactionBlocksRequest,
) ([]*suijsonrpc.SuiTransactionBlockResponse, error) {
	resp := []*suijsonrpc.SuiTransactionBlockResponse{}
	return resp, c.transport.Call(ctx, &resp, multiGetTransactionBlocks, req.Digests, req.Options)
}

type TryGetPastObjectRequest struct {
	ObjectID *sui.ObjectID
	Version  uint64
	Options  *suijsonrpc.SuiObjectDataOptions // optional
}

func (c *Client) TryGetPastObject(
	ctx context.Context,
	req TryGetPastObjectRequest,
) (*suijsonrpc.SuiPastObjectResponse, error) {
	var resp suijsonrpc.SuiPastObjectResponse
	return &resp, c.transport.Call(ctx, &resp, tryGetPastObject, req.ObjectID, req.Version, req.Options)
}

type TryMultiGetPastObjectsRequest struct {
	PastObjects []*suijsonrpc.SuiGetPastObjectRequest
	Options     *suijsonrpc.SuiObjectDataOptions // optional
}

func (c *Client) TryMultiGetPastObjects(
	ctx context.Context,
	req TryMultiGetPastObjectsRequest,
) ([]*suijsonrpc.SuiPastObjectResponse, error) {
	var resp []*suijsonrpc.SuiPastObjectResponse
	return resp, c.transport.Call(ctx, &resp, tryMultiGetPastObjects, req.PastObjects, req.Options)
}
