package profiling

import (
	"net/http"
	"runtime"

	// import required to profile
	_ "net/http/pprof"

	profile "github.com/bygui86/multi-profile/v2"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
)

const PluginName = "Profiling"

var log *logger.Logger

// Init gets the plugin instance.
func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	if !parameters.GetBool(parameters.ProfilingEnabled) {
		return
	}

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	runtime.SetCPUProfileRate(5)

	if parameters.GetBool(parameters.ProfilingWriteProfiles) {
		profConfig := &profile.Config{
			Path:                "./profiles",
			EnableInterruptHook: true,
		}

		profs := make([]*profile.Profile, 7)
		profs[0] = profile.CPUProfile(profConfig).Start()
		profs[1] = profile.MemProfile(profConfig).Start()
		profs[2] = profile.GoroutineProfile(profConfig).Start()
		profs[3] = profile.MutexProfile(profConfig).Start()
		profs[4] = profile.BlockProfile(profConfig).Start()
		profs[5] = profile.TraceProfile(profConfig).Start()
		profs[6] = profile.ThreadCreationProfile(profConfig).Start()

		err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
			<-shutdownSignal
			for _, p := range profs {
				p.Stop()
			}
			log.Infof("%s shutdown,writing performance profiles", PluginName)
		})
		if err != nil {
			panic(err)
		}
	}

	go func() {
		bindAddr := parameters.GetString(parameters.ProfilingBindAddress)
		log.Infof("%s started, bind-address=%s", PluginName, bindAddr)
		err := http.ListenAndServe(bindAddr, nil)
		if err != nil {
			panic(err)
		}
	}()
}
