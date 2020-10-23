package statequery

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
)

type Request struct {
	QueryGeneralData bool
	KeyQueries       []*KeyQuery
}

type Results struct {
	KeyQueryResults []*QueryResult
	byKey           map[kv.Key]*QueryResult

	// returned only when QueryGeneralData = true
	StateIndex uint32
	Timestamp  time.Time
	StateHash  *hashing.HashValue
	StateTxId  *jsonable.ValueTxID
	Requests   []*coretypes.RequestID
}

type KeyQuery struct {
	Key    []byte
	Type   ValueType
	Params json.RawMessage // one of DictQueryParams, ArrayQueryParams, ...
}

type ValueType string

const (
	ValueTypeScalar        = ValueType("scalar")
	ValueTypeArray         = ValueType("array")
	ValueTypeDict          = ValueType("dict")
	ValueTypeDictElement   = ValueType("dict-elem")
	ValueTypeTLogSlice     = ValueType("tlog_slice")
	ValueTypeTLogSliceData = ValueType("tlog_slice_data")
)

// TLogSliceQueryParams request slice of timestamped log. If FromTs or ToTs is 0,
// it is considered 'earliest' and 'latest' respectively.
// For 0,0 slice corresponds to the whole log
type TLogSliceQueryParams struct {
	FromTs int64
	ToTs   int64
}

// TLogSliceDataQueryParams request data for the slice
type TLogSliceDataQueryParams struct {
	FromIndex  uint32
	ToIndex    uint32
	Descending bool
}

type DictQueryParams struct {
	Limit uint32
}

type DictElementQueryParams struct {
	Key []byte
}

type ArrayQueryParams struct {
	From uint16
	To   uint16
}

type QueryResult struct {
	Key   []byte
	Type  ValueType
	Value json.RawMessage // one of []byte, DictResult, ArrayResult, ...
}

type KeyValuePair struct {
	Key   []byte
	Value []byte
}

type DictResult struct {
	Len     uint32
	Entries []KeyValuePair
}

type DictElementResult struct {
	Value []byte
}

type ArrayResult struct {
	Len    uint16
	Values [][]byte
}

type TLogSliceResult struct {
	IsNotEmpty bool
	FirstIndex uint32
	LastIndex  uint32
	Earliest   int64
	Latest     int64
}

type TLogSliceDataResult struct {
	Values [][]byte
}

func NewRequest() *Request {
	return &Request{}
}

func (q *Request) AddGeneralData() {
	q.QueryGeneralData = true
}

func (q *Request) AddScalar(key kv.Key) {
	q.KeyQueries = append(q.KeyQueries, &KeyQuery{
		Key:    []byte(key),
		Type:   ValueTypeScalar,
		Params: nil,
	})
}

func (q *Request) AddArray(key kv.Key, from uint16, to uint16) {
	p := &ArrayQueryParams{From: from, To: to}
	params, _ := json.Marshal(p)
	q.KeyQueries = append(q.KeyQueries, &KeyQuery{
		Key:    []byte(key),
		Type:   ValueTypeArray,
		Params: json.RawMessage(params),
	})
}

func (q *Request) AddDictionary(key kv.Key, limit uint32) {
	p := &DictQueryParams{Limit: limit}
	params, _ := json.Marshal(p)
	q.KeyQueries = append(q.KeyQueries, &KeyQuery{
		Key:    []byte(key),
		Type:   ValueTypeDict,
		Params: json.RawMessage(params),
	})
}

func (q *Request) AddDictionaryElement(dictKey kv.Key, elemKey []byte) {
	p := &DictElementQueryParams{Key: elemKey}
	params, _ := json.Marshal(p)
	q.KeyQueries = append(q.KeyQueries, &KeyQuery{
		Key:    []byte(dictKey),
		Type:   ValueTypeDictElement,
		Params: json.RawMessage(params),
	})
}

func (q *Request) AddTLogSlice(key kv.Key, fromTs, toTs int64) {
	p := TLogSliceQueryParams{
		FromTs: fromTs,
		ToTs:   toTs,
	}
	params, _ := json.Marshal(p)
	q.KeyQueries = append(q.KeyQueries, &KeyQuery{
		Key:    []byte(key),
		Type:   ValueTypeTLogSlice,
		Params: json.RawMessage(params),
	})
}

func (q *Request) AddTLogSliceData(key kv.Key, fromIndex, toIndex uint32, descending bool) {
	p := TLogSliceDataQueryParams{
		FromIndex:  fromIndex,
		ToIndex:    toIndex,
		Descending: descending,
	}
	params, _ := json.Marshal(p)
	q.KeyQueries = append(q.KeyQueries, &KeyQuery{
		Key:    []byte(key),
		Type:   ValueTypeTLogSliceData,
		Params: json.RawMessage(params),
	})
}

func (r *QueryResult) MustBytes() []byte {
	var b []byte
	err := json.Unmarshal(r.Value, &b)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *QueryResult) MustInt64() (int64, bool) {
	b := r.MustBytes()
	if b == nil {
		return 0, false
	}
	n, err := kv.DecodeInt64(b)
	if err != nil {
		panic(err)
	}
	return n, true
}

func (r *QueryResult) MustString() (string, bool) {
	b := r.MustBytes()
	if b == nil {
		return "", false
	}
	return string(b), true
}

