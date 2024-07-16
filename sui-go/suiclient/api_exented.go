package suiclient

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/tidwall/gjson"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type GetDynamicFieldObjectRequest struct {
	ParentObjectID *sui.ObjectID
	Name           *sui.DynamicFieldName
}

func (s *Client) GetDynamicFieldObject(
	ctx context.Context,
	req GetDynamicFieldObjectRequest,
) (*suijsonrpc.SuiObjectResponse, error) {
	var resp suijsonrpc.SuiObjectResponse
	return &resp, s.http.CallContext(ctx, &resp, getDynamicFieldObject, req.ParentObjectID, req.Name)
}

type GetDynamicFieldsRequest struct {
	ParentObjectID *sui.ObjectID
	Cursor         *sui.ObjectID // optional
	Limit          *uint         // optional
}

func (s *Client) GetDynamicFields(
	ctx context.Context,
	req GetDynamicFieldsRequest,
) (*suijsonrpc.DynamicFieldPage, error) {
	var resp suijsonrpc.DynamicFieldPage
	return &resp, s.http.CallContext(ctx, &resp, getDynamicFields, req.ParentObjectID, req.Cursor, req.Limit)
}

type GetOwnedObjectsRequest struct {
	// Address is the owner's Sui address
	Address *sui.Address
	// [optional] Query is the objects query criteria.
	Query *suijsonrpc.SuiObjectResponseQuery
	// [optional] Cursor is an optional paging cursor.
	// If provided, the query will start from the next item after the specified cursor.
	Cursor *sui.ObjectID
	// [optional] Limit is the maximum number of items returned per page, defaults to [QUERY_MAX_RESULT_LIMIT_OBJECTS] if not
	// provided
	Limit *uint
}

func (s *Client) GetOwnedObjects(
	ctx context.Context,
	req GetOwnedObjectsRequest,
) (*suijsonrpc.ObjectsPage, error) {
	var resp suijsonrpc.ObjectsPage
	return &resp, s.http.CallContext(ctx, &resp, getOwnedObjects, req.Address, req.Query, req.Cursor, req.Limit)
}

type QueryEventsRequest struct {
	Query           *suijsonrpc.EventFilter
	Cursor          *suijsonrpc.EventId // optional
	Limit           *uint               // optional
	DescendingOrder bool                // optional
}

func (s *Client) QueryEvents(
	ctx context.Context,
	req QueryEventsRequest,
) (*suijsonrpc.EventPage, error) {
	var resp suijsonrpc.EventPage
	return &resp, s.http.CallContext(ctx, &resp, queryEvents, req.Query, req.Cursor, req.Limit, req.DescendingOrder)
}

type QueryTransactionBlocksRequest struct {
	Query           *suijsonrpc.SuiTransactionBlockResponseQuery
	Cursor          *sui.TransactionDigest // optional
	Limit           *uint                  // optional
	DescendingOrder bool                   // optional
}

func (s *Client) QueryTransactionBlocks(
	ctx context.Context,
	req QueryTransactionBlocksRequest,
) (*suijsonrpc.TransactionBlocksPage, error) {
	resp := suijsonrpc.TransactionBlocksPage{}
	return &resp, s.http.CallContext(ctx, &resp, queryTransactionBlocks, req.Query, req.Cursor, req.Limit, req.DescendingOrder)
}

func (s *Client) ResolveNameServiceAddress(ctx context.Context, suiName string) (*sui.Address, error) {
	var resp sui.Address
	err := s.http.CallContext(ctx, &resp, resolveNameServiceAddress, suiName)
	if err != nil && err.Error() == "nil address" {
		return nil, errors.New("sui name not found")
	}
	return &resp, nil
}

type ResolveNameServiceNamesRequest struct {
	Owner  *sui.Address
	Cursor *sui.ObjectID // optional
	Limit  *uint         // optional
}

func (s *Client) ResolveNameServiceNames(
	ctx context.Context,
	req ResolveNameServiceNamesRequest,
) (*suijsonrpc.SuiNamePage, error) {
	var resp suijsonrpc.SuiNamePage
	return &resp, s.http.CallContext(ctx, &resp, resolveNameServiceNames, req.Owner, req.Cursor, req.Limit)
}

func (s *WebsocketClient) SubscribeEvent(
	ctx context.Context,
	filter *suijsonrpc.EventFilter,
	resultCh chan suijsonrpc.SuiEvent,
) error {
	resp := make(chan []byte, 10)
	err := s.ws.CallContext(ctx, resp, subscribeEvent, filter)
	if err != nil {
		return err
	}
	go func() {
		for messageData := range resp {
			var result suijsonrpc.SuiEvent
			if gjson.ParseBytes(messageData).Get("error").Exists() {
				log.Fatal(gjson.ParseBytes(messageData).Get("error").String())
			}

			err := json.Unmarshal([]byte(gjson.ParseBytes(messageData).Get("params.result").String()), &result)
			if err != nil {
				log.Fatal(err)
			}

			resultCh <- result
		}
	}()
	return nil
}

// TODO SubscribeTransaction
