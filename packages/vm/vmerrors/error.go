package vmerrors

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
	"hash/crc32"
)

type ErrorDefinition struct {
	prefixId      uint32
	id            uint16
	messageFormat string
}

func NewErrorDefinition(prefixId uint32, errorId uint16, messageFormat string) *ErrorDefinition {
	return &ErrorDefinition{prefixId: prefixId, id: errorId, messageFormat: messageFormat}
}

func (e *ErrorDefinition) Error() string {
	return e.messageFormat
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

func (e *ErrorDefinition) Create(params ...interface{}) error {
	return &Error{
		PrefixId:      e.PrefixId(),
		Id:            e.Id(),
		MessageFormat: e.MessageFormat(),
		Params:        params,
	}
}

func (e *ErrorDefinition) CreateTyped(params ...interface{}) *Error {
	err := e.Create(params...)
	return err.(*Error)
}

func (e *ErrorDefinition) Panic(params ...interface{}) {
	panic(e.Create(params...))
}

func (e *ErrorDefinition) Serialize(mu *marshalutil.MarshalUtil) {
	messageFormatBytes := []byte(e.MessageFormat())

	mu.WriteUint32(e.PrefixId()).
		WriteUint16(e.Id()).
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

	if e.prefixId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

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

type Error struct {
	PrefixId      uint32
	Id            uint16
	MessageFormat string
	Params        []interface{}
}

func (e *Error) RequestMessageFormat(resolver ErrorMessageResolver) error {
	var err error

	if e.MessageFormat, err = resolver(e); err != nil {
		return err
	}

	return nil
}

func (e *Error) HasMessageFormat() bool {
	return e.MessageFormat != ""
}

func (e *Error) Error() string {
	if !e.HasMessageFormat() {
		return ""
	}

	return fmt.Sprintf(e.MessageFormat, e.Params...)
}

func (e *Error) Definition() ErrorDefinition {
	return ErrorDefinition{prefixId: e.PrefixId, id: e.Id, messageFormat: e.MessageFormat}
}

func (e *Error) Hash() uint32 {
	if !e.HasMessageFormat() {
		return 0
	}

	hash := crc32.Checksum([]byte(e.Error()), crc32.IEEETable)
	return hash
}

func (e *Error) serializeParams(mu *marshalutil.MarshalUtil) error {
	bytes, err := json.Marshal(e.Params)

	mu.WriteUint16(uint16(len(bytes)))
	mu.WriteBytes(bytes)

	return err
}

func (e *Error) Serialize(mu *marshalutil.MarshalUtil) error {
	hash := e.Hash()

	mu.WriteUint32(e.PrefixId).
		WriteUint16(e.Id).
		WriteUint32(hash)

	// For now, JSON encoded.
	err := e.serializeParams(mu)

	return err
}

func (e *Error) deserializeParams(mu *marshalutil.MarshalUtil) error {
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

type ErrorMessageResolver func(*Error) (string, error)

func ErrorFromBytes(mu *marshalutil.MarshalUtil, errorMessageResolver ErrorMessageResolver) (*Error, error) {
	var err error
	var hash uint32

	blockError := Error{}

	if blockError.PrefixId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	if blockError.Id, err = mu.ReadUint16(); err != nil {
		return nil, err
	}

	if hash, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	if err = blockError.deserializeParams(mu); err != nil {
		return nil, err
	}

	if errorMessageResolver == nil {
		return &blockError, nil
	}

	if blockError.MessageFormat, err = errorMessageResolver(&blockError); err != nil {
		return nil, err
	}

	// If message format is resolved, do hash check

	if hash == blockError.Hash() {
		return nil, xerrors.Errorf("Hash of error %v:%v does not match serialized error! (old: %v, now: %v)", blockError.PrefixId, blockError.Id, hash, blockError.Hash())
	}

	return &blockError, nil
}

func GetErrorIdFromMessageFormat(messageFormat string) uint16 {
	messageFormatHash := hashing.HashStrings(messageFormat).Bytes()
	mu := marshalutil.New(messageFormatHash)

	errorId, err := mu.ReadUint16()

	if err != nil {
		panic(err)
	}

	return errorId
}

func IsDefinition(error Error, definition ErrorDefinition) bool {
	if error.Id == definition.Id() && error.PrefixId == definition.PrefixId() {
		return true
	}

	return false
}

func Is(error *Error, errorComp *Error) bool {
	if error.Id == errorComp.Id &&
		error.PrefixId == errorComp.PrefixId {
		return true
	}

	return false
}

func Panic(definition ErrorDefinition, params ...interface{}) {
	panic(definition.Create(params...))
}
