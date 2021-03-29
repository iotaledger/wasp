package chains

import (
	txstream "github.com/iotaledger/goshimmer/packages/txstream/client"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"golang.org/x/xerrors"
	"net"
	"testing"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	ch := New(logger)

	nconn := txstream.New("dummyID", logger, func() (addr string, conn net.Conn, err error) {
		return "", nil, xerrors.New("dummy dial error")
	})
	ch.Attach(nconn)
}