func (r *QueryResult) MustAddress() *address.Address {
	b := r.MustBytes()
	if b == nil {
		return nil
	}
	addr, _, err := address.FromBytes(b)
	if err != nil {
		panic(err)
	}
	return &addr
}

func (r *QueryResult) MustHashValue() *hashing.HashValue {
	b := r.MustBytes()
	if b == nil {
		return nil
	}
	h, err := hashing.HashValueFromBytes(b)
	if err != nil {
		panic(err)
	}
	return &h
}

func (r *QueryResult) MustArrayResult() *ArrayResult {
	var ar ArrayResult
	err := json.Unmarshal(r.Value, &ar)
	if err != nil {
		panic(err)
	}
	return &ar
}

func (r *QueryResult) MustDictionaryResult() *DictResult {
	var dr DictResult
	err := json.Unmarshal(r.Value, &dr)
	if err != nil {
		panic(err)
	}
	return &dr
}

func (r *QueryResult) MustDictionaryElementResult() []byte {
	var dr *DictElementResult
	err := json.Unmarshal(r.Value, &dr)
	if err != nil {
		panic(err)
	}
	if dr == nil {
		return nil
	}
	return dr.Value
}

func (r *QueryResult) MustTLogSliceResult() *TLogSliceResult {
	var sr TLogSliceResult
	err := json.Unmarshal(r.Value, &sr)
	if err != nil {
		panic(err) // TODO panicing on wrong external data?
	}
	return &sr
}

func (r *QueryResult) MustTLogSliceDataResult() *TLogSliceDataResult {
	var sr TLogSliceDataResult
	err := json.Unmarshal(r.Value, &sr)
	if err != nil {
		panic(err)
	}
	return &sr
}

func (r *Results) Get(key kv.Key) *QueryResult {
	if r.byKey == nil {
		r.byKey = make(map[kv.Key]*QueryResult)
		for _, qr := range r.KeyQueryResults {
			r.byKey[kv.Key(qr.Key)] = qr
		}
	}
	return r.byKey[key]
}

func (q *KeyQuery) Execute(vars kv.BufferedKVStore) (*QueryResult, error) {
	key := kv.Key(q.Key)
	switch q.Type {
	case ValueTypeScalar:
		value, err := vars.Get(key)
		if err != nil {
			return nil, err
		}
		return q.makeResult(value)

	case ValueTypeArray:
		var params ArrayQueryParams
		err := json.Unmarshal(q.Params, &params)
		if err != nil {
			return nil, err
		}

		arr, err := vars.Codec().GetArray(key)
		if err != nil {
			return nil, err
		}

		size := arr.Len()
		values := make([][]byte, 0)
		for i := params.From; i < size && i < params.To; i++ {
			v, err := arr.GetAt(i)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return q.makeResult(ArrayResult{Len: size, Values: values})

	case ValueTypeDict:
		var params DictQueryParams
		err := json.Unmarshal(q.Params, &params)
		if err != nil {
			return nil, err
		}

		dict, err := vars.Codec().GetDictionary(key)
		if err != nil {
			return nil, err
		}

		entries := make([]KeyValuePair, 0)
		err = dict.Iterate(func(elemKey []byte, value []byte) bool {
			entries = append(entries, KeyValuePair{Key: elemKey, Value: value})
			return len(entries) < int(params.Limit)
		})
		if err != nil {
			return nil, err
		}
		return q.makeResult(DictResult{Len: dict.Len(), Entries: entries})

	case ValueTypeDictElement:
		var params DictElementQueryParams
		err := json.Unmarshal(q.Params, &params)
		if err != nil {
			return nil, err
		}

		dict, err := vars.Codec().GetDictionary(key)
		if err != nil {
			return nil, err
		}

		v, err := dict.GetAt(params.Key)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, nil
		}
		return q.makeResult(DictElementResult{Value: v})

	case ValueTypeTLogSlice:
		var params TLogSliceQueryParams
		err := json.Unmarshal(q.Params, &params)
		if err != nil {
			return nil, err
		}

		tlog, err := vars.Codec().GetTimestampedLog(key)
		if err != nil {
			return nil, err
		}

		tsl, err := tlog.TakeTimeSlice(params.FromTs, params.ToTs)
		if err != nil {
			return nil, err
		}
		if tsl.IsEmpty() {
			return q.makeResult(TLogSliceResult{})
		}
		ret := TLogSliceResult{
			IsNotEmpty: true,
			Earliest:   tsl.Earliest(),
			Latest:     tsl.Latest(),
		}
		ret.FirstIndex, ret.LastIndex = tsl.FromToIndices()
		return q.makeResult(ret)

	case ValueTypeTLogSliceData:
		var params TLogSliceDataQueryParams
		err := json.Unmarshal(q.Params, &params)
		if err != nil {
			return nil, err
		}

		tlog, err := vars.Codec().GetTimestampedLog(key)
		if err != nil {
			return nil, err
		}

		ret := TLogSliceDataResult{}
		ret.Values, err = tlog.LoadRecordsRaw(params.FromIndex, params.ToIndex, params.Descending)
		if err != nil {
			return nil, err
		}
		return q.makeResult(ret)
	}

	return nil, fmt.Errorf("No handler for type %s", q.Type)
}

func (q *KeyQuery) makeResult(value interface{}) (*QueryResult, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &QueryResult{
		Key:   q.Key,
		Type:  q.Type,
		Value: json.RawMessage(b),
	}, nil
}
