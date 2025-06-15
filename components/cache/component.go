// Package cache implements caching functionality for the application.
package cache

import (
	"context"
	"time"

	"fortio.org/safecast"

	"github.com/dustin/go-humanize"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"

	"github.com/iotaledger/wasp/packages/cache"
	"github.com/iotaledger/wasp/packages/daemon"
)

func init() {
	Component = &app.Component{
		Name:      "Cache",
		IsEnabled: func(_ *dig.Container) bool { return ParamsCache.Enabled },
		Params:    params,
		Run:       run,
	}
}

var Component *app.Component

func run() error {
	size, err := humanize.ParseBytes(ParamsCache.CacheSize)
	if err != nil {
		Component.LogPanicf("invalid CacheSize")
	}

	sizeInt, err := safecast.Convert[int](size)
	if err != nil {
		Component.LogPanicf("CacheSize overflows int")
	}
	if err := cache.SetCacheSize(sizeInt); err != nil {
		Component.LogPanic(err.Error())
	}

	if err := Component.Daemon().BackgroundWorker("Cache statistics", func(ctx context.Context) {
		Component.LogInfof("Started cache statistics ticker")

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

				hitsPercent := float64(stats.GetCalls-stats.Misses) / float64(stats.GetCalls+1) * 100.0

				cacheUsed := humanize.IBytes(stats.BytesSize)
				cacheSize := humanize.IBytes(stats.MaxBytesSize)

				Component.LogDebugf("handles: %d, gets: %d, sets: %d, misses: %d, hits(%%): %.3f, misses(%%): %.3f, collisions: %d, corruptions: %d, entries: %d, bytesize: %s, maxbytesize: %s",
					stats.NumHandles, stats.GetCalls, stats.SetCalls, stats.Misses, hitsPercent, 100.0-hitsPercent, stats.Collisions,
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
