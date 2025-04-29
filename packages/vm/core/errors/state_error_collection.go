// Package errors provides state management functionality for handling error templates
// and collections in smart contracts. It implements interfaces for both reading and
// writing error states, allowing contracts to register, store, and retrieve custom
// error templates with associated message formats. The package supports both read-only
// and write operations through StateErrorCollectionReader and StateErrorCollectionWriter
// implementations respectively.
package errors

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
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
	w     *StateWriter
	hname isc.Hname
}

func (s *StateWriter) ErrorCollection(hname isc.Hname) coreerrors.ErrorCollectionWriter {
	return &StateErrorCollectionWriter{
		w:     s,
		hname: hname,
	}
}

func (e *StateErrorCollectionWriter) getErrorTemplateMap() *collections.Map {
	return collections.NewMap(e.w.state, errorTemplateKey(e.hname))
}

func (e *StateErrorCollectionWriter) Get(errorID uint16) (*isc.VMErrorTemplate, bool) {
	errorMap := e.getErrorTemplateMap()
	errorIDKey := codec.Encode[uint16](errorID)

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

	errorMap.SetAt(codec.Encode(errorID), newError.Bytes())

	return newError, nil
}

// StateErrorCollectionReader implements ErrorCollection partially. Is used for contract internal error readings only.
// It requires a reference to a KVStoreReader such as the vmctx and the hname of the caller.
type StateErrorCollectionReader struct {
	r     *StateReader
	hname isc.Hname
}

func (e *StateErrorCollectionReader) getErrorTemplateMap() *collections.ImmutableMap {
	return collections.NewMapReadOnly(e.r.state, errorTemplateKey(e.hname))
}

func (s *StateReader) ErrorCollection(hname isc.Hname) coreerrors.ErrorCollection {
	return &StateErrorCollectionReader{
		r:     s,
		hname: hname,
	}
}

func (e *StateErrorCollectionReader) Get(errorID uint16) (*isc.VMErrorTemplate, bool) {
	errorMap := e.getErrorTemplateMap()
	errorIDKey := codec.Encode[uint16](errorID)

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
