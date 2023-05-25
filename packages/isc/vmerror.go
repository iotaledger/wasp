package isc

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
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

// VMErrorBase is the common interface of UnresolvedVMError and VMError
type VMErrorBase interface {
	error
	Code() VMErrorCode
}

type VMErrorTemplate struct {
	code          VMErrorCode
	messageFormat string
}

var _ VMErrorBase = &VMErrorTemplate{}

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

func (e *VMErrorTemplate) Create(params ...any) *VMError {
	validateParams(params)
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
	messageInBytes, err := mu.ReadBytes(int(messageLength))
	if err != nil {
		return nil, err
	}
	e.messageFormat = string(messageInBytes)
	return e, nil
}

type UnresolvedVMError struct {
	ErrorCode VMErrorCode `json:"code"`
	Params    []any       `json:"params"`
}

var _ VMErrorBase = &UnresolvedVMError{}

func (e *UnresolvedVMError) Error() string {
	return fmt.Sprintf("UnresolvedVMError(code: %s)", e.ErrorCode)
}

func (e *UnresolvedVMError) Code() VMErrorCode {
	return e.ErrorCode
}

func (e *UnresolvedVMError) Bytes() []byte {
	mu := marshalutil.New()
	e.ErrorCode.Serialize(mu)
	serializeParams(mu, e.Params)
	return mu.Bytes()
}

func (e *UnresolvedVMError) AsGoError() error {
	// this is necessary because *UnresolvedVMError(nil) != error(nil)
	if e == nil {
		return nil
	}
	return e
}

type UnresolvedVMErrorJSON struct {
	Params    []string `json:"params"`
	ErrorCode string   `json:"code"`
}

// produce the params as humanly readably json, and the uints as strings
func (e *UnresolvedVMError) ToJSONStruct() *UnresolvedVMErrorJSON {
	if e == nil {
		return &UnresolvedVMErrorJSON{
			Params:    []string{},
			ErrorCode: "",
		}
	}
	return &UnresolvedVMErrorJSON{
		Params:    humanlyReadableParams(e.Params),
		ErrorCode: e.ErrorCode.String(),
	}
}

func humanlyReadableParams(params []any) []string {
	res := make([]string, len(params))
	for i, param := range params {
		res[i] = fmt.Sprintf("%v:%s", param, reflect.TypeOf(param).String())
	}
	return res
}

type VMError struct {
	template *VMErrorTemplate
	params   []any
}

var _ VMErrorBase = &VMError{}

func (e *VMError) Code() VMErrorCode {
	return e.template.code
}

func (e *VMError) MessageFormat() string {
	return e.template.messageFormat
}

func (e *VMError) Params() []any {
	return e.params
}

func (e *VMError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf(e.MessageFormat(), e.params...)
}

func (e *VMError) Bytes() []byte {
	mu := marshalutil.New()
	e.template.code.Serialize(mu)
	serializeParams(mu, e.params)
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
		ErrorCode: e.template.code,
		Params:    e.params,
	}
}

func (e *VMError) AsTemplate() *VMErrorTemplate {
	return e.template
}

func UnresolvedVMErrorFromMarshalUtil(mu *marshalutil.MarshalUtil) (*UnresolvedVMError, error) {
	var err error
	unresolvedError := &UnresolvedVMError{}
	if unresolvedError.ErrorCode, err = VMErrorCodeFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if unresolvedError.Params, err = deserializeParams(mu); err != nil {
		return nil, err
	}
	return unresolvedError, nil
}

func GetErrorIDFromMessageFormat(messageFormat string) uint16 {
	messageFormatHash := hashing.HashStrings(messageFormat).Bytes()
	mu := marshalutil.New(messageFormatHash)

	errorID, err := mu.ReadUint16()
	if err != nil {
		panic(err)
	}

	return errorID
}

// VMErrorIs returns true if the error includes a VMErrorCode in its chain that matches the given code
func VMErrorIs(err error, expected VMErrorBase) bool {
	var vmError VMErrorBase
	if errors.As(err, &vmError) {
		return vmError.Code() == expected.Code()
	}
	return false
}

