package iscp

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
	"hash/crc32"
	"math"
)

const VMErrorMessageLimit = math.MaxUint16

type VMErrorTemplate struct {
	prefixId      Hname
	id            uint16
	messageFormat string
}

func NewVMErrorTemplate(prefixId Hname, errorId uint16, messageFormat string) *VMErrorTemplate {

	return &VMErrorTemplate{prefixId: prefixId, id: errorId, messageFormat: messageFormat}
}

func (e *VMErrorTemplate) Error() string {
	return e.messageFormat
}

func (e *VMErrorTemplate) PrefixId() Hname {
	return e.prefixId
}

func (e *VMErrorTemplate) Id() uint16 {
	return e.id
}

func (e *VMErrorTemplate) MessageFormat() string {
	return e.messageFormat
}

func (e *VMErrorTemplate) Create(params ...interface{}) *VMError {
	return &VMError{
		prefixId:      e.PrefixId(),
		id:            e.Id(),
		messageFormat: e.MessageFormat(),
		params:        params,
	}
}

func (e *VMErrorTemplate) Panic(params ...interface{}) {
	panic(e.Create(params...))
}

func (e *VMErrorTemplate) Serialize(mu *marshalutil.MarshalUtil) {
	messageFormatBytes := []byte(e.MessageFormat())

	mu.WriteUint32(uint32(e.PrefixId())).
		WriteUint16(e.Id()).
		WriteUint16(uint16(len(messageFormatBytes))).
		WriteBytes(messageFormatBytes)
}

func (e *VMErrorTemplate) Bytes() []byte {
	mu := marshalutil.New()
	e.Serialize(mu)
	return mu.Bytes()
}

func VMErrorTemplateFromMarshalUtil(mu *marshalutil.MarshalUtil) (*VMErrorTemplate, error) {
	var err error
	var messageLength uint16

	e := VMErrorTemplate{}

	var prefixId uint32

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

type VMError struct {
	prefixId      Hname
	id            uint16
	messageFormat string
	params        []interface{}
}

func (e *VMError) PrefixId() Hname {
	return e.prefixId
}

func (e *VMError) Id() uint16 {
	return e.id
}

func (e *VMError) MessageFormat() string {
	return e.messageFormat
}

func (e *VMError) Params() []interface{} {
	return e.params
}

func (e *VMError) RequestMessageFormat(resolver VMErrorMessageResolver) error {
	var err error

	if e.messageFormat, err = resolver(e); err != nil {
		return err
	}

	return nil
}

func (e *VMError) HasMessageFormat() bool {
	return e.MessageFormat() != ""
}

func (e *VMError) Error() string {
	if !e.HasMessageFormat() {
		return ""
	}

	return fmt.Sprintf(e.MessageFormat(), e.Params()...)
}

func (e *VMError) Definition() VMErrorTemplate {
	return VMErrorTemplate{prefixId: e.PrefixId(), id: e.Id(), messageFormat: e.MessageFormat()}
}

func (e *VMError) Hash() uint32 {
	if !e.HasMessageFormat() {
		return 0
	}

	hash := crc32.Checksum([]byte(e.Error()), crc32.IEEETable)
	return hash
}

func (e *VMError) serializeParams(mu *marshalutil.MarshalUtil) {
	bytes, err := json.Marshal(e.Params())

	if err != nil {
		panic(err)
	}

	mu.WriteUint16(uint16(len(bytes)))
	mu.WriteBytes(bytes)
}

func (e *VMError) Bytes() []byte {
	mu := marshalutil.New()
	hash := e.Hash()

	mu.WriteUint32(uint32(e.PrefixId())).
		WriteUint16(e.Id()).
		WriteUint32(hash)

	// For now, JSON encoded.
	e.serializeParams(mu)

	return mu.Bytes()
}

func (e *VMError) deserializeParams(mu *marshalutil.MarshalUtil) error {
	var err error
	var paramLength uint16
	var params []byte

	if paramLength, err = mu.ReadUint16(); err != nil {
		return err
	}

	if params, err = mu.ReadBytes(int(paramLength)); err != nil {
		return err
	}

	if err = json.Unmarshal(params, &e.params); err != nil {
		return err
	}

	return err
}

type VMErrorMessageResolver func(*VMError) (string, error)

func VMErrorFromMarshalUtil(mu *marshalutil.MarshalUtil, errorMessageResolver VMErrorMessageResolver) (*VMError, error) {
	var err error
	var hash uint32

	blockError := VMError{}

	var prefixId uint32
	if prefixId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	blockError.prefixId = Hname(prefixId)

	if blockError.id, err = mu.ReadUint16(); err != nil {
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

	if blockError.messageFormat, err = errorMessageResolver(&blockError); err != nil {
		return nil, err
	}

	// If message format is resolved, do hash check

	if hash != blockError.Hash() {
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

func IsDefinition(error VMError, definition VMErrorTemplate) bool {
	if error.Id() == definition.Id() && error.PrefixId() == definition.PrefixId() {
		return true
	}

	return false
}

func Is(error *VMError, errorComp *VMError) bool {
	if error.Id() == errorComp.Id() &&
		error.PrefixId() == errorComp.PrefixId() {
		return true
	}

	return false
}

func Panic(definition VMErrorTemplate, params ...interface{}) {
	panic(definition.Create(params...))
}
