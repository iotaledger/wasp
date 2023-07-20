package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/packages/cache"
	"github.com/iotaledger/wasp/packages/daemon"
)

func init() {
	Component = &app.Component{
		Name:   "Cache",
		Params: params,
		Run:    run,
	}
}

var (
	Component *app.Component
)

func run() error {
	// cache is disabled, don't initialize it
	if !ParamsCache.CacheEnabled {
		return nil
	}

	size, err := humanize.ParseBytes(ParamsCache.CacheSize)
	if err != nil {
		Component.LogPanicf("invalid CacheSize")
	}

	if err := cache.InitCache(int(size)); err != nil {
		Component.LogPanic("cache initialization failed")
	}
	Component.LogInfof("cache initialized ...")

	if err := Component.Daemon().BackgroundWorker("Cache statistics", func(ctx context.Context) {
		ticker := time.NewTicker(ParamsCache.CacheStatsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// print cache statistics
				stats := cache.GetStats()

				// no statistics, continue
				if stats == nil {
					continue
				}

				var hitsPercent = "-"
				var missesPercent = "-"
				if stats.GetCalls > 0 { // Avoid division by zero
					hp := float64(stats.GetCalls-stats.Misses) / float64(stats.GetCalls) * 100.0
					hitsPercent = fmt.Sprintf("%.3f", hp)
					missesPercent = fmt.Sprintf("%.3f", 100.0-hp)
				}
				cacheUsed := humanize.IBytes(stats.BytesSize)
				cacheSize := humanize.IBytes(stats.MaxBytesSize)

				Component.LogDebugf("gets: %d, sets: %d, misses: %d, hits(%%): %s, misses(%%): %s, collisions: %d, corruptions: %d, entries: %d, bytesize: %s, maxbytesize: %s",
					stats.GetCalls, stats.SetCalls, stats.Misses, hitsPercent, missesPercent, stats.Collisions,
					stats.Corruptions, stats.EntriesCount, cacheUsed, cacheSize)

			case <-ctx.Done():
				return
			}
		}
	}, daemon.PriorityCloseDatabase); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
