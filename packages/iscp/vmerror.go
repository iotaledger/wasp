package iscp

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const VMErrorMessageLimit = math.MaxUint16

// VMCoreErrorContractID defines that all errors with a MaxUint32 contract id will be considered as core errors.
const VMCoreErrorContractID = math.MaxUint32

type VMErrorCode struct {
	ContractID Hname
	ID         uint16
}

func NewVMErrorCode(contractID Hname, id uint16) VMErrorCode {
	return VMErrorCode{ContractID: contractID, ID: id}
}

func NewCoreVMErrorCode(id uint16) VMErrorCode {
	return VMErrorCode{ContractID: VMCoreErrorContractID, ID: id}
}

func (c VMErrorCode) String() string {
	if c.ContractID == VMCoreErrorContractID {
		return fmt.Sprintf("%04x", c.ID)
	}
	return fmt.Sprintf("%s:%04x", c.ContractID, c.ID)
}

func (c VMErrorCode) Bytes() []byte {
	mu := marshalutil.New()
	c.Serialize(mu)
	return mu.Bytes()
}

func (c VMErrorCode) Serialize(mu *marshalutil.MarshalUtil) {
	mu.Write(c.ContractID).
		WriteUint16(c.ID)
}

func VMErrorCodeFromBytes(b []byte) (code VMErrorCode, err error) {
	return VMErrorCodeFromMarshalUtil(marshalutil.New(b))
}

func VMErrorCodeFromMarshalUtil(mu *marshalutil.MarshalUtil) (code VMErrorCode, err error) {
	if code.ContractID, err = HnameFromMarshalUtil(mu); err != nil {
		return
	}
	if code.ID, err = mu.ReadUint16(); err != nil {
		return
	}
	return
}

// vmErrorCode is the common interface of UnresolvedVMError and VMError
type vmErrorCode interface {
	error
	Code() VMErrorCode
}

type VMErrorTemplate struct {
	code          VMErrorCode
	messageFormat string
}

var _ vmErrorCode = &VMErrorTemplate{}

func NewVMErrorTemplate(code VMErrorCode, messageFormat string) *VMErrorTemplate {
	return &VMErrorTemplate{code: code, messageFormat: messageFormat}
}

// VMErrorTemplate implements error just in case someone panics with
// VMErrorTemplate by mistake, so that we don't crash the VM because of that.
func (e *VMErrorTemplate) Error() string {
	// calling Sprintf so that it marks missing parameters as errors
	return fmt.Sprintf(e.messageFormat)
}

func (e *VMErrorTemplate) Code() VMErrorCode {
	return e.code
}

func (e *VMErrorTemplate) MessageFormat() string {
	return e.messageFormat
}

func (e *VMErrorTemplate) Create(params ...interface{}) *VMError {
	return &VMError{
		template: e,
		params:   params,
	}
}

func (e *VMErrorTemplate) Serialize(mu *marshalutil.MarshalUtil) {
	e.code.Serialize(mu)

	messageFormatBytes := []byte(e.MessageFormat())
	mu.WriteUint16(uint16(len(messageFormatBytes))).
		WriteBytes(messageFormatBytes)
}

func (e *VMErrorTemplate) Bytes() []byte {
	mu := marshalutil.New()
	e.Serialize(mu)
	return mu.Bytes()
}

func VMErrorTemplateFromMarshalUtil(mu *marshalutil.MarshalUtil) (*VMErrorTemplate, error) {
	var err error
	e := &VMErrorTemplate{}

	if e.code, err = VMErrorCodeFromMarshalUtil(mu); err != nil {
		return nil, err
	}

	var messageLength uint16
	if messageLength, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if messageInBytes, err := mu.ReadBytes(int(messageLength)); err != nil {
		return nil, err
	} else {
		e.messageFormat = string(messageInBytes)
	}

	return e, nil
}

type UnresolvedVMError struct {
	code   VMErrorCode
	params []interface{}
	hash   uint32
}

var _ vmErrorCode = &UnresolvedVMError{}

func (e *UnresolvedVMError) Error() string {
	return fmt.Sprintf("UnresolvedVMError(code: %s, hash: %x)", e.code, e.hash)
}

func (e *UnresolvedVMError) Hash() uint32 {
	return e.hash
}

func (e *UnresolvedVMError) Code() VMErrorCode {
	return e.code
}

func (e *UnresolvedVMError) Params() []interface{} {
	return e.params
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
	e.code.Serialize(mu)
	mu.WriteUint32(e.hash)
	e.serializeParams(mu)
	return mu.Bytes()
}

func (e *UnresolvedVMError) AsGoError() error {
	// this is necessary because *UnresolvedVMError(nil) != error(nil)
	if e == nil {
		return nil
	}
	return e
}

type VMError struct {
	template *VMErrorTemplate
	params   []interface{}
}

var _ vmErrorCode = &VMError{}

func (e *VMError) Code() VMErrorCode {
	return e.template.code
}

func (e *VMError) MessageFormat() string {
	return e.template.messageFormat
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
	e.template.code.Serialize(mu)
	mu.WriteUint32(e.Hash())
	e.serializeParams(mu)
	return mu.Bytes()
}

func (e *VMError) AsGoError() error {
	// this is necessary because *UnresolvedVMError(nil) != error(nil)
	if e == nil {
		return nil
	}
	return e
}

func (e *VMError) AsUnresolvedError() *UnresolvedVMError {
	return &UnresolvedVMError{
		code:   e.template.code,
		params: e.Params(),
		hash:   e.Hash(),
	}
}

func (e *VMError) AsTemplate() *VMErrorTemplate {
	return e.template
}

func UnresolvedVMErrorFromMarshalUtil(mu *marshalutil.MarshalUtil) (*UnresolvedVMError, error) {
	var err error
	unresolvedError := &UnresolvedVMError{}
	if unresolvedError.code, err = VMErrorCodeFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if unresolvedError.hash, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	if err = unresolvedError.deserializeParams(mu); err != nil {
		return nil, err
	}
	return unresolvedError, nil
}

func GetErrorIDFromMessageFormat(messageFormat string) uint16 {
	messageFormatHash := hashing.HashStrings(messageFormat).Bytes()
	mu := marshalutil.New(messageFormatHash)

	errorId, err := mu.ReadUint16()

	if err != nil {
		panic(err)
	}

	return errorId
}

// VMErrorIs returns true if both objects are vmErrorCode and their VMErrorCode match
func VMErrorIs(error1 interface{}, error2 interface{}) bool {
	if error1, ok := error1.(vmErrorCode); ok {
		if error2, ok := error2.(vmErrorCode); ok {
			return error1.Code() == error2.Code()
		}
	}
	return false
}

// VMErrorMustBe tests VMError, VMErrorTemplate and UnresolvedVMError types against each other by their unique ids and panics if it fails
func VMErrorMustBe(error1 interface{}, error2 interface{}) {
	if !VMErrorIs(error1, error2) {
		panic(fmt.Sprintf("%v does not match %v", error1, error2))
	}
}
