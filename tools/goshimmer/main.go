package main

import (
	"github.com/iotaledger/goshimmer/dapps/waspconn"
	_ "net/http/pprof"

	"github.com/iotaledger/goshimmer/plugins"
	"github.com/iotaledger/hive.go/node"
)

func main() {
	node.Run(
		plugins.Core,
		plugins.Research,
		plugins.UI,
		plugins.WebAPI,
		waspconn.PLUGINS,
	)
}
