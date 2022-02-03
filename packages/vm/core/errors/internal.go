package errors

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

func NewErrorDefinitionMap(partition kv.KVStore, hname iscp.Hname) *collections.Map {
	mapName := ParamErrorDefinitionMap + string(hname.Bytes())

	return collections.NewMap(partition, mapName)
}

func HasErrorDefinition(partition kv.KVStore, hname iscp.Hname, errorId uint16) (bool, error) {
	errorDefinition, err := GetErrorDefinition(partition, hname, errorId)

	if err != nil {
		return false, err
	}

	if errorDefinition != nil {
		return false, nil
	}

	return true, nil
}

/*
func GetErrorDefinitions(partition kv.KVStore, hname iscp.Hname) ([]*ErrorDefinition, error) {
	errorMap := NewErrorDefinitionMap(partition, hname)
	errorDefinitions := map[uint16]*ErrorDefinition{}

	iterator := func(key []byte, value []byte) bool {
		errorDefinition, _ := ErrorDefinitionFromMarshalUtil(marshalutil.New(value))
		errorId, _ := codec.DecodeUint16(key)

		errorDefinitions[errorId] = errorDefinition

		return true
	}

	errorMap.Iterate(iterator)

	if err != nil {
		return errorDefinitions, nil
	}

	return errorDefinitions, nil
}*/

func GetErrorDefinition(partition kv.KVStore, hname iscp.Hname, errorId uint16) (*ErrorDefinition, error) {
	errorMap := NewErrorDefinitionMap(partition, hname)
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

func AddErrorDefinition(partition kv.KVStore, hname iscp.Hname, id uint16, messageFormat string) (bool, error) {
	errorMap := NewErrorDefinitionMap(partition, hname)
	mapKey := codec.EncodeUint16(id)

	if errorBytes, err := errorMap.GetAt(mapKey); err != nil {
		return false, err
	} else if len(errorBytes) > 0 {
		return false, xerrors.Errorf("Error already registered")
	}

	newError := ErrorDefinition{id: id, messageFormat: messageFormat}

	if err := errorMap.SetAt(mapKey, newError.Bytes()); err != nil {
		return false, err
	}

	return true, nil
}
