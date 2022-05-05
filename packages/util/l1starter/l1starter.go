package l1starter

import (
	"context"
	"flag"
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
)

type L1Starter struct {
	Config             nodeconn.L1Config
	privtangleNumNodes int
	Privtangle         *privtangle.PrivTangle
}

const pvtTangleAPIPort = 16500

// New sets up the CLI flags relevant to L1/privtangle configuration in the given FlagSet.
func New(flags *flag.FlagSet) *L1Starter {
	s := &L1Starter{}
	flags.StringVar(&s.Config.Hostname, "layer1-host", "", "layer1 host")
	flags.IntVar(&s.Config.APIPort, "layer1-api-port", 0, "layer1 API port")
	flags.IntVar(&s.Config.FaucetPort, "layer1-faucet-port", 0, "layer1 faucet port")
	flags.IntVar(&s.privtangleNumNodes, "privtangle-num-nodes", 2, "number of hornet nodes to be spawned in the private tangle")
	return s
}

func (s *L1Starter) PrivtangleEnabled() bool {
	return s.Config.Hostname == "" || s.Privtangle != nil
}

// StartPrivtangleIfNecessary starts a private tangle, unless an L1 host was provided via cli flags
func (s *L1Starter) StartPrivtangleIfNecessary(logfunc privtangle.LogFunc) {
	if s.Config.Hostname != "" || s.Privtangle != nil {
		return
	}
	s.Privtangle = privtangle.Start(
		context.Background(),
		path.Join(os.TempDir(), "privtangle"),
		pvtTangleAPIPort,
		s.privtangleNumNodes,
		logfunc,
	)
	s.Config = nodeconn.L1Config{
		Hostname:   "http://localhost",
		APIPort:    s.Privtangle.NodePortRestAPI(0),
		FaucetPort: s.Privtangle.NodePortFaucet(0),
		FaucetKey:  s.Privtangle.FaucetKeyPair,
	}
}

func (s *L1Starter) Stop() {
	if s.Privtangle != nil {
		s.Privtangle.Stop()
	}
}
