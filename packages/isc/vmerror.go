package isc

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
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

func VMErrorCodeFromBytes(b []byte) (VMErrorCode, error) {
	return bcs.Unmarshal[VMErrorCode](b)
}

func (c VMErrorCode) Bytes() []byte {
	return bcs.MustMarshal(&c)
}

func (c VMErrorCode) String() string {
	if c.ContractID == VMCoreErrorContractID {
		return fmt.Sprintf("%04x", c.ID)
	}
	return fmt.Sprintf("%s:%04x", c.ContractID, c.ID)
}

// VMErrorBase is the common interface of UnresolvedVMError and VMError
type VMErrorBase interface {
	error
	Code() VMErrorCode
}

type VMErrorTemplate struct {
	code          VMErrorCode `bcs:"export"`
	messageFormat string      `bcs:"export"`
}

var _ VMErrorBase = &VMErrorTemplate{}

func NewVMErrorTemplate(code VMErrorCode, messageFormat string) *VMErrorTemplate {
	return &VMErrorTemplate{code: code, messageFormat: messageFormat}
}

func VMErrorTemplateFromBytes(data []byte) (*VMErrorTemplate, error) {
	return bcs.Unmarshal[*VMErrorTemplate](data)
}

func (e *VMErrorTemplate) Bytes() []byte {
	return bcs.MustMarshal(e)
}

func (e *VMErrorTemplate) Code() VMErrorCode {
	return e.code
}

func (e *VMErrorTemplate) Create(params ...VMErrorParam) *VMError {
	if params == nil {
		params = []VMErrorParam{}
	}
	validateParams(params)
	return &VMError{
		template: e,
		params:   params,
	}
}

// Error implements the error interface. VMErrorTemplate implements error just in case someone panics with it.
// VMErrorTemplate by mistake, so that we don't crash the VM because of that.
func (e *VMErrorTemplate) Error() string {
	// calling Sprintf so that it marks missing parameters as errors
	return fmt.Sprintf(e.messageFormat)
}

func (e *VMErrorTemplate) MessageFormat() string {
	return e.messageFormat
}

type UnresolvedVMError struct {
	ErrorCode VMErrorCode    `json:"code"`
	Params    []VMErrorParam `json:"params"`
}

var _ VMErrorBase = &UnresolvedVMError{}

func (e *UnresolvedVMError) AsGoError() error {
	// this is necessary because *UnresolvedVMError(nil) != error(nil)
	if e == nil {
		return nil
	}
	return e
}

func (e *UnresolvedVMError) Bytes() []byte {
	return bcs.MustMarshal(e)
}

func (e *UnresolvedVMError) Code() VMErrorCode {
	return e.ErrorCode
}

func (e *UnresolvedVMError) Error() string {
	return fmt.Sprintf("UnresolvedVMError(code: %s)", e.ErrorCode)
}

type VMError struct {
	template *VMErrorTemplate `bcs:"export"`
	params   []VMErrorParam   `bcs:"export"`
}

var _ VMErrorBase = &VMError{}

func (e *VMError) AsGoError() error {
	// this is necessary because *UnresolvedVMError(nil) != error(nil)
	if e == nil {
		return nil
	}
	return e
}

func (e *VMError) AsTemplate() *VMErrorTemplate {
	return e.template
}

func (e *VMError) AsUnresolvedError() *UnresolvedVMError {
	return &UnresolvedVMError{
		ErrorCode: e.template.code,
		Params:    e.params,
	}
}

func (e *VMError) Bytes() []byte {
	return bcs.MustMarshal(e)
}

func (e *VMError) Code() VMErrorCode {
	return e.template.code
}

func (e *VMError) Error() string {
	if e == nil {
		return ""
	}

	anyArr := lo.Map(e.params, func(p VMErrorParam, _ int) any { return p })

	return fmt.Sprintf(e.MessageFormat(), anyArr...)
}

func (e *VMError) MessageFormat() string {
	return e.template.messageFormat
}

func (e *VMError) Params() []VMErrorParam {
	return e.params
}

func GetErrorIDFromMessageFormat(messageFormat string) uint16 {
	messageFormatHash := hashing.HashStrings(messageFormat).Bytes()
	rr := rwutil.NewBytesReader(messageFormatHash)
	errorID := rr.ReadUint16()
	if rr.Err != nil {
		panic(rr.Err)
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

func validateParams(params []VMErrorParam) {
	if len(params) > 255 {
		panic("params too long")
	}
	for _, param := range params {
		t := reflect.TypeOf(param)
		v := reflect.ValueOf(param)
		switch t.Kind() {
		case reflect.String:
			s := v.String()
			if len(s) > 1024 {
				panic("string param too long")
			}

		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:

		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:

		default:
			panic(fmt.Sprintf("no handler for param of type %s", t.Name()))
		}
	}
}

type VMErrorParam any

func init() {
	bcs.RegisterEnumType9[VMErrorParam, int8, uint8, int16, uint16, int32, uint32, int64, uint64, string]()
}
