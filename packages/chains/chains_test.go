package chains

import (
	"net"
	"testing"

	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"golang.org/x/xerrors"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	ch := New(logger, processors.NewConfig())

	nconn := txstream.New("dummyID", logger, func() (addr string, conn net.Conn, err error) {
		return "", nil, xerrors.New("dummy dial error")
	})
	ch.Attach(nconn)
}
