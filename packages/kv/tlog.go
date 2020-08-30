package kv

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
)

// TimestampedLog implement limitless append-only array of records where each record
// is indexed sequentially and consistently timestamped
// sequence of timestamps is considered consistent if for any indices i<j, Ti<=Tj,
// i.e. non-decreasing
type TimestampedLog struct {
	kv             KVStore
	name           Key
	cachedLen      uint32
	cachedLatest   int64
	cachedEarliest int64
}

type MustTimestampedLog struct {
	tlog TimestampedLog
}

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

func newTimestampedLog(kv KVStore, name Key) (*TimestampedLog, error) {
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

func newMustTimestampedLog(tlog *TimestampedLog) *MustTimestampedLog {
	return &MustTimestampedLog{*tlog}
}

const (
	tslSizeKeyCode = byte(iota)
	tslElemKeyCode
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
		panic("implement me")
		// l.kv.Del(l.getSizeKey())
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

func (l *MustTimestampedLog) Len() uint32 {
	return l.tlog.Len()
}

// Append appends data with timestamp to the end of the log.
// Returns error if timestamp is inconsistent, i.e. less than the latest timestamp
func (l *TimestampedLog) Append(ts int64, data []byte) error {
	if ts < l.cachedLatest {
		return errors.New("TimestampedLog.append: wrong timestamp")
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

func (l *MustTimestampedLog) Append(ts int64, data []byte) {
	err := l.tlog.Append(ts, data)
	if err != nil {
		panic(err)
	}
}

// Latest returns latest timestamp in the log
func (l *TimestampedLog) Latest() int64 {
	return l.cachedLatest
}

func (l *MustTimestampedLog) Latest() int64 {
	return l.tlog.Latest()
}

// latest loads latest timestamp from the DB
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
		return 0, errors.New("TimestampedLog: corrupted data")
	}
	return int64(util.Uint64From8Bytes(data[:8])), nil
}

// Earliest returns timestamp of the first record in the log, if any, or otherwise it is 0
func (l *TimestampedLog) Earliest() int64 {
	return l.cachedEarliest
}

func (l *MustTimestampedLog) Earliest() int64 {
	return l.tlog.Earliest()
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
		return 0, errors.New("TimestampedLog: corrupted data")
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
		return nil, errors.New("TimestampedLog: corrupted data")
	}
	return &LogRecord{
		Index:     idx,
		Timestamp: int64(util.Uint64From8Bytes(v[:8])),
		Data:      v[8:],
	}, nil
}

// TakeTimeSlice returns a slice structure, which contains existing indices
// firstIdx and lastIdx.
// Timestamps of all records between indices (inclusive) satisfy the condition >= T(firstIdx) and <=T(lastIdx)
// Any other pair of indices i1<fistId and/or i2>lastIdx does not satisfy the condition.
// In other words, returned slice contains all possible indices with timestamps between the two given
// Returned slice may be empty
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
		// empty slice
		return nil, nil
	}
	if lowerIdx > upperIdx {
		// empty slice
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

func (l *MustTimestampedLog) TakeTimeSlice(fromTs, toTs int64) *TimeSlice {
	tsl, err := l.tlog.TakeTimeSlice(fromTs, toTs)
	if err != nil {
		panic(err)
	}
	return tsl
}

func (l *TimestampedLog) findLowerIdx(ts int64, fromIdx, toIdx uint32) (uint32, bool, error) {
	if fromIdx > toIdx {
		return 0, false, nil
	}
	if fromIdx >= l.Len() || toIdx >= l.Len() {
		return 0, false, fmt.Errorf("TimestampedLog.findLowerIdx: wrong arguments: %d, %d, %d", ts, fromIdx, toIdx)
	}
	r, err := l.getRecordAtIndex(fromIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		panic(fmt.Errorf("TimestampedLog.findLowerIdx: r == nil: args: %d, %d, %d", ts, fromIdx, toIdx))
	}
	lowerTs := r.Timestamp
	switch {
	case ts <= lowerTs:
		return fromIdx, true, nil
	case fromIdx == toIdx:
		return 0, false, nil
	}
	if !(ts > lowerTs && fromIdx < toIdx) {
		panic(fmt.Errorf("TimestampedLog.findLowerIdx: assertion failed: ts > lowerTs && fromIdx < toIdx: args: %d, %d, %d", ts, fromIdx, toIdx))
	}
	r, err = l.getRecordAtIndex(toIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		panic(fmt.Errorf("TimestampedLog.findLowerIdx: assertion failed: r == nil: args: %d, %d, %d", ts, fromIdx, toIdx))
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
		panic(fmt.Errorf("TimestampedLog.findLowerIdx: assertion failed: fromIdx >= l.Len() || toIdx >= l.Len(): args: %d, %d, %d", ts, fromIdx, toIdx))
	}
	r, err := l.getRecordAtIndex(toIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		return 0, false, fmt.Errorf("inconsistency: missing data at index %d", toIdx)
	}
	upperTs := r.Timestamp
	switch {
	case upperTs <= ts:
		return toIdx, true, nil
	case fromIdx == toIdx:
		return 0, false, nil

	}
	if !(ts < upperTs && fromIdx < toIdx) {
		panic(fmt.Errorf("TimestampedLog.findUpperIdx: assertion failed: ts < upperTs && fromIdx < toIdx: args: %d, %d, %d", ts, fromIdx, toIdx))
	}
	r, err = l.getRecordAtIndex(fromIdx)
	if err != nil {
		return 0, false, err
	}
	if r == nil {
		return 0, false, fmt.Errorf("inconsistency: missing data at index %d", fromIdx)
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

// TODO not finished fith Erase

func (l *TimestampedLog) Erase() {
	panic("implement me")
}

// IsEmpty returns true if slice does not contains points
func (sl *TimeSlice) IsEmpty() bool {
	return sl == nil || sl.firstIdx > sl.lastIdx
}

// NumPoints return number of indices (records) in the slice
func (sl *TimeSlice) NumPoints() uint32 {
	if sl.IsEmpty() {
		return 0
	}
	return sl.lastIdx - sl.firstIdx + 1
}

// Earliest return timestamp of the first index or 0 if empty
func (sl *TimeSlice) Earliest() int64 {
	if sl.IsEmpty() {
		return 0
	}
	return sl.earliest
}

// Earliest returns timestamp of the last index or 0 if empty
func (sl *TimeSlice) Latest() int64 {
	if sl.IsEmpty() {
		return 0
	}
	if sl.IsEmpty() {
		return 0
	}
	return sl.latest
}

// LoadSlice returns all records in the slice
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
