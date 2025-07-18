package iotaclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

type GetDynamicFieldObjectRequest struct {
	ParentObjectID *iotago.ObjectID
	Name           *iotago.DynamicFieldName
}

func (c *Client) GetDynamicFieldObject(
	ctx context.Context,
	req GetDynamicFieldObjectRequest,
) (*iotajsonrpc.IotaObjectResponse, error) {
	var resp iotajsonrpc.IotaObjectResponse
	err := c.transport.Call(ctx, &resp, getDynamicFieldObject, req.ParentObjectID, req.Name)
	if err != nil {
		return &resp, err
	} else if resp.ResponseError() != nil {
		return &resp, resp.ResponseError()
	}
	return &resp, nil
}

type GetDynamicFieldsRequest struct {
	ParentObjectID *iotago.ObjectID
	Cursor         *iotago.ObjectID // optional
	Limit          *uint            // optional
}

func (c *Client) GetDynamicFields(
	ctx context.Context,
	req GetDynamicFieldsRequest,
) (*iotajsonrpc.DynamicFieldPage, error) {
	var resp iotajsonrpc.DynamicFieldPage
	return &resp, c.transport.Call(ctx, &resp, getDynamicFields, req.ParentObjectID, req.Cursor, req.Limit)
}

type GetOwnedObjectsRequest struct {
	// Address is the owner's Iota address
	Address *iotago.Address
	// [optional] Query is the objects query criteria.
	Query *iotajsonrpc.IotaObjectResponseQuery
	// [optional] Cursor is an optional paging cursor.
	// If provided, the query will start from the next item after the specified cursor.
	Cursor *iotago.ObjectID
	// [optional] Limit is the maximum number of items returned per page, defaults to [QUERY_MAX_RESULT_LIMIT_OBJECTS] if not
	// provided
	Limit *uint
}

func (c *Client) GetOwnedObjects(
	ctx context.Context,
	req GetOwnedObjectsRequest,
) (*iotajsonrpc.ObjectsPage, error) {
	var resp iotajsonrpc.ObjectsPage
	err := c.transport.Call(ctx, &resp, getOwnedObjects, req.Address, req.Query, req.Cursor, req.Limit)
	if err != nil {
		return &resp, err
	}
	for i, elt := range resp.Data {
		if elt.ResponseError() != nil {
			return &resp, fmt.Errorf("index: %d: %w", i, elt.ResponseError())
		}
	}
	return &resp, nil
}

type QueryEventsRequest struct {
	Query           *iotajsonrpc.EventFilter
	Cursor          *iotajsonrpc.EventId // optional
	Limit           *uint                // optional
	DescendingOrder bool                 // optional
}

func (c *Client) QueryEvents(
	ctx context.Context,
	req QueryEventsRequest,
) (*iotajsonrpc.EventPage, error) {
	var resp iotajsonrpc.EventPage
	return &resp, c.transport.Call(ctx, &resp, queryEvents, req.Query, req.Cursor, req.Limit, req.DescendingOrder)
}

type QueryTransactionBlocksRequest struct {
	Query           *iotajsonrpc.IotaTransactionBlockResponseQuery
	Cursor          *iotago.TransactionDigest // optional
	Limit           *uint                     // optional
	DescendingOrder bool                      // optional
}

func (c *Client) QueryTransactionBlocks(
	ctx context.Context,
	req QueryTransactionBlocksRequest,
) (*iotajsonrpc.TransactionBlocksPage, error) {
	resp := iotajsonrpc.TransactionBlocksPage{}
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

func (c *Client) ResolveNameServiceAddress(ctx context.Context, iotaName string) (*iotago.Address, error) {
	var resp iotago.Address
	err := c.transport.Call(ctx, &resp, resolveNameServiceAddress, iotaName)
	if err != nil && err.Error() == "nil address" {
		return nil, errors.New("iota name not found")
	}
	return &resp, nil
}

type ResolveNameServiceNamesRequest struct {
	Owner  *iotago.Address
	Cursor *iotago.ObjectID // optional
	Limit  *uint            // optional
}

func (c *Client) ResolveNameServiceNames(
	ctx context.Context,
	req ResolveNameServiceNamesRequest,
) (*iotajsonrpc.IotaNamePage, error) {
	var resp iotajsonrpc.IotaNamePage
	return &resp, c.transport.Call(ctx, &resp, resolveNameServiceNames, req.Owner, req.Cursor, req.Limit)
}

func (c *Client) SubscribeEvent(
	ctx context.Context,
	filter *iotajsonrpc.EventFilter,
	resultCh chan<- *iotajsonrpc.IotaEvent,
) error {
	wsCh := make(chan []byte, 10)
	err := c.transport.Subscribe(ctx, wsCh, subscribeEvent, filter)
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
				var result *iotajsonrpc.IotaEvent
				if err := json.Unmarshal(messageData, &result); err != nil {
					log.Fatal(err)
				}
				resultCh <- result
			}
		}
	}()
	return nil
}

func (c *Client) SubscribeTransaction(
	ctx context.Context,
	filter *iotajsonrpc.TransactionFilter,
	resultCh chan<- *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects],
) error {
	wsCh := make(chan []byte, 10)
	err := c.transport.Subscribe(ctx, wsCh, subscribeTransaction, filter)
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
				var result *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
				if err := json.Unmarshal(messageData, &result); err != nil {
					log.Fatal(err)
				}
				resultCh <- result
			}
		}
	}()
	return nil
}
