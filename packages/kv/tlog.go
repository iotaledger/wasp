package kv

import (
	"bytes"
	"errors"
	"github.com/iotaledger/wasp/packages/util"
)

type TimestampedLog interface {
	Append(ts int64, data []byte) error
	Len() uint32
	Latest() int64
	GetTimeSlice(fromTs, toTs int64) (uint32, uint32)
	LoadSlice(fromIdx, toIdx uint32) []*LogRecord
	Erase()
}

type LogRecord struct {
	Index     uint32
	Timestamp int64
	Data      []byte
}

type tslStruct struct {
	kv           KVStore
	name         string
	cachedLen    uint32
	cachedLatest int64
}

func newTimestampedLog(kv KVStore, name string) TimestampedLog {
	ret := &tslStruct{
		kv:   kv,
		name: name,
	}
	ret.cachedLen = ret.len()
	ret.cachedLatest = ret.latest()
	return ret
}

const (
	tslSizeKeyCode = byte(0)
	tslElemKeyCode = byte(1)
)

func (l *tslStruct) getSizeKey() Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(tslSizeKeyCode)
	return Key(buf.Bytes())
}

func (l *tslStruct) getElemKey(idx uint32) Key {
	var buf bytes.Buffer
	buf.Write([]byte(l.name))
	buf.WriteByte(tslElemKeyCode)
	_ = util.WriteUint32(&buf, idx)
	return Key(buf.Bytes())
}

func (l *tslStruct) setSize(size uint32) {
	if size == 0 {
		l.kv.Del(l.getSizeKey())
	} else {
		l.kv.Set(l.getSizeKey(), util.Uint32To4Bytes(size))
	}
	l.cachedLen = size
}

func (l *tslStruct) len() uint32 {
	v, err := l.kv.Get(l.getSizeKey())
	if err != nil || len(v) != 4 {
		return 0
	}
	return util.Uint32From4Bytes(v)
}

// Len == 0/empty/non-existent are equivalent
func (l *tslStruct) Len() uint32 {
	return l.cachedLen
}

func (l *tslStruct) Append(ts int64, data []byte) error {
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
	return nil
}

func (l *tslStruct) Latest() int64 {
	return l.cachedLatest
}

func (l *tslStruct) latest() int64 {
	idx := l.Len()
	if idx == 0 {
		return 0
	}
	data, err := l.kv.Get(l.getElemKey(idx - 1))
	if err != nil {
		return 0
	}
	if len(data) < 8 {
		return 0
	}
	return int64(util.Uint64From8Bytes(data[:8]))
}

func (l *tslStruct) getTsAtIndex(idx uint32) int64 {
	if idx >= l.cachedLen {
		return 0
	}
	v, err := l.kv.Get(l.getElemKey(idx))
	if err != nil {
		return 0
	}
	if len(v) < 8 {
		return 0
	}
	return int64(util.Uint64From8Bytes(v[:8]))
}

// binary search
func (l *tslStruct) GetTimeSlice(fromTs, toTs int64) (uint32, uint32) {
	panic("implement me")
}

func (l *tslStruct) findLowerIdx(ts int64, fromIdx, toIdx uint32) (uint32, bool) {
	panic("implement me")
}

func (l *tslStruct) LoadSlice(fromIdx, toIdx uint32) []*LogRecord {
	panic("implement me")
}

func (l *tslStruct) Erase() {
	panic("implement me")
}
