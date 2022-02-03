package errors

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

// StateErrorCollection implements IErrorCollection. Is used for contract internal errors.
// It requires a reference to a KVStore such as the vmctx and the hname of the caller.
type StateErrorCollection struct {
	partition kv.KVStore
	hname     iscp.Hname
}

func NewStateErrorCollection(partition kv.KVStore, hname iscp.Hname) IErrorCollection {
	errorCollection := StateErrorCollection{
		partition: partition,
		hname:     hname,
	}

	return &errorCollection
}

func (e *StateErrorCollection) getErrorDefinitionMap() *collections.Map {
	mapName := ParamErrorDefinitionMap + string(e.hname.Bytes())

	return collections.NewMap(e.partition, mapName)
}

func (e *StateErrorCollection) Get(errorId uint16) (*ErrorDefinition, error) {
	errorMap := e.getErrorDefinitionMap()
	errorIdKey := codec.EncodeUint16(errorId)

	errorBytes, err := errorMap.GetAt(errorIdKey)

	if err != nil {
		return nil, err
	}

	errorDefinition, err := ErrorDefinitionFromMarshalUtil(marshalutil.New(errorBytes))

	if err != nil {
		return nil, err
	}

	return errorDefinition, nil
}

func (e *StateErrorCollection) Register(errorId uint16, messageFormat string) (*ErrorDefinition, error) {
	errorMap := e.getErrorDefinitionMap()
	mapKey := codec.EncodeUint16(errorId)

	if errorBytes, err := errorMap.GetAt(mapKey); err != nil {
		return nil, err
	} else if len(errorBytes) > 0 {
		return nil, xerrors.Errorf("Error already registered")
	}

	newError := ErrorDefinition{id: errorId, messageFormat: messageFormat}

	if err := errorMap.SetAt(mapKey, newError.Bytes()); err != nil {
		return nil, err
	}

	return &newError, nil
}
