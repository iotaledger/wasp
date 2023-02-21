package events

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/webapi/websocket"
)

const (
	ISCEventKindNewBlock      = "new_block"
	ISCEventKindReceipt       = "receipt" // issuer will be the request sender
	ISCEventKindSmartContract = "contract"
	ISCEventIssuerVM          = "vm"
)

type ISCEvent struct {
	Kind      string
	Issuer    string // (AgentID) nil means issued by the VM
	RequestID string // (isc.RequestID)
	ChainID   string // (isc.ChainID)
	Content   any
}

// kind is not printed right now, because it is added when calling p.publish
func (e *ISCEvent) String() string {
	issuerStr := "vm"
	if e.Issuer != "" {
		issuerStr = e.Issuer
	}
	// chainid | issuer (kind):
	return fmt.Sprintf("%s | %s (%s): %v", e.ChainID, issuerStr, e.Kind, e.Content)
}

type EventManager struct {
	client    *websockethub.Client
	publisher *websocket.Publisher
}

func (e *EventManager) SendEvent(chainID string, kind string) {
	if e.publisher.IsClientAllowed(e.client, chainID, kind) {
		
	}
}
