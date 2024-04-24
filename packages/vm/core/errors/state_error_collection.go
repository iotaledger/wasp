package errors

import (
	"errors"

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

func (e *StateErrorCollectionWriter) Get(errorID uint16) (*isc.VMErrorTemplate, bool) {
	errorMap := e.getErrorTemplateMap()
	errorIDKey := codec.Uint16.Encode(errorID)

	errorBytes := errorMap.GetAt(errorIDKey)
	if errorBytes == nil {
		return nil, false
	}

	template, err := isc.VMErrorTemplateFromBytes(errorBytes)
	if err != nil {
		panic(err)
	}

	return template, true
}

func (e *StateErrorCollectionWriter) Register(messageFormat string) (*isc.VMErrorTemplate, error) {
	errorMap := e.getErrorTemplateMap()
	errorID := isc.GetErrorIDFromMessageFormat(messageFormat)

	if len(messageFormat) > isc.VMErrorMessageLimit {
		return nil, coreerrors.ErrErrorMessageTooLong
	}

	if t, ok := e.Get(errorID); ok && messageFormat != t.MessageFormat() {
		return nil, coreerrors.ErrErrorTemplateConflict.Create(errorID)
	}

	newError := isc.NewVMErrorTemplate(isc.NewVMErrorCode(e.hname, errorID), messageFormat)

	errorMap.SetAt(codec.Uint16.Encode(errorID), newError.Bytes())

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

func (e *StateErrorCollectionReader) Get(errorID uint16) (*isc.VMErrorTemplate, bool) {
	errorMap := e.getErrorTemplateMap()
	errorIDKey := codec.Uint16.Encode(errorID)

	errorBytes := errorMap.GetAt(errorIDKey)
	if errorBytes == nil {
		return nil, false
	}

	template, err := isc.VMErrorTemplateFromBytes(errorBytes)
	if err != nil {
		panic(err)
	}

	return template, true
}

func (e *StateErrorCollectionReader) Register(messageFormat string) (*isc.VMErrorTemplate, error) {
	return nil, errors.New("registering in read only maps is unsupported")
}
