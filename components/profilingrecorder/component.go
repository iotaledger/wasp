// Package profilingrecorder provides functionality for recording and analyzing performance metrics.
package profilingrecorder

import (
	"context"
	"runtime"

	profile "github.com/bygui86/multi-profile/v2"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/packages/daemon"
)

func init() {
	Component = &app.Component{
		Name:      "profilingRecorder",
		Params:    params,
		IsEnabled: func(_ *dig.Container) bool { return ParamsProfilingRecorder.Enabled },
		Run:       run,
	}
}

var Component *app.Component

func run() error {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	runtime.SetCPUProfileRate(5)

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

	err := Component.Daemon().BackgroundWorker(Component.Name, func(ctx context.Context) {
		<-ctx.Done()
		for _, p := range profs {
			p.Stop()
		}
	}, daemon.PriorityProfilingRecorder)
	if err != nil {
		panic(err)
	}

	return nil
}
