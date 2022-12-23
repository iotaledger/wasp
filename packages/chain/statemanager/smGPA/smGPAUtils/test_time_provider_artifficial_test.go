// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package smGPAUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestArtifficialTimeProvider(t *testing.T) {
	now := time.Now()
	tp := NewArtifficialTimeProvider(now)
	ch30s := tp.After(30 * time.Second)
	ch40s := tp.After(40 * time.Second)
	ch20s := tp.After(20 * time.Second)
	ch70s := tp.After(70 * time.Second)
	ch60s := tp.After(60 * time.Second)
	ch10s := tp.After(10 * time.Second)
	ch50s := tp.After(50 * time.Second)
	now = now.Add(25 * time.Second)
	tp.SetNow(now) // start + 25s
	require.True(t, requireTime(t, now, ch10s))
	require.True(t, requireTime(t, now, ch20s))
	require.False(t, requireTime(t, now, ch30s))
	require.False(t, requireTime(t, now, ch40s))
	require.False(t, requireTime(t, now, ch50s))
	require.False(t, requireTime(t, now, ch60s))
	require.False(t, requireTime(t, now, ch70s))
	ch35s := tp.After(10 * time.Second)
	ch80s := tp.After(55 * time.Second)
	now = now.Add(20 * time.Second)
	tp.SetNow(now) // start + 45s
	require.True(t, requireTime(t, now, ch30s))
	require.True(t, requireTime(t, now, ch35s))
	require.True(t, requireTime(t, now, ch40s))
	require.False(t, requireTime(t, now, ch50s))
	require.False(t, requireTime(t, now, ch60s))
	require.False(t, requireTime(t, now, ch70s))
	require.False(t, requireTime(t, now, ch80s))
	now = now.Add(1 * time.Second)
	tp.SetNow(now) // start + 46s
	require.False(t, requireTime(t, now, ch50s))
	require.False(t, requireTime(t, now, ch60s))
	require.False(t, requireTime(t, now, ch70s))
	require.False(t, requireTime(t, now, ch80s))
	ch49s := tp.After(3 * time.Second)
	ch65s := tp.After(19 * time.Second)
	now = now.Add(20 * time.Second)
	tp.SetNow(now) // start + 66s
	require.True(t, requireTime(t, now, ch49s))
	require.True(t, requireTime(t, now, ch50s))
	require.True(t, requireTime(t, now, ch60s))
	require.True(t, requireTime(t, now, ch65s))
	require.False(t, requireTime(t, now, ch70s))
	require.False(t, requireTime(t, now, ch80s))
	now = now.Add(19 * time.Second)
	tp.SetNow(now) // start + 85s
	require.True(t, requireTime(t, now, ch70s))
	require.True(t, requireTime(t, now, ch80s))
	ch90s := tp.After(5 * time.Second)
	now = now.Add(10 * time.Second)
	tp.SetNow(now) // start + 95s
	require.True(t, requireTime(t, now, ch90s))
}

func requireTime(t *testing.T, expectedTime time.Time, ch <-chan time.Time) bool {
	select {
	case result := <-ch:
		require.Equal(t, expectedTime, result)
		return true
	default:
		return false
	}
}
