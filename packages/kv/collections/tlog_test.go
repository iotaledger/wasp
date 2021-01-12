package collections

import (
	"math/rand"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
)

func TestTlogBasic(t *testing.T) {
	vars := dict.New()
	tl := NewTimestampedLog(vars, "testTlog")
	assert.Zero(t, tl.MustLen())

	d1 := []byte("datum1")
	d2 := []byte("datum2")
	d3 := []byte("datum3")
	d4 := []byte("datum4")
	d5 := []byte("datum5")
	d6 := []byte("datum6")

	nowis := time.Now().UnixNano()
	nowisNext1 := nowis + 100
	nowisNext2 := nowis + 10000

	tl.MustAppend(nowis, d1)
	assert.EqualValues(t, 1, tl.MustLen())

	tl.MustAppend(nowis, d2)
	assert.EqualValues(t, 2, tl.MustLen())

	tl.MustAppend(nowisNext1, d3)
	assert.EqualValues(t, 3, tl.MustLen())

	assert.Panics(t, func() {
		tl.MustAppend(nowis, d4)
	})
	assert.EqualValues(t, 3, tl.MustLen())

	tl.MustAppend(nowisNext1, d4)
	assert.EqualValues(t, 4, tl.MustLen())

	tl.MustAppend(nowisNext1, d5)
	assert.EqualValues(t, 5, tl.MustLen())

	tl.MustAppend(nowisNext2, d6)
	assert.EqualValues(t, 6, tl.MustLen())

	assert.EqualValues(t, nowis, tl.MustEarliest())
	assert.EqualValues(t, nowisNext2, tl.MustLatest())

	tl.MustAppend(nowisNext2, nil)
	assert.EqualValues(t, 7, tl.MustLen())
}

const (
	numPoints     = 100
	changeTsEvery = 5
	step          = 1000
)

func initLog(t *testing.T, tl *TimestampedLog) {
	data := make([][]byte, numPoints)

	for i := 0; i < numPoints; i++ {
		data[i] = util.Uint32To4Bytes(uint32(rand.Int31()))
	}
	timeStart := time.Now().UnixNano()
	nowis := timeStart

	var latest int64
	for i, d := range data {
		tl.MustAppend(nowis, d)
		latest = nowis
		if (i+1)%changeTsEvery == 0 {
			nowis += step
		}
	}
	assert.EqualValues(t, timeStart, tl.MustEarliest())
	assert.EqualValues(t, latest, tl.MustLatest())
}

func TestTlogBig(t *testing.T) {
	vars := dict.New()
	tl := NewTimestampedLog(vars, "testTimestampedlog")

	initLog(t, tl)

	assert.EqualValues(t, numPoints, tl.MustLen())

	latest := tl.MustLatest()
	assert.Panics(t, func() {
		tl.MustAppend(latest-10000, nil)
	})

	tslice := tl.MustTakeTimeSlice(tl.MustEarliest(), tl.MustEarliest())
	assert.EqualValues(t, changeTsEvery, tslice.NumPoints())

	tslice = tl.MustTakeTimeSlice(tl.MustEarliest(), tl.MustEarliest()+step)
	assert.EqualValues(t, 2*changeTsEvery, tslice.NumPoints())

	tslice = tl.MustTakeTimeSlice(tl.MustLatest(), tl.MustLatest())
	assert.EqualValues(t, changeTsEvery, tslice.NumPoints())

	tslice = tl.MustTakeTimeSlice(tl.MustEarliest(), tl.MustLatest())
	assert.EqualValues(t, tl.MustLen(), tslice.NumPoints())
	assert.EqualValues(t, tl.MustLen(), tslice.NumPoints())
}
