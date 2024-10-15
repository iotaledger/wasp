package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
)

func (c *Client) GetChainIdentifier(ctx context.Context) (string, error) {
	var resp string
	return resp, c.transport.Call(ctx, &resp, getChainIdentifier)
}

func (c *Client) GetCheckpoint(ctx context.Context, checkpointId *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error) {
	var resp iotajsonrpc.Checkpoint
	return &resp, c.transport.Call(ctx, &resp, getCheckpoint, checkpointId)
}

type GetCheckpointsRequest struct {
	Cursor          *iotajsonrpc.BigInt // optional
	Limit           *uint64             // optional
	DescendingOrder bool
}

func (c *Client) GetCheckpoints(ctx context.Context, req GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error) {
	var resp iotajsonrpc.CheckpointPage
	return &resp, c.transport.Call(ctx, &resp, getCheckpoints, req.Cursor, req.Limit, req.DescendingOrder)
}

func (c *Client) GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error) {
	var resp []*iotajsonrpc.IotaEvent
	return resp, c.transport.Call(ctx, &resp, getEvents, digest)
}

func (c *Client) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	var resp string
	return resp, c.transport.Call(ctx, &resp, getLatestCheckpointSequenceNumber)
}

// TODO getLoadedChildObjects

type GetObjectRequest struct {
	ObjectID *iotago.ObjectID
	Options  *iotajsonrpc.IotaObjectDataOptions // optional
}

func (c *Client) GetObject(ctx context.Context, req GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	var resp iotajsonrpc.IotaObjectResponse
	return &resp, c.transport.Call(ctx, &resp, getObject, req.ObjectID, req.Options)
}

func (c *Client) GetProtocolConfig(
	ctx context.Context,
	version *iotajsonrpc.BigInt, // optional
) (*iotajsonrpc.ProtocolConfig, error) {
	var resp iotajsonrpc.ProtocolConfig
	return &resp, c.transport.Call(ctx, &resp, getProtocolConfig, version)
}

func (c *Client) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	var resp string
	return resp, c.transport.Call(ctx, &resp, getTotalTransactionBlocks)
}

type GetTransactionBlockRequest struct {
	Digest  *iotago.TransactionDigest
	Options *iotajsonrpc.IotaTransactionBlockResponseOptions // optional
}

func (c *Client) GetTransactionBlock(
	ctx context.Context,
	req GetTransactionBlockRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	resp := iotajsonrpc.IotaTransactionBlockResponse{}
	return &resp, c.transport.Call(ctx, &resp, getTransactionBlock, req.Digest, req.Options)
}

type MultiGetObjectsRequest struct {
	ObjectIDs []*iotago.ObjectID
	Options   *iotajsonrpc.IotaObjectDataOptions // optional
}

func (c *Client) MultiGetObjects(ctx context.Context, req MultiGetObjectsRequest) (
	[]iotajsonrpc.IotaObjectResponse,
	error,
) {
	var resp []iotajsonrpc.IotaObjectResponse
	return resp, c.transport.Call(ctx, &resp, multiGetObjects, req.ObjectIDs, req.Options)
}

type MultiGetTransactionBlocksRequest struct {
	Digests []*iotago.Digest
	Options *iotajsonrpc.IotaTransactionBlockResponseOptions // optional
}

func (c *Client) MultiGetTransactionBlocks(
	ctx context.Context,
	req MultiGetTransactionBlocksRequest,
) ([]*iotajsonrpc.IotaTransactionBlockResponse, error) {
	resp := []*iotajsonrpc.IotaTransactionBlockResponse{}
	return resp, c.transport.Call(ctx, &resp, multiGetTransactionBlocks, req.Digests, req.Options)
}

type TryGetPastObjectRequest struct {
	ObjectID *iotago.ObjectID
	Version  uint64
	Options  *iotajsonrpc.IotaObjectDataOptions // optional
}

func (c *Client) TryGetPastObject(
	ctx context.Context,
	req TryGetPastObjectRequest,
) (*iotajsonrpc.IotaPastObjectResponse, error) {
	var resp iotajsonrpc.IotaPastObjectResponse
	return &resp, c.transport.Call(ctx, &resp, tryGetPastObject, req.ObjectID, req.Version, req.Options)
}

type TryMultiGetPastObjectsRequest struct {
	PastObjects []*iotajsonrpc.IotaGetPastObjectRequest
	Options     *iotajsonrpc.IotaObjectDataOptions // optional
}

func (c *Client) TryMultiGetPastObjects(
	ctx context.Context,
	req TryMultiGetPastObjectsRequest,
) ([]*iotajsonrpc.IotaPastObjectResponse, error) {
	var resp []*iotajsonrpc.IotaPastObjectResponse
	return resp, c.transport.Call(ctx, &resp, tryMultiGetPastObjects, req.PastObjects, req.Options)
}
