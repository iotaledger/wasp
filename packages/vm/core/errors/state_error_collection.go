package errors

import (
	"errors"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

func errorTemplateKey(contractID isc.Hname) string {
	return prefixErrorTemplateMap + string(contractID.Bytes())
}

// StateErrorCollectionWriter implements ErrorCollection. Is used for contract internal errors.
// It requires a reference to a KVStore such as the vmctx and the hname of the caller.
type StateErrorCollectionWriter struct {
	partition kv.KVStore
	hname     isc.Hname
}

func NewStateErrorCollectionWriter(partition kv.KVStore, hname isc.Hname) coreerrors.ErrorCollection {
	errorCollection := StateErrorCollectionWriter{
		partition: partition,
		hname:     hname,
	}

	return &errorCollection
}

func (e *StateErrorCollectionWriter) getErrorTemplateMap() *collections.Map {
	return collections.NewMap(e.partition, errorTemplateKey(e.hname))
}

func (e *StateErrorCollectionWriter) Get(errorID uint16) (*isc.VMErrorTemplate, error) {
	errorMap := e.getErrorTemplateMap()
	errorIDKey := codec.EncodeUint16(errorID)

	errorBytes, err := errorMap.GetAt(errorIDKey)
	if err != nil {
		return nil, err
	}

	template, err := isc.VMErrorTemplateFromMarshalUtil(marshalutil.New(errorBytes))
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (e *StateErrorCollectionWriter) Register(messageFormat string) (*isc.VMErrorTemplate, error) {
	errorMap := e.getErrorTemplateMap()
	errorID := isc.GetErrorIDFromMessageFormat(messageFormat)

	if len(messageFormat) > isc.VMErrorMessageLimit {
		return nil, coreerrors.ErrErrorMessageTooLong
	}

	mapKey := codec.EncodeUint16(errorID)

	if errorBytes, err := errorMap.GetAt(mapKey); err != nil {
		return nil, err
	} else if len(errorBytes) > 0 {
		return nil, coreerrors.ErrErrorAlreadyRegistered.Create(errorID)
	}

	newError := isc.NewVMErrorTemplate(isc.NewVMErrorCode(e.hname, errorID), messageFormat)

	if err := errorMap.SetAt(mapKey, newError.Bytes()); err != nil {
		return nil, err
	}

	return newError, nil
}

// StateErrorCollectionReader implements ErrorCollection partially. Is used for contract internal error readings only.
// It requires a reference to a KVStoreReader such as the vmctx and the hname of the caller.
type StateErrorCollectionReader struct {
	partition kv.KVStoreReader
	hname     isc.Hname
}

func (e *StateErrorCollectionReader) getErrorTemplateMap() *collections.ImmutableMap {
	return collections.NewMapReadOnly(e.partition, errorTemplateKey(e.hname))
}

func NewStateErrorCollectionReader(partition kv.KVStoreReader, hname isc.Hname) coreerrors.ErrorCollection {
	errorCollection := StateErrorCollectionReader{
		partition: partition,
		hname:     hname,
	}

	return &errorCollection
}

func (e *StateErrorCollectionReader) Get(errorID uint16) (*isc.VMErrorTemplate, error) {
	errorMap := e.getErrorTemplateMap()
	errorIDKey := codec.EncodeUint16(errorID)

	errorBytes, err := errorMap.GetAt(errorIDKey)
	if err != nil {
		return nil, err
	}

	template, err := isc.VMErrorTemplateFromMarshalUtil(marshalutil.New(errorBytes))
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (e *StateErrorCollectionReader) Register(messageFormat string) (*isc.VMErrorTemplate, error) {
	return nil, errors.New("registering in read only maps is unsupported")
}
