package l1starter

import (
	"context"
	"flag"
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
	"github.com/iotaledger/wasp/packages/testutil/privtangle/privtangledefaults"
)

type L1Starter struct {
	Config             l1connection.Config
	privtangleNumNodes int
	Privtangle         *privtangle.PrivTangle
}

// New sets up the CLI flags relevant to L1/privtangle configuration in the given FlagSet.
func New(l1flags, inxFlags *flag.FlagSet) *L1Starter {
	s := &L1Starter{}
	l1flags.StringVar(&s.Config.APIAddress, "layer1-api", "", "layer1 API address")
	inxFlags.StringVar(&s.Config.INXAddress, "layer1-inx", "", "layer1 INX address")
	l1flags.StringVar(&s.Config.FaucetAddress, "layer1-faucet", "", "layer1 faucet port")
	l1flags.BoolVar(&s.Config.UseRemotePoW, "layer1-remote-pow", false, "use remote PoW (must be enabled on the Hornet node)")
	l1flags.IntVar(&s.privtangleNumNodes, "privtangle-num-nodes", 2, "number of hornet nodes to be spawned in the private tangle")
	return s
}

func (s *L1Starter) PrivtangleEnabled() bool {
	return s.Config.APIAddress == "" || s.Privtangle != nil
}

// StartPrivtangleIfNecessary starts a private tangle, unless an L1 host was provided via cli flags
func (s *L1Starter) StartPrivtangleIfNecessary(logfunc privtangle.LogFunc) {
	if s.Config.APIAddress != "" {
		return
	}
	s.Privtangle = privtangle.Start(
		context.Background(),
		path.Join(os.TempDir(), "privtangle"),
		privtangledefaults.BasePort,
		s.privtangleNumNodes,
		logfunc,
	)
	s.Config = s.Privtangle.L1Config()
}

func (s *L1Starter) Stop() {
	if s.Privtangle != nil {
		s.Privtangle.Stop()
	}
}

func (s *L1Starter) StartExistingServers() {
	if s.Privtangle != nil {
		s.Privtangle.StartServers(false)
	}
}
