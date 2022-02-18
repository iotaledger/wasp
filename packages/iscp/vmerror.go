package iscp

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
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
		messageFormat: e.MessageFormat(),
		prefixId:      e.PrefixId(),
		id:            e.Id(),
		params:        params,
	}
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

type UnresolvedVMError struct {
	prefixId Hname
	id       uint16
	params   []interface{}
	hash     uint32
}

func (e *UnresolvedVMError) Error() string {
	return fmt.Sprintf("UnresolvedError (prefixId: %d, errorId: %d, hash: %x)", e.PrefixId(), e.Id(), e.hash)
}

func (e *UnresolvedVMError) Hash() uint32 {
	return e.hash
}

func (e *UnresolvedVMError) PrefixId() Hname {
	return e.prefixId
}

func (e *UnresolvedVMError) Id() uint16 {
	return e.id
}

func (e *UnresolvedVMError) Params() []interface{} {
	return e.params
}

type VMErrorMessageResolver func(*UnresolvedVMError) (string, error)

func (e *UnresolvedVMError) ResolveToVMError(resolver VMErrorMessageResolver) (*VMError, error) {
	if e == nil {
		return nil, nil
	}

	vmError := VMError{
		prefixId: e.PrefixId(),
		id:       e.Id(),
		params:   e.Params(),
	}

	var err error

	if vmError.messageFormat, err = resolver(e); err != nil {
		return nil, err
	}

	return &vmError, nil
}

func (e *UnresolvedVMError) deserializeParams(mu *marshalutil.MarshalUtil) error {
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

func (e *UnresolvedVMError) serializeParams(mu *marshalutil.MarshalUtil) {
	bytes, err := json.Marshal(e.Params())

	if err != nil {
		panic(err)
	}

	mu.WriteUint16(uint16(len(bytes)))
	mu.WriteBytes(bytes)
}

func (e *UnresolvedVMError) Bytes() []byte {
	mu := marshalutil.New()

	mu.WriteUint32(uint32(e.PrefixId())).
		WriteUint16(e.Id()).
		WriteUint32(e.hash)

	e.serializeParams(mu)

	return mu.Bytes()
}

func (e *UnresolvedVMError) Definition() VMErrorTemplate {
	return VMErrorTemplate{prefixId: e.PrefixId(), id: e.Id()}
}

func (e *UnresolvedVMError) AsGoError() error {
	if e == nil {
		return nil
	}

	return e
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

func (e *VMError) Error() string {
	return fmt.Sprintf(e.MessageFormat(), e.Params()...)
}

func (e *VMError) Hash() uint32 {
	if len(e.MessageFormat()) == 0 {
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

	e.serializeParams(mu)

	return mu.Bytes()
}

func (e *VMError) AsGoError() error {
	if e == nil {
		return nil
	}

	return e
}

func (e *VMError) AsUnresolvedError() *UnresolvedVMError {
	return &UnresolvedVMError{
		hash:     e.Hash(),
		id:       e.Id(),
		params:   e.Params(),
		prefixId: e.PrefixId(),
	}
}

func (e *VMError) AsTemplate() VMErrorTemplate {
	return VMErrorTemplate{prefixId: e.PrefixId(), id: e.Id(), messageFormat: e.MessageFormat()}
}

func UnresolvedVMErrorFromMarshalUtil(mu *marshalutil.MarshalUtil) (*UnresolvedVMError, error) {
	var err error

	unresolvedError := UnresolvedVMError{}

	var prefixId uint32

	if prefixId, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	unresolvedError.prefixId = Hname(prefixId)

	if unresolvedError.id, err = mu.ReadUint16(); err != nil {
		return nil, err
	}

	if unresolvedError.hash, err = mu.ReadUint32(); err != nil {
		return nil, err
	}

	if err = unresolvedError.deserializeParams(mu); err != nil {
		return nil, err
	}

	return &unresolvedError, nil
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

func extractErrorPrimitive(err interface{}) (bool, Hname, uint16) {
	if e, ok := err.(*VMError); ok {
		return true, e.PrefixId(), e.Id()
	}

	if e, ok := err.(*UnresolvedVMError); ok {
		return true, e.PrefixId(), e.Id()
	}

	if e, ok := err.(*VMErrorTemplate); ok {
		return true, e.PrefixId(), e.Id()
	}

	return false, 0, 0
}

// VMErrorIs tests VMError, VMErrorTemplate and UnresolvedVMError types against each other by their unique ids
func VMErrorIs(error1 interface{}, error2 interface{}) bool {
	isError, prefixId1, errorId1 := extractErrorPrimitive(error1)

	if !isError {
		return false
	}

	isError, prefixId2, errorId2 := extractErrorPrimitive(error2)

	if !isError {
		return false
	}

	if prefixId1 == prefixId2 && errorId1 == errorId2 {
		return true
	}

	return false
}

// VMErrorMustBe tests VMError, VMErrorTemplate and UnresolvedVMError types against each other by their unique ids and panics if it fails
func VMErrorMustBe(error1 interface{}, error2 interface{}) {
	if !VMErrorIs(error1, error2) {
		panic(fmt.Sprintf("%v does not match %v", error1, error2))
	}
}
