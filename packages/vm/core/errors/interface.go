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

type IErrorCollection interface {
	Get(errorId uint16) (*ErrorDefinition, error)
	//Create(errorId uint16, params ...interface{}) (*BlockError, error)
	Register(errorId uint16, messageFormat string) (*ErrorDefinition, error)
}

type ErrorDefinition struct {
	id            uint16
	messageFormat string
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

type BlockError struct {
	Id            uint16
	MessageFormat string
	Params        []interface{}
}

func (e *BlockError) Message() string {
	return fmt.Sprintf(e.MessageFormat, e.Params...)
}

func (e *BlockError) Error() error {
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

	mu.WriteUint16(e.Id).
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
