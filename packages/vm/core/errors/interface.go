package errors

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"golang.org/x/xerrors"
	"hash/crc32"
)

var Contract = coreutil.NewContract(coreutil.CoreContractError, "Error contract")

const (
	prefixBlockRegistry = string('a' + iota)
	prefixControlAddresses
)

var (
	FuncRegisterError = coreutil.Func("registerError")
)

const (
	// parameters
	ParamErrorDefinitionMap   = "e"
	ParamErrorId              = "i"
	ParamErrorMessageFormat   = "m"
	ParamErrorDefinitionAdded = "a"
)

type ErrorDefinition struct {
	prefixId      uint32
	id            uint16
	messageFormat string
}

func (e *ErrorDefinition) PrefixId() uint32 {
	return e.prefixId
}

func (e *ErrorDefinition) Id() uint16 {
	return e.id
}

func (e *ErrorDefinition) MessageFormat() string {
	return e.messageFormat
}

func (e *ErrorDefinition) Create(params ...interface{}) BlockError {
	return BlockError{
		Id:            e.Id(),
		PrefixId:      e.PrefixId(),
		MessageFormat: e.MessageFormat(),
		Params:        params,
	}
}

func (e *ErrorDefinition) Serialize(mu *marshalutil.MarshalUtil) {
	messageFormatBytes := []byte(e.MessageFormat())

	mu.WriteUint16(e.Id()).
		WriteUint16(uint16(len(messageFormatBytes))).
		WriteBytes(messageFormatBytes)
}

func (e *ErrorDefinition) Bytes() []byte {
	mu := marshalutil.New()
	e.Serialize(mu)
	return mu.Bytes()
}

func ErrorDefinitionFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ErrorDefinition, error) {
	var err error
	var messageLength uint16

	e := ErrorDefinition{}

	if e.id, err = mu.ReadUint16(); err != nil {
		return nil, err
	}

	if messageLength, err = mu.ReadUint16(); err != nil {
		return nil, err
	}

	if messageInBytes, err := mu.ReadBytes(int(messageLength)); err != nil {
		return nil, err
	} else {
		e.messageFormat = string(messageInBytes)
	}

	return &e, nil
}

func ErrorDefinitionsFromMarshalUtil(mu *marshalutil.MarshalUtil) ([]*ErrorDefinition, error) {
	doneReading, err := mu.DoneReading()
	errorDefinitions := make([]*ErrorDefinition, 0)

	if err != nil {
		return nil, err
	} else if doneReading {
		return errorDefinitions, nil
	}

	for !doneReading {
		errorDefinition, err := ErrorDefinitionFromMarshalUtil(mu)

		if err != nil {
			return nil, err
		}

		errorDefinitions = append(errorDefinitions, errorDefinition)
	}

	return errorDefinitions, nil
}

type BlockError struct {
	PrefixId      uint32
	Id            uint16
	MessageFormat string
	Params        []interface{}
}

func (e *BlockError) Message() string {
	return fmt.Sprintf(e.MessageFormat, e.Params...)
}

func (e *BlockError) AsError() error {
	return xerrors.Errorf(e.MessageFormat, e.Params...)
}

func (e *BlockError) Hash() uint32 {
	hash := crc32.Checksum([]byte(e.Message()), crc32.IEEETable)
	return hash
}

func (e *BlockError) serializeParams(mu *marshalutil.MarshalUtil) error {
	bytes, err := json.Marshal(e.Params)

	mu.WriteUint32(uint32(len(bytes)))
	mu.WriteBytes(bytes)

	return err
}

func (e *BlockError) Serialize(mu *marshalutil.MarshalUtil) error {
	hash := e.Hash()

	mu.WriteUint32(e.PrefixId).
		WriteUint16(e.Id).
		WriteUint32(hash)

	// For now, JSON encoded.
	err := e.serializeParams(mu)

	return err
}

/*
func NewBlockErrorDefinition(collection ErrorCollection, messageFormat string) *ErrorDefinition {
	return collection.Register(messageFormat)
}*/

func (e *BlockError) deserializeParams(mu *marshalutil.MarshalUtil) error {
	var err error
	var paramLength uint16
	var params []byte

	if paramLength, err = mu.ReadUint16(); err != nil {
		return err
	}

	if params, err = mu.ReadBytes(int(paramLength)); err != nil {
		return err
	}

	if err = json.Unmarshal(params, &e.Params); err != nil {
		return err
	}

	return err
}
