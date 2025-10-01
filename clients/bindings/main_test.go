package bindings_test

import (
	"testing"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/bindings"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

var bindingClient clients.L1Client

func TestMain(m *testing.M) {
	l1starter.TestMain(m)

	bindingClient = bindings.NewBindingClient(iotaconn.LocalnetEndpointURL)
}
