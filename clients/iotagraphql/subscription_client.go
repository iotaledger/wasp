package iotagraphql

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

func (c *GraphQLClient) SubscribeTransaction(
	ctx context.Context,
	filter *TransactionFilter,
	resultCh chan<- *IotaTransactionBlockEffects,
) error {
	if filter.FromAddress == nil {
		return fmt.Errorf("subscribeTransaction via GraphQL requires FromAddress filter")
	}

	if filter.ChangedObject == nil {
		return fmt.Errorf("subscribeTransaction via GraphQL requires ChangedObject filter")
	}

	wsClient, err := c.newWebSocketClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to start WebSocket connection: %w", err)
	}

	fmt.Printf("subscribing to transactions from address: %s\n", filter.FromAddress.String())
	dataChan, _, err := graphqltypes.TransactionsBySigner(ctx, wsClient, *filter.FromAddress)
	if err != nil {
		wsClient.Close()
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}

	go c.forwardTransactionResponses(ctx, dataChan, resultCh, *filter.ChangedObject)

	return nil
}

func (c *GraphQLClient) forwardTransactionResponses(
	ctx context.Context,
	dataChan <-chan graphqltypes.TransactionsBySignerWsResponse,
	resultCh chan<- *IotaTransactionBlockEffects,
	changedObjectFilter iotago.ObjectID,
) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("context done: %v", ctx.Err())
			return
		case resp, ok := <-dataChan:
			fmt.Printf("received transaction response: %+v\n", resp)
			if !ok {
				fmt.Printf("data channel closed: %v", dataChan)
				return
			}
			if len(resp.Errors) > 0 {
				fmt.Printf("error forwarding transaction responses: %v", resp.Errors)
				continue
			}
			txBlock := resp.GetTxBySignerTransactionBlock()
			if txBlock == nil {
				fmt.Printf("can't get transaction block from response: %v", resp)
				continue
			}

			if !transactionChangedObject(txBlock, changedObjectFilter) {
				continue
			}

			effects := convertGraphQLTxToEffects(txBlock)

			fmt.Printf("forwarding transaction effects: %+v", effects)
			select {
			case resultCh <- effects:
			case <-ctx.Done():
				fmt.Printf("context done: %v", ctx.Err())
				return
			}
		}
	}
}

func transactionChangedObject(txBlock *graphqltypes.TransactionsBySignerTransactionsTransactionBlock, objectID iotago.ObjectID) bool {
	for _, change := range txBlock.Effects.ObjectChanges.Nodes {
		if change.Address == objectID {
			return true
		}
	}
	return false
}

func convertGraphQLTxToEffects(txBlock *graphqltypes.TransactionsBySignerTransactionsTransactionBlock) *IotaTransactionBlockEffects {
	mutated := make([]struct {
		Reference iotago.ObjectRef
	}, 0, len(txBlock.Effects.ObjectChanges.Nodes))

	for _, change := range txBlock.Effects.ObjectChanges.Nodes {
		objectID := iotago.ObjectID(change.Address)
		mutated = append(mutated, struct {
			Reference iotago.ObjectRef
		}{
			Reference: iotago.ObjectRef{
				ObjectID: &objectID,
				Version:  iotago.SequenceNumber(change.OutputState.Version),
			},
		})
	}

	return &IotaTransactionBlockEffects{
		V1: &IotaTransactionBlockEffectsV1{
			Mutated: mutated,
		},
	}
}

func (c *GraphQLClient) SubscribeEvent(
	ctx context.Context,
	filter *IotaEventFilter,
	resultCh chan<- *IotaEvent,
) error {
	if filter.MoveModule == nil {
		return fmt.Errorf("subscribeEvent via GraphQL requires MoveModule filter")
	}

	if filter.MoveModule.Package == nil {
		return fmt.Errorf("subscribeEvent via GraphQL requires MoveModule.Package filter")
	}

	wsClient, err := c.newWebSocketClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to start WebSocket connection: %w", err)
	}

	// Format: "package" or "package::module"
	var emittingModule string
	if filter.MoveEventType.Module == "" {
		emittingModule = filter.MoveEventType.Address.String()
	} else {
		emittingModule = fmt.Sprintf("%s::%s", filter.MoveEventType.Address, filter.MoveEventType.Module)
	}

	fmt.Printf("subscribing to events from module: %s\n", emittingModule)
	dataChan, _, err := graphqltypes.EventsByModule(ctx, wsClient, emittingModule)
	if err != nil {
		wsClient.Close()
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	go c.forwardEventResponses(ctx, dataChan, resultCh)

	return nil
}

func (c *GraphQLClient) forwardEventResponses(
	ctx context.Context,
	dataChan <-chan graphqltypes.EventsByModuleWsResponse,
	resultCh chan<- *IotaEvent,
) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("context done: %v", ctx.Err())
			return
		case resp, ok := <-dataChan:
			fmt.Printf("received event response: %+v\n", resp)
			if !ok {
				fmt.Printf("data channel closed: %v", dataChan)
				return
			}
			if len(resp.Errors) > 0 {
				fmt.Printf("error forwarding event responses: %v", resp.Errors)
				continue
			}
			event := resp.GetEvent()
			if event == nil {
				fmt.Printf("can't get event from response: %v", resp)
				continue
			}

			iotaEvent := convertGraphQLEventToIotaEvent(event)

			fmt.Printf("forwarding event: %+v", iotaEvent)
			select {
			case resultCh <- iotaEvent:
			case <-ctx.Done():
				fmt.Printf("context done: %v", ctx.Err())
				return
			}
		}
	}
}

func convertGraphQLEventToIotaEvent(event *graphqltypes.EventsByModuleEventsEvent) *IotaEvent {
	var sender *iotago.Address
	senderAddr := event.Sender.Address
	if senderAddr != (iotago.Address{}) {
		sender = &senderAddr
	}

	var eventType *iotago.StructTag
	typeRepr := event.Type.Repr
	if typeRepr != "" {
		st, err := iotago.StructTagFromString(typeRepr)
		if err == nil {
			eventType = st
		}
	}

	packageAddr := event.SendingModule.Package.Address
	packageID := iotago.ObjectID(packageAddr)

	return &IotaEvent{
		PackageID:         &packageID,
		TransactionModule: iotago.Identifier(event.SendingModule.Name),
		Sender:            sender,
		Type:              eventType,
		Bcs:               event.Bcs,
	}
}
