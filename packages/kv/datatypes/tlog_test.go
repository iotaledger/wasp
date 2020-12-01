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
	tl, err := NewTimestampedLog(vars, "testTlog")
	assert.NoError(t, err)
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

	err = tl.Append(nowis, d1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, tl.Len())

	err = tl.Append(nowis, d2)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, tl.Len())

	err = tl.Append(nowisNext1, d3)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, tl.Len())

	err = tl.Append(nowis, d4)
	assert.Error(t, err)
	assert.EqualValues(t, 3, tl.Len())

	err = tl.Append(nowisNext1, d4)
	assert.NoError(t, err)
	assert.EqualValues(t, 4, tl.Len())

	err = tl.Append(nowisNext1, d5)
	assert.NoError(t, err)
	assert.EqualValues(t, 5, tl.Len())

	err = tl.Append(nowisNext2, d6)
	assert.NoError(t, err)
	assert.EqualValues(t, 6, tl.Len())

	assert.EqualValues(t, nowis, tl.Earliest())
	assert.EqualValues(t, nowisNext2, tl.Latest())

	err = tl.Append(nowisNext2, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, 7, tl.Len())
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

	var err error
	var latest int64
	for i, d := range data {
		err = tl.Append(nowis, d)
		latest = nowis
		assert.NoError(t, err)
		if (i+1)%changeTsEvery == 0 {
			nowis += step
		}
	}
	assert.EqualValues(t, timeStart, tl.Earliest())
	assert.EqualValues(t, latest, tl.Latest())
}

func TestTlogBig(t *testing.T) {
	vars := dict.New()
	tl, err := NewTimestampedLog(vars, "testTimestampedlog")
	assert.NoError(t, err)

	initLog(t, tl)

	assert.EqualValues(t, numPoints, tl.Len())

	latest := tl.Latest()
	err = tl.Append(latest-10000, nil)
	assert.Error(t, err)

	tslice, err := tl.TakeTimeSlice(tl.Earliest(), tl.Earliest())
	assert.NoError(t, err)
	assert.EqualValues(t, changeTsEvery, tslice.NumPoints())

	tslice, err = tl.TakeTimeSlice(tl.Earliest(), tl.Earliest()+step)
	assert.NoError(t, err)
	assert.EqualValues(t, 2*changeTsEvery, tslice.NumPoints())

	tslice, err = tl.TakeTimeSlice(tl.Latest(), tl.Latest())
	assert.NoError(t, err)
	assert.EqualValues(t, changeTsEvery, tslice.NumPoints())

	tslice, err = tl.TakeTimeSlice(tl.Earliest(), tl.Latest())
	assert.EqualValues(t, tl.Len(), tslice.NumPoints())
	assert.NoError(t, err)
	assert.EqualValues(t, tl.Len(), tslice.NumPoints())
}
