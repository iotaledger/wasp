package chains

import (
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"golang.org/x/xerrors"
	"net"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	nconn := nodeconn.New("dummyNetId", logger, func() (addr string, conn net.Conn, err error) {
		time.Sleep(1 * time.Second)
		return "127.0.0.1", nil, xerrors.New("dummy error")
	})
	_ = New(logger, nconn)
}
