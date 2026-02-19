package iotagraphql

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
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

	dataChan, _, err := TransactionsBySigner(ctx, c.wsClient, *filter.FromAddress)
	if err != nil {
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}

	go c.forwardTransactionResponses(ctx, dataChan, resultCh, *filter.ChangedObject)

	return nil
}

func (c *GraphQLClient) forwardTransactionResponses(
	ctx context.Context,
	dataChan <-chan TransactionsBySignerWsResponse,
	resultCh chan<- *IotaTransactionBlockEffects,
	changedObjectFilter iotago.ObjectID,
) {
	for {
		select {
		case <-ctx.Done():
			c.log.LogWarnf("context done: %v", ctx.Err())
			return
		case resp, ok := <-dataChan:
			if !ok {
				c.log.LogWarnf("data channel closed: %v", dataChan)
				return
			}
			if len(resp.Errors) > 0 {
				c.log.LogErrorf("error forwarding transaction responses: %v", resp.Errors)
				continue
			}
			txBlock := resp.GetTransactionBlock()
			if txBlock == nil {
				c.log.LogWarnf("can't get transaction block from response: %v", resp)
				continue
			}

			if !transactionChangedObject(txBlock, changedObjectFilter) {
				continue
			}

			effects := convertGraphQLTxToEffects(txBlock)

			select {
			case resultCh <- effects:
			case <-ctx.Done():
				c.log.LogWarnf("context done: %v", ctx.Err())
				return
			}
		}
	}
}

func transactionChangedObject(txBlock *TransactionsBySignerTransactionsTransactionBlock, objectID iotago.ObjectID) bool {
	for _, change := range txBlock.Effects.ObjectChanges.Nodes {
		if change.Address == objectID {
			return true
		}
	}
	return false
}

func convertGraphQLTxToEffects(txBlock *TransactionsBySignerTransactionsTransactionBlock) *IotaTransactionBlockEffects {
	mutated := make([]OwnedObjectRef, 0, len(txBlock.Effects.ObjectChanges.Nodes))

	for _, change := range txBlock.Effects.ObjectChanges.Nodes {
		objectID := iotago.ObjectID(change.Address)
		mutated = append(mutated, OwnedObjectRef{
			Reference: IotaObjectRef{
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

	// Format: "package" or "package::module"
	var emittingModule string
	if filter.MoveEventType.Module == "" {
		emittingModule = filter.MoveEventType.Address.String()
	} else {
		emittingModule = fmt.Sprintf("%s::%s", filter.MoveEventType.Address, filter.MoveEventType.Module)
	}

	dataChan, _, err := EventsByModule(ctx, c.wsClient, emittingModule)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	go c.forwardEventResponses(ctx, dataChan, resultCh)

	return nil
}

func (c *GraphQLClient) forwardEventResponses(
	ctx context.Context,
	dataChan <-chan EventsByModuleWsResponse,
	resultCh chan<- *IotaEvent,
) {
	for {
		select {
		case <-ctx.Done():
			c.log.LogWarnf("context done: %v", ctx.Err())
			return
		case resp, ok := <-dataChan:
			if !ok {
				c.log.LogWarnf("data channel closed: %v", dataChan)
				return
			}
			if len(resp.Errors) > 0 {
				c.log.LogErrorf("error forwarding event responses: %v", resp.Errors)
				continue
			}
			event := resp.GetEvent()
			if event == nil {
				c.log.LogWarnf("can't get event from response: %v", resp)
				continue
			}

			iotaEvent := convertGraphQLEventToIotaEvent(event)

			select {
			case resultCh <- iotaEvent:
			case <-ctx.Done():
				c.log.LogWarnf("context done: %v", ctx.Err())
				return
			}
		}
	}
}

func convertGraphQLEventToIotaEvent(event *EventsByModuleEventsEvent) *IotaEvent {
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
		ParsedJSON:        event.Json,
	}
}
