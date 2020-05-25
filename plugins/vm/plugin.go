package vm

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/hashing"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "VM"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin   = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log      *logger.Logger
	vmDaemon = daemon.New()
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		// globally initialize VM
		go vmDaemon.Run()

		<-shutdownSignal

		vmDaemon.Shutdown()
		log.Infof("shutdown VM...  Done")
	})
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}

func getProcessor(programHash hashing.HashValue) (Processor, error) {

}

// RunComputationsAsync runs computations in the background and call function upn finishing it
func RunComputationsAsync(ctx *RuntimeContext, onFinish func()) error {
	if processor, err := getProcessor(ctx.ProgramHash); err != nil {
		return err
	}
	err := vmDaemon.BackgroundWorker(ctx.taskName(), func(shutdownSignal <-chan struct{}) {
		panic("implement me")
	})
	return err
}
