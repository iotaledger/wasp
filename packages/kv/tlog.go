package kv

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
)

type LogRecord struct {
	Index     uint32
	Timestamp int64
	Data      []byte
}

type TimeSlice struct {
	tslog    *TimestampedLog
	firstIdx uint32
	lastIdx  uint32
	earliest int64
	latest   int64
}

type TimestampedLog struct {
	kv             KVStore
	name           string
	cachedLen      uint32
	cachedLatest   int64
	cachedEarliest int64
}

func newTimestampedLog(kv KVStore, name string) (*TimestampedLog, error) {
	ret := &TimestampedLog{
		kv:   kv,
		name: name,
	}
	var err error
	if ret.cachedLen, err = ret.len(); err != nil {
		return nil, err
	}
	if ret.cachedLatest, err = ret.latest(); err != nil {
		return nil, err
	}
	if ret.cachedEarliest, err = ret.earliest(); err != nil {
		return nil, err
	}
	return ret, nil
}

const (
	tslSizeKeyCode = byte(0)
	tslElemKeyCode = byte(1)
)

func (l *TimestampedLog) getSizeKey() Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(tslSizeKeyCode)
	return Key(buf.Bytes())
}

func (l *TimestampedLog) getElemKey(idx uint32) Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(tslElemKeyCode)
	_ = util.WriteUint32(&buf, idx)
	return Key(buf.Bytes())
}

func (l *TimestampedLog) setSize(size uint32) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
	} else {
		l.kv.Set(l.getSizeKey(), util.Uint32To4Bytes(size))
	}
	l.cachedLen = size
}

func (l *TimestampedLog) len() (uint32, error) {
	v, err := l.kv.Get(l.getSizeKey())
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	if len(v) != 4 {
		return 0, errors.New("corrupted data")
	}
	return util.Uint32From4Bytes(v), nil
}

// Len == 0/empty/non-existent are equivalent
func (l *TimestampedLog) Len() uint32 {
	return l.cachedLen
}

func (l *TimestampedLog) Append(ts int64, data []byte) error {
	if ts < l.cachedLatest {
		return errors.New("wrong timestamp")
	}
	idx := l.cachedLen

	var buf bytes.Buffer
	buf.Write(util.Uint64To8Bytes(uint64(ts)))
	buf.Write(data)
	l.kv.Set(l.getElemKey(idx), buf.Bytes())
	l.setSize(idx + 1)
	l.cachedLatest = ts
	if idx == 0 {
		l.cachedEarliest = ts
	}
	return nil
}

func (l *TimestampedLog) Latest() int64 {
	return l.cachedLatest
}

func (l *TimestampedLog) latest() (int64, error) {
	idx := l.Len()
	if idx == 0 {
		return 0, nil
	}
	data, err := l.kv.Get(l.getElemKey(idx - 1))
	if err != nil {
		return 0, err
	}
	if len(data) < 8 {
		return 0, errors.New("corrupted data")
	}
	return int64(util.Uint64From8Bytes(data[:8])), nil
}

func (l *TimestampedLog) Earliest() int64 {
	return l.cachedEarliest
}

func (l *TimestampedLog) earliest() (int64, error) {
	if l.Len() == 0 {
		return 0, nil
	}
	data, err := l.kv.Get(l.getElemKey(0))
	if err != nil {
		return 0, err
	}
	if len(data) < 8 {
		return 0, errors.New("corrupted data")
	}
	return int64(util.Uint64From8Bytes(data[:8])), nil
}

func (l *TimestampedLog) getRecordAtIndex(idx uint32) (*LogRecord, error) {
	if idx >= l.cachedLen {
		return nil, nil
	}
	v, err := l.kv.Get(l.getElemKey(idx))
	if err != nil {
		return nil, err
	}
	if len(v) < 8 {
		return nil, errors.New("corrupted data")
	}
	return &LogRecord{
		Index:     idx,
		Timestamp: int64(util.Uint64From8Bytes(v[:8])),
		Data:      v[8:],
	}, nil
}

