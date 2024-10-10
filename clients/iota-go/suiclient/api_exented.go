package suiclient

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/sui/serialization"
	"github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
)

type GetDynamicFieldObjectRequest struct {
	ParentObjectID *sui.ObjectID
	Name           *sui.DynamicFieldName
}

func (c *Client) GetDynamicFieldObject(
	ctx context.Context,
	req GetDynamicFieldObjectRequest,
) (*suijsonrpc.SuiObjectResponse, error) {
	var resp suijsonrpc.SuiObjectResponse
	return &resp, c.transport.Call(ctx, &resp, getDynamicFieldObject, req.ParentObjectID, req.Name)
}

type GetDynamicFieldsRequest struct {
	ParentObjectID *sui.ObjectID
	Cursor         *sui.ObjectID // optional
	Limit          *uint         // optional
}

func (c *Client) GetDynamicFields(
	ctx context.Context,
	req GetDynamicFieldsRequest,
) (*suijsonrpc.DynamicFieldPage, error) {
	var resp suijsonrpc.DynamicFieldPage
	return &resp, c.transport.Call(ctx, &resp, getDynamicFields, req.ParentObjectID, req.Cursor, req.Limit)
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

func (c *Client) GetOwnedObjects(
	ctx context.Context,
	req GetOwnedObjectsRequest,
) (*suijsonrpc.ObjectsPage, error) {
	var resp suijsonrpc.ObjectsPage
	return &resp, c.transport.Call(ctx, &resp, getOwnedObjects, req.Address, req.Query, req.Cursor, req.Limit)
}

type QueryEventsRequest struct {
	Query           *suijsonrpc.EventFilter
	Cursor          *suijsonrpc.EventId // optional
	Limit           *uint               // optional
	DescendingOrder bool                // optional
}

func (c *Client) QueryEvents(
	ctx context.Context,
	req QueryEventsRequest,
) (*suijsonrpc.EventPage, error) {
	var resp suijsonrpc.EventPage
	return &resp, c.transport.Call(ctx, &resp, queryEvents, req.Query, req.Cursor, req.Limit, req.DescendingOrder)
}

type QueryTransactionBlocksRequest struct {
	Query           *suijsonrpc.SuiTransactionBlockResponseQuery
	Cursor          *sui.TransactionDigest // optional
	Limit           *uint                  // optional
	DescendingOrder bool                   // optional
}

func (c *Client) QueryTransactionBlocks(
	ctx context.Context,
	req QueryTransactionBlocksRequest,
) (*suijsonrpc.TransactionBlocksPage, error) {
	resp := suijsonrpc.TransactionBlocksPage{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		queryTransactionBlocks,
		req.Query,
		req.Cursor,
		req.Limit,
		req.DescendingOrder,
	)
}

func (c *Client) ResolveNameServiceAddress(ctx context.Context, suiName string) (*sui.Address, error) {
	var resp sui.Address
	err := c.transport.Call(ctx, &resp, resolveNameServiceAddress, suiName)
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

func (c *Client) ResolveNameServiceNames(
	ctx context.Context,
	req ResolveNameServiceNamesRequest,
) (*suijsonrpc.SuiNamePage, error) {
	var resp suijsonrpc.SuiNamePage
	return &resp, c.transport.Call(ctx, &resp, resolveNameServiceNames, req.Owner, req.Cursor, req.Limit)
}

func (s *Client) SubscribeEvent(
	ctx context.Context,
	filter *suijsonrpc.EventFilter,
	resultCh chan<- *suijsonrpc.SuiEvent,
) error {
	wsCh := make(chan []byte, 10)
	err := s.transport.Subscribe(ctx, wsCh, subscribeEvent, filter)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case messageData, ok := <-wsCh:
				if !ok {
					return
				}
				var result *suijsonrpc.SuiEvent
				if err := json.Unmarshal(messageData, &result); err != nil {
					log.Fatal(err)
				}
				resultCh <- result
			}
		}
	}()
	return nil
}

func (s *Client) SubscribeTransaction(
	ctx context.Context,
	filter *suijsonrpc.TransactionFilter,
	resultCh chan<- *serialization.TagJson[suijsonrpc.SuiTransactionBlockEffects],
) error {
	wsCh := make(chan []byte, 10)
	err := s.transport.Subscribe(ctx, wsCh, subscribeTransaction, filter)
	if err != nil {
		return err
	}
	go func() {
		defer close(resultCh)
		for {
			select {
			case <-ctx.Done():
				return
			case messageData, ok := <-wsCh:
				if !ok {
					return
				}
				var result *serialization.TagJson[suijsonrpc.SuiTransactionBlockEffects]
				if err := json.Unmarshal(messageData, &result); err != nil {
					log.Fatal(err)
				}
				resultCh <- result
			}
		}
	}()
	return nil
}
