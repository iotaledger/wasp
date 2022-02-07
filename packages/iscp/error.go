package iscp

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"golang.org/x/xerrors"
	"hash/crc32"
)

type ErrorDefinition struct {
	prefixId      Hname
	id            uint16
	messageFormat string
}

func (e *ErrorDefinition) Error() string {
	return e.messageFormat
}

func (e *ErrorDefinition) PrefixId() Hname {
	return e.prefixId
}

func (e *ErrorDefinition) Id() uint16 {
	return e.id
}

func (e *ErrorDefinition) MessageFormat() string {
	return e.messageFormat
}

func (e *ErrorDefinition) Create(params ...interface{}) *Error {
	return &Error{
		PrefixId:      e.PrefixId(),
		Id:            e.Id(),
		MessageFormat: e.MessageFormat(),
		Params:        params,
	}
}

func (e *ErrorDefinition) Panic(params ...interface{}) Error {
	panic(*e.Create(params...))
}

func (e *ErrorDefinition) Serialize(mu *marshalutil.MarshalUtil) {
	messageFormatBytes := []byte(e.MessageFormat())

	mu.WriteUint32(uint32(e.PrefixId())).
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
	var prefixId uint32

	e := ErrorDefinition{}

	if prefixId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	e.prefixId = Hname(prefixId)

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
	PrefixId      Hname
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

	mu.WriteUint32(uint32(e.PrefixId)).
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
	var prefixId uint32
	var hash uint32

	blockError := Error{}

	if prefixId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	blockError.PrefixId = Hname(prefixId)

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
