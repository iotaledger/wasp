package errors

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"golang.org/x/xerrors"
)

// StateErrorCollectionWriter implements ErrorCollection. Is used for contract internal errors.
// It requires a reference to a KVStore such as the vmctx and the hname of the caller.
type StateErrorCollectionWriter struct {
	partition kv.KVStore
	hname     iscp.Hname
}

func NewStateErrorCollectionWriter(partition kv.KVStore, hname iscp.Hname) coreerrors.ErrorCollection {
	errorCollection := StateErrorCollectionWriter{
		partition: partition,
		hname:     hname,
	}

	return &errorCollection
}

func (e *StateErrorCollectionWriter) getErrorDefinitionMap() *collections.Map {
	mapName := ParamErrorDefinitionMap + string(e.hname.Bytes())

	return collections.NewMap(e.partition, mapName)
}

func (e *StateErrorCollectionWriter) Get(errorId uint16) (*iscp.VMErrorTemplate, error) {
	errorMap := e.getErrorDefinitionMap()
	errorIdKey := codec.EncodeUint16(errorId)

	errorBytes, err := errorMap.GetAt(errorIdKey)

	if err != nil {
		return nil, err
	}

	errorDefinition, err := iscp.VMErrorTemplateFromMarshalUtil(marshalutil.New(errorBytes))

	if err != nil {
		return nil, err
	}

	return errorDefinition, nil
}

func (e *StateErrorCollectionWriter) Register(messageFormat string) (*iscp.VMErrorTemplate, error) {
	errorMap := e.getErrorDefinitionMap()
	errorId := iscp.GetErrorIdFromMessageFormat(messageFormat)

	if len(messageFormat) > iscp.VMErrorMessageLimit {
		return nil, coreerrors.ErrErrorMessageTooLong
	}

	mapKey := codec.EncodeUint16(errorId)

	if errorBytes, err := errorMap.GetAt(mapKey); err != nil {
		return nil, err
	} else if len(errorBytes) > 0 {
		return nil, coreerrors.ErrErrorAlreadyRegistered.Create(errorId)
	}

	newError := iscp.NewVMErrorTemplate(e.hname, errorId, messageFormat)

	if err := errorMap.SetAt(mapKey, newError.Bytes()); err != nil {
		return nil, err
	}

	return newError, nil
}

// StateErrorCollectionReader implements ErrorCollection partially. Is used for contract internal error readings only.
// It requires a reference to a KVStoreReader such as the vmctx and the hname of the caller.
type StateErrorCollectionReader struct {
	partition kv.KVStoreReader
	hname     iscp.Hname
}

func (e *StateErrorCollectionReader) getErrorDefinitionMap() *collections.ImmutableMap {
	mapName := ParamErrorDefinitionMap + string(e.hname.Bytes())

	return collections.NewMapReadOnly(e.partition, mapName)
}

func NewStateErrorCollectionReader(partition kv.KVStoreReader, hname iscp.Hname) coreerrors.ErrorCollection {
	errorCollection := StateErrorCollectionReader{
		partition: partition,
		hname:     hname,
	}

	return &errorCollection
}

func (e *StateErrorCollectionReader) Get(errorId uint16) (*iscp.VMErrorTemplate, error) {
	errorMap := e.getErrorDefinitionMap()
	errorIdKey := codec.EncodeUint16(errorId)

	errorBytes, err := errorMap.GetAt(errorIdKey)

	if err != nil {
		return nil, err
	}

	errorDefinition, err := iscp.VMErrorTemplateFromMarshalUtil(marshalutil.New(errorBytes))

	if err != nil {
		return nil, err
	}

	return errorDefinition, nil
}

func (e *StateErrorCollectionReader) Register(messageFormat string) (*iscp.VMErrorTemplate, error) {
	return nil, xerrors.Errorf("Registering in read only maps is unsupported")
}
