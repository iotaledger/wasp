package errors

import (
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
	"hash/fnv"
)

type ErrorCollection interface {
	Get(errorId uint32) *ErrorDefinition
	Register(messageFormat string) *ErrorDefinition
	Create(errorId uint32, params ...interface{}) (*BlockError, error)
}

type NamedErrorCollection struct {
	Name   string
	errors map[uint32]ErrorDefinition
}

func NewNamedErrorCollection(name string) *NamedErrorCollection {
	return &NamedErrorCollection{
		Name:   name,
		errors: map[uint32]ErrorDefinition{},
	}
}

func (e *NamedErrorCollection) Get(errorId uint32) *ErrorDefinition {
	if errorDefinition, ok := e.errors[errorId]; ok {
		return &errorDefinition
	}

	return nil
}

func (e *NamedErrorCollection) Create(errorId uint32, params ...interface{}) (*BlockError, error) {
	if errorDefinition := e.Get(errorId); errorDefinition == nil {
		return nil, xerrors.Errorf("No error found with id %v", errorId)
	} else {
		errorInstance := errorDefinition.Create(params...)
		return &errorInstance, nil
	}
}

func (e *NamedErrorCollection) generateErrorId(message string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(message))
	return hash.Sum32()
}

func (e *NamedErrorCollection) Register(messageFormat string) *ErrorDefinition {

	prefix := iscp.Hn(e.Name)
	prefixId := binary.LittleEndian.Uint32(prefix.Bytes())
	id := e.generateErrorId(messageFormat)

	errorDefinition := ErrorDefinition{
		prefixId:      prefixId,
		id:            id,
		messageFormat: messageFormat,
	}

	return &errorDefinition
}
