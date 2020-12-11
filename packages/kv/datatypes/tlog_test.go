package datatypes

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
	tl := NewMustTimestampedLog(vars, "testTlog")
	assert.Zero(t, tl.Len())

	d1 := []byte("datum1")
	d2 := []byte("datum2")
	d3 := []byte("datum3")
	d4 := []byte("datum4")
	d5 := []byte("datum5")
	d6 := []byte("datum6")

	nowis := time.Now().UnixNano()
	nowisNext1 := nowis + 100
	nowisNext2 := nowis + 10000

	tl.Append(nowis, d1)
	assert.EqualValues(t, 1, tl.Len())

	tl.Append(nowis, d2)
	assert.EqualValues(t, 2, tl.Len())

	tl.Append(nowisNext1, d3)
	assert.EqualValues(t, 3, tl.Len())

	assert.Panics(t, func() {
		tl.Append(nowis, d4)
	})
	assert.EqualValues(t, 3, tl.Len())

	tl.Append(nowisNext1, d4)
	assert.EqualValues(t, 4, tl.Len())

	tl.Append(nowisNext1, d5)
	assert.EqualValues(t, 5, tl.Len())

	tl.Append(nowisNext2, d6)
	assert.EqualValues(t, 6, tl.Len())

	assert.EqualValues(t, nowis, tl.Earliest())
	assert.EqualValues(t, nowisNext2, tl.Latest())

	tl.Append(nowisNext2, nil)
	assert.EqualValues(t, 7, tl.Len())
}

const (
	numPoints     = 100
	changeTsEvery = 5
	step          = 1000
)

func initLog(t *testing.T, tl *MustTimestampedLog) {
	data := make([][]byte, numPoints)

	for i := 0; i < numPoints; i++ {
		data[i] = util.Uint32To4Bytes(uint32(rand.Int31()))
	}
	timeStart := time.Now().UnixNano()
	nowis := timeStart

	var latest int64
	for i, d := range data {
		tl.Append(nowis, d)
		latest = nowis
		if (i+1)%changeTsEvery == 0 {
			nowis += step
		}
	}
	assert.EqualValues(t, timeStart, tl.Earliest())
	assert.EqualValues(t, latest, tl.Latest())
}

func TestTlogBig(t *testing.T) {
	vars := dict.New()
	tl := NewMustTimestampedLog(vars, "testTimestampedlog")

	initLog(t, tl)

	assert.EqualValues(t, numPoints, tl.Len())

	latest := tl.Latest()
	assert.Panics(t, func() {
		tl.Append(latest-10000, nil)
	})

	tslice := tl.TakeTimeSlice(tl.Earliest(), tl.Earliest())
	assert.EqualValues(t, changeTsEvery, tslice.NumPoints())

	tslice = tl.TakeTimeSlice(tl.Earliest(), tl.Earliest()+step)
	assert.EqualValues(t, 2*changeTsEvery, tslice.NumPoints())

	tslice = tl.TakeTimeSlice(tl.Latest(), tl.Latest())
	assert.EqualValues(t, changeTsEvery, tslice.NumPoints())

	tslice = tl.TakeTimeSlice(tl.Earliest(), tl.Latest())
	assert.EqualValues(t, tl.Len(), tslice.NumPoints())
	assert.EqualValues(t, tl.Len(), tslice.NumPoints())
}
