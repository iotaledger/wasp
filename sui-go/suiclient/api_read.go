package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (s *Client) GetChainIdentifier(ctx context.Context) (string, error) {
	var resp string
	return resp, s.http.CallContext(ctx, &resp, getChainIdentifier)
}

func (s *Client) GetCheckpoint(ctx context.Context, checkpointId *suijsonrpc.BigInt) (*suijsonrpc.Checkpoint, error) {
	var resp suijsonrpc.Checkpoint
	return &resp, s.http.CallContext(ctx, &resp, getCheckpoint, checkpointId)
}

type GetCheckpointsRequest struct {
	Cursor          *suijsonrpc.BigInt // optional
	Limit           *uint64            // optional
	DescendingOrder bool
}

func (s *Client) GetCheckpoints(ctx context.Context, req GetCheckpointsRequest) (*suijsonrpc.CheckpointPage, error) {
	var resp suijsonrpc.CheckpointPage
	return &resp, s.http.CallContext(ctx, &resp, getCheckpoints, req.Cursor, req.Limit, req.DescendingOrder)
}

func (s *Client) GetEvents(ctx context.Context, digest *sui.TransactionDigest) ([]*suijsonrpc.SuiEvent, error) {
	var resp []*suijsonrpc.SuiEvent
	return resp, s.http.CallContext(ctx, &resp, getEvents, digest)
}

func (s *Client) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	var resp string
	return resp, s.http.CallContext(ctx, &resp, getLatestCheckpointSequenceNumber)
}

// TODO getLoadedChildObjects

type GetObjectRequest struct {
	ObjectID *sui.ObjectID
	Options  *suijsonrpc.SuiObjectDataOptions // optional
}

func (s *Client) GetObject(ctx context.Context, req GetObjectRequest) (*suijsonrpc.SuiObjectResponse, error) {
	var resp suijsonrpc.SuiObjectResponse
	return &resp, s.http.CallContext(ctx, &resp, getObject, req.ObjectID, req.Options)
}

func (s *Client) GetProtocolConfig(
	ctx context.Context,
	version *suijsonrpc.BigInt, // optional
) (*suijsonrpc.ProtocolConfig, error) {
	var resp suijsonrpc.ProtocolConfig
	return &resp, s.http.CallContext(ctx, &resp, getProtocolConfig, version)
}

func (s *Client) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	var resp string
	return resp, s.http.CallContext(ctx, &resp, getTotalTransactionBlocks)
}

type GetTransactionBlockRequest struct {
	Digest  *sui.TransactionDigest
	Options *suijsonrpc.SuiTransactionBlockResponseOptions // optional
}

func (s *Client) GetTransactionBlock(ctx context.Context, req GetTransactionBlockRequest) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	resp := suijsonrpc.SuiTransactionBlockResponse{}
	return &resp, s.http.CallContext(ctx, &resp, getTransactionBlock, req.Digest, req.Options)
}

type MultiGetObjectsRequest struct {
	ObjectIDs []*sui.ObjectID
	Options   *suijsonrpc.SuiObjectDataOptions // optional
}

func (s *Client) MultiGetObjects(ctx context.Context, req MultiGetObjectsRequest) ([]suijsonrpc.SuiObjectResponse, error) {
	var resp []suijsonrpc.SuiObjectResponse
	return resp, s.http.CallContext(ctx, &resp, multiGetObjects, req.ObjectIDs, req.Options)
}

type MultiGetTransactionBlocksRequest struct {
	Digests []*sui.Digest
	Options *suijsonrpc.SuiTransactionBlockResponseOptions // optional
}

func (s *Client) MultiGetTransactionBlocks(
	ctx context.Context,
	req MultiGetTransactionBlocksRequest,
) ([]*suijsonrpc.SuiTransactionBlockResponse, error) {
	resp := []*suijsonrpc.SuiTransactionBlockResponse{}
	return resp, s.http.CallContext(ctx, &resp, multiGetTransactionBlocks, req.Digests, req.Options)
}

type TryGetPastObjectRequest struct {
	ObjectID *sui.ObjectID
	Version  uint64
	Options  *suijsonrpc.SuiObjectDataOptions // optional
}

func (s *Client) TryGetPastObject(
	ctx context.Context,
	req TryGetPastObjectRequest,
) (*suijsonrpc.SuiPastObjectResponse, error) {
	var resp suijsonrpc.SuiPastObjectResponse
	return &resp, s.http.CallContext(ctx, &resp, tryGetPastObject, req.ObjectID, req.Version, req.Options)
}

type TryMultiGetPastObjectsRequest struct {
	PastObjects []*suijsonrpc.SuiGetPastObjectRequest
	Options     *suijsonrpc.SuiObjectDataOptions // optional
}

func (s *Client) TryMultiGetPastObjects(
	ctx context.Context,
	req TryMultiGetPastObjectsRequest,
) ([]*suijsonrpc.SuiPastObjectResponse, error) {
	var resp []*suijsonrpc.SuiPastObjectResponse
	return resp, s.http.CallContext(ctx, &resp, tryMultiGetPastObjects, req.PastObjects, req.Options)
}
