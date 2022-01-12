package nodeconnimpl

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"golang.org/x/xerrors"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)

	var addr ledgerstate.AliasAddress
	nconn := txstream.New("dummyID", logger, func() (addr string, conn net.Conn, err error) {
		return "", nil, xerrors.New("dummy dial error")
	})
	nconnimpl := NewNodeConnection(nconn, nodeconnmetrics.NewEmptyNodeConnectionMetrics(), logger)
	nconnimpl.AttachToTransactionReceived(&addr, func(*ledgerstate.Transaction) {})
	nconnimpl.AttachToInclusionStateReceived(&addr, func(ledgerstate.TransactionID, ledgerstate.InclusionState) {})
	nconnimpl.AttachToOutputReceived(&addr, func(ledgerstate.Output) {})
	nconnimpl.AttachToUnspentAliasOutputReceived(&addr, func(*ledgerstate.AliasOutput, time.Time) {})
}
