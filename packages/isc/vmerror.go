package isc

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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

func VMErrorCodeFromBytes(b []byte) (ret VMErrorCode, err error) {
	_, err = rwutil.ReadFromBytes(b, &ret)
	return ret, err
}

func (c VMErrorCode) Bytes() []byte {
	return rwutil.WriteToBytes(&c)
}

func (c VMErrorCode) String() string {
	if c.ContractID == VMCoreErrorContractID {
		return fmt.Sprintf("%04x", c.ID)
	}
	return fmt.Sprintf("%s:%04x", c.ContractID, c.ID)
}

func (c *VMErrorCode) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&c.ContractID)
	c.ID = rr.ReadUint16()
	return rr.Err
}

func (c *VMErrorCode) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&c.ContractID)
	ww.WriteUint16(c.ID)
	return ww.Err
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

func VMErrorTemplateFromBytes(data []byte) (*VMErrorTemplate, error) {
	return rwutil.ReadFromBytes(data, new(VMErrorTemplate))
}

func (e *VMErrorTemplate) Bytes() []byte {
	return rwutil.WriteToBytes(e)
}

func (e *VMErrorTemplate) Code() VMErrorCode {
	return e.code
}

func (e *VMErrorTemplate) Create(params ...any) *VMError {
	validateParams(params)
	return &VMError{
		template: e,
		params:   params,
	}
}

// VMErrorTemplate implements error just in case someone panics with
// VMErrorTemplate by mistake, so that we don't crash the VM because of that.
func (e *VMErrorTemplate) Error() string {
	// calling Sprintf so that it marks missing parameters as errors
	return fmt.Sprintf(e.messageFormat)
}

func (e *VMErrorTemplate) MessageFormat() string {
	return e.messageFormat
}

func (e *VMErrorTemplate) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&e.code)
	e.messageFormat = rr.ReadString()
	return rr.Err
}

func (e *VMErrorTemplate) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&e.code)
	ww.WriteString(e.messageFormat)
	return ww.Err
}

type UnresolvedVMError struct {
	ErrorCode VMErrorCode `json:"code"`
	Params    []any       `json:"params"`
}

var _ VMErrorBase = &UnresolvedVMError{}

type UnresolvedVMErrorJSON struct {
	ErrorCode string   `json:"code"`
	Params    []string `json:"params"`
}

func (e *UnresolvedVMError) AsGoError() error {
	// this is necessary because *UnresolvedVMError(nil) != error(nil)
	if e == nil {
		return nil
	}
	return e
}

func (e *UnresolvedVMError) Bytes() []byte {
	return rwutil.WriteToBytes(e)
}

func (e *UnresolvedVMError) Code() VMErrorCode {
	return e.ErrorCode
}

func (e *UnresolvedVMError) Error() string {
	return fmt.Sprintf("UnresolvedVMError(code: %s)", e.ErrorCode)
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

func (e *UnresolvedVMError) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&e.ErrorCode)
	e.Params = readParams(rr)
	return rr.Err
}

func (e *UnresolvedVMError) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&e.ErrorCode)
	writeParams(ww, e.Params)
	return ww.Err
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
	return rwutil.WriteToBytes(e)
}

func (e *VMError) Code() VMErrorCode {
	return e.template.code
}

func (e *VMError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf(e.MessageFormat(), e.params...)
}

func (e *VMError) MessageFormat() string {
	return e.template.messageFormat
}

func (e *VMError) Params() []any {
	return e.params
}

func (e *VMError) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&e.template.code)
	e.params = readParams(rr)
	return rr.Err
}

func (e *VMError) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&e.template.code)
	writeParams(ww, e.params)
	return ww.Err
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

//nolint:gocyclo
func readParams(rr *rwutil.Reader) []any {
	size := rr.ReadSize16()
	if rr.Err != nil || size == 0 {
		return nil
	}
	ret := make([]any, size)
	for i := 0; i < size; i++ {
		kind := rr.ReadUint8()
		switch reflect.Kind(kind) {
		case reflect.String:
			ret[i] = rr.ReadString()

		case reflect.Uint:
			ret[i] = uint(rr.ReadUint32())
		case reflect.Uint8:
			ret[i] = rr.ReadUint8()
		case reflect.Uint16:
			ret[i] = rr.ReadUint16()
		case reflect.Uint32:
			ret[i] = rr.ReadUint32()
		case reflect.Uint64:
			ret[i] = rr.ReadUint64()

		case reflect.Int:
			ret[i] = int(rr.ReadInt32())
		case reflect.Int8:
			ret[i] = rr.ReadInt8()
		case reflect.Int16:
			ret[i] = rr.ReadInt16()
		case reflect.Int32:
			ret[i] = rr.ReadInt32()
		case reflect.Int64:
			ret[i] = rr.ReadInt64()

		default:
			panic(fmt.Sprintf("no handler for param of kind %s", reflect.Kind(kind)))
		}
	}
	if rr.Err != nil || size == 0 {
		return nil
	}
	return ret
}

func writeParams(ww *rwutil.Writer, params []any) {
	ww.WriteSize16(len(params))
	for _, param := range params {
		t := reflect.TypeOf(param)
		v := reflect.ValueOf(param)
		ww.WriteUint8(uint8(t.Kind()))
		switch t.Kind() {
		case reflect.String:
			ww.WriteString(v.String())

		case reflect.Uint:
			ww.WriteUint32(uint32(v.Uint()))
		case reflect.Uint8:
			ww.WriteUint8(uint8(v.Uint()))
		case reflect.Uint16:
			ww.WriteUint16(uint16(v.Uint()))
		case reflect.Uint32:
			ww.WriteUint32(uint32(v.Uint()))
		case reflect.Uint64:
			ww.WriteUint64(v.Uint())

		case reflect.Int:
			ww.WriteInt32(int32(v.Int()))
		case reflect.Int8:
			ww.WriteInt8(int8(v.Int()))
		case reflect.Int16:
			ww.WriteInt16(int16(v.Int()))
		case reflect.Int32:
			ww.WriteInt32(int32(v.Int()))
		case reflect.Int64:
			ww.WriteInt64(v.Int())

		default:
			panic(fmt.Sprintf("no handler for param of type %s", t.Name()))
		}
	}
}