// binary search. Return 2 indices, i1 < i2, where [i1:i2] (i2 not including) contains all
// records with timestamp from 'fromTs' to 'toTs' (inclusive).
func (l *TimestampedLog) TakeTimeSlice(fromTs, toTs int64) (*TimeSlice, error) {
	if l.Len() == 0 {
		return nil, nil
	}
	if fromTs > toTs {
		return nil, nil
	}
	lowerIdx, ok, err := l.findLowerIdx(fromTs, 0, l.Len()-1)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	upperIdx, ok, err := l.findUpperIdx(toTs, 0, l.Len()-1)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	if lowerIdx > upperIdx {
		return nil, nil
	}
	earliest, err := l.getRecordAtIndex(lowerIdx)
	if err != nil {
		return nil, err
	}
	latest, err := l.getRecordAtIndex(upperIdx)
	if err != nil {
		return nil, err
	}

	return &TimeSlice{
		tslog:    l,
		firstIdx: lowerIdx,
		lastIdx:  upperIdx,
		earliest: earliest.Timestamp,
		latest:   latest.Timestamp,
	}, nil
}

func (l *TimestampedLog) findLowerIdx(ts int64, fromIdx, toIdx uint32) (uint32, bool, error) {
	if fromIdx > toIdx {
		return 0, false, nil
	}
	if fromIdx >= l.Len() || toIdx >= l.Len() {
		return 0, false, errors.New("wrong arguments")
	}
	r, err := l.getRecordAtIndex(fromIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		panic("internal error 1: r == nil")
	}
	lowerTs := r.Timestamp
	switch {
	case ts <= lowerTs:
		return fromIdx, true, nil
	case fromIdx == toIdx:
		return 0, false, nil
	}
	if !(ts > lowerTs && fromIdx < toIdx) {
		panic("assertion failed: ts > lowerTs && fromIdx < toIdx")
	}
	r, err = l.getRecordAtIndex(toIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		panic("internal error 1: r == nil")
	}
	upperTs := r.Timestamp
	if ts > upperTs {
		return 0, false, nil
	}
	// lowerTs < ts <= upperTs && fromIdx < toIdx
	if fromIdx+1 == toIdx {
		return toIdx, true, nil
	}
	// index is somewhere in between two different
	middleIdx := (fromIdx + toIdx) / 2

	ret, ok, err := l.findLowerIdx(ts, fromIdx, middleIdx)
	if err != nil {
		return 0, false, err
	}
	if ok {
		return ret, true, nil
	}
	return l.findLowerIdx(ts, middleIdx, toIdx)
}

func (l *TimestampedLog) findUpperIdx(ts int64, fromIdx, toIdx uint32) (uint32, bool, error) {
	if fromIdx > toIdx {
		return 0, false, nil
	}
	if fromIdx >= l.Len() || toIdx >= l.Len() {
		panic("fromIdx >= l.Len() || toIdx >= l.Len()")
	}
	r, err := l.getRecordAtIndex(toIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		return 0, false, fmt.Errorf("missing index %d", toIdx)
	}
	upperTs := r.Timestamp
	switch {
	case ts >= upperTs:
		return toIdx, true, nil
	case fromIdx == toIdx:
		return 0, false, nil

	}
	if !(ts < upperTs && fromIdx < toIdx) {
		panic("internal error: ts < upperTs && fromIdx < toIdx")
	}
	r, err = l.getRecordAtIndex(fromIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		return 0, false, fmt.Errorf("missing index %d", fromIdx)
	}
	lowerTs := r.Timestamp
	if ts < lowerTs {
		return 0, false, nil
	}
	if fromIdx+1 == toIdx {
		return fromIdx, true, nil
	}
	// lowerTs <= ts < upperTs && fromIdx < toIdx
	// index is somewhere in between two different
	middleIdx := (fromIdx + toIdx) / 2

	ret, ok, err := l.findUpperIdx(ts, middleIdx, toIdx)
	if err != nil {
		return 0, false, err
	}
	if ok {
		return ret, true, nil
	}
	return l.findUpperIdx(ts, fromIdx, middleIdx)
}

// TODO not finished

func (l *TimestampedLog) Erase() {
	panic("implement me")
}

func (sl *TimeSlice) IsEmpty() bool {
	return sl == nil || sl.firstIdx > sl.lastIdx
}

func (sl *TimeSlice) NumPoints() uint32 {
	if sl.IsEmpty() {
		return 0
	}
	return sl.lastIdx - sl.firstIdx + 1
}

func (sl *TimeSlice) Earliest() int64 {
	return sl.earliest
}

func (sl *TimeSlice) Latest() int64 {
	if sl.IsEmpty() {
		return 0
	}
	return sl.latest
}

func (sl *TimeSlice) LoadSlice() ([]*LogRecord, error) {
	ret := make([]*LogRecord, 0, sl.NumPoints())
	for i := sl.firstIdx; i <= sl.lastIdx; i++ {
		r, err := sl.tslog.getRecordAtIndex(i)
		if err != nil {
			return nil, err
		}
		ret = append(ret, r)
	}
	return ret, nil
}