// VMErrorMustBe panics unless the error includes a VMErrorCode in its chain that matches the given code
func VMErrorMustBe(err error, expected VMErrorBase) {
	if !VMErrorIs(err, expected) {
		panic(fmt.Sprintf("%v does not match %v", err, expected))
	}
}

func validateParams(params []any) {
	if len(params) > 255 {
		panic("params too long")
	}
	for _, param := range params {
		t := reflect.TypeOf(param)
		v := reflect.ValueOf(param)
		switch t.Kind() {
		case reflect.String:
			s := v.String()
			if len(s) > 255 {
				panic("string param too long")
			}
		case reflect.Uint8,
			reflect.Uint16,
			reflect.Uint, reflect.Uint32,
			reflect.Uint64,
			reflect.Int8,
			reflect.Int16,
			reflect.Int, reflect.Int32,
			reflect.Int64:
		default:
			panic(fmt.Sprintf("no handler for param of type %s", t.Name()))
		}
	}
}

func serializeParams(mu *marshalutil.MarshalUtil, params []any) {
	mu.WriteUint8(uint8(len(params)))
	for _, param := range params {
		t := reflect.TypeOf(param)
		v := reflect.ValueOf(param)
		mu.WriteUint8(uint8(t.Kind()))
		switch t.Kind() {
		case reflect.String:
			s := v.String()
			mu.WriteUint8(uint8(len(s)))
			mu.WriteBytes([]byte(s))
		case reflect.Uint8:
			mu.WriteUint8(uint8(v.Uint()))
		case reflect.Uint16:
			mu.WriteUint16(uint16(v.Uint()))
		case reflect.Uint, reflect.Uint32:
			mu.WriteUint32(uint32(v.Uint()))
		case reflect.Uint64:
			mu.WriteUint64(v.Uint())
		case reflect.Int8:
			mu.WriteInt8(int8(v.Int()))
		case reflect.Int16:
			mu.WriteInt16(int16(v.Int()))
		case reflect.Int, reflect.Int32:
			mu.WriteInt32(int32(v.Int()))
		case reflect.Int64:
			mu.WriteInt64(v.Int())
		default:
			panic(fmt.Sprintf("no handler for param of type %s", t.Name()))
		}
	}
}

//nolint:gocyclo,funlen
func deserializeParams(mu *marshalutil.MarshalUtil) ([]any, error) {
	amount, err := mu.ReadUint8()
	if err != nil {
		return nil, err
	}
	if amount == 0 {
		return nil, nil
	}
	ret := make([]any, amount)
	for i := 0; i < int(amount); i++ {
		kind, err := mu.ReadUint8()
		if err != nil {
			return nil, err
		}
		switch reflect.Kind(kind) {
		case reflect.String:
			sz, err := mu.ReadUint8()
			if err != nil {
				return nil, err
			}
			b, err := mu.ReadBytes(int(sz))
			if err != nil {
				return nil, err
			}
			ret[i] = string(b)
		case reflect.Uint8:
			n, err := mu.ReadUint8()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Uint16:
			n, err := mu.ReadUint16()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Uint:
			n, err := mu.ReadUint32()
			if err != nil {
				return nil, err
			}
			ret[i] = uint(n)
		case reflect.Uint32:
			n, err := mu.ReadUint32()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Uint64:
			n, err := mu.ReadUint64()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Int8:
			n, err := mu.ReadInt8()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Int16:
			n, err := mu.ReadInt16()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Int:
			n, err := mu.ReadInt32()
			if err != nil {
				return nil, err
			}
			ret[i] = int(n)
		case reflect.Int32:
			n, err := mu.ReadInt32()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		case reflect.Int64:
			n, err := mu.ReadInt64()
			if err != nil {
				return nil, err
			}
			ret[i] = n
		default:
			panic(fmt.Sprintf("no handler for param of kind %s", reflect.Kind(kind)))
		}
	}
	return ret, nil
}
