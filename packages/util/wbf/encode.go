package wbf

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/samber/lo"
)

type Encodable interface {
	WBFEncode(e *Encoder) error
}

type EncoderConfig struct {
	TagName            string
	DefaultTypeOptions *TypeOptions
	// IncludeUnexported bool
	// IncludeUntaggedUnexported bool
	// ExcludeUntagged           bool
	CustomEncoders map[reflect.Type]CustomEncoder
}

func (c *EncoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "wbf"
	}
	if c.DefaultTypeOptions == nil {
		c.DefaultTypeOptions = &DefaultTypeOptions
	}
	if c.CustomEncoders == nil {
		c.CustomEncoders = CustomEncoders
	}
}

func (c *EncoderConfig) Validate() error {
	if err := c.DefaultTypeOptions.Validate(); err != nil {
		return fmt.Errorf("default array len size: %w", err)
	}

	return nil
}

func NewEncoder(dest io.Writer, cfg EncoderConfig) *Encoder {
	cfg.InitializeDefaults()

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return &Encoder{
		cfg: cfg,
		w:   *rwutil.NewWriter(dest),
	}
}

type Encoder struct {
	cfg EncoderConfig
	w   rwutil.Writer
}

func (e *Encoder) Encode(v any) error {
	if v == nil {
		return fmt.Errorf("cannot encode a nil value")
	}

	return e.encodeValue(reflect.ValueOf(v), nil)
}

func (e *Encoder) encodeValue(v reflect.Value, customTypeOptions *TypeOptions) error {
	v, customEncoder, typeOptions := e.dereferenceValue(v)

	if customEncoder != nil {
		if err := customEncoder(e, v); err != nil {
			return fmt.Errorf("%v: custom encoder: %w", v.Type(), err)
		}

		return nil
	}

	if typeOptions == nil {
		o := *e.cfg.DefaultTypeOptions
		typeOptions = &o
	}

	if customTypeOptions != nil {
		typeOptions.Update(*customTypeOptions)
	}

	switch v.Kind() {
	case reflect.Bool:
		e.w.WriteBool(v.Bool())
	case reflect.Int8:
		e.encodeInt(v, Value1Byte, typeOptions.Bytes)
	case reflect.Uint8:
		e.encodeUint(v, Value1Byte, typeOptions.Bytes)
	case reflect.Int16:
		e.encodeInt(v, Value2Bytes, typeOptions.Bytes)
	case reflect.Uint16:
		e.encodeUint(v, Value2Bytes, typeOptions.Bytes)
	case reflect.Int32:
		e.encodeInt(v, Value4Bytes, typeOptions.Bytes)
	case reflect.Uint32:
		e.encodeUint(v, Value4Bytes, typeOptions.Bytes)
	case reflect.Int64:
		e.encodeInt(v, Value8Bytes, typeOptions.Bytes)
	case reflect.Uint64:
		e.encodeUint(v, Value8Bytes, typeOptions.Bytes)
	case reflect.Int:
		e.encodeInt(v, Value8Bytes, typeOptions.Bytes)
	case reflect.String:
		e.w.WriteString(v.String())
	case reflect.Array, reflect.Slice:
		if err := e.encodeArray(v, typeOptions); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Map:
		if err := e.encodeMap(v, typeOptions); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Struct:
		if err := e.encodeStruct(v); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	default:
		return fmt.Errorf("%v: cannot encode unknown type type", v.Type())
	}

	if e.w.Err != nil {
		return fmt.Errorf("%v: %w", v.Type(), e.w.Err)
	}

	return nil
}

func (e *Encoder) dereferenceValue(v reflect.Value) (dereferenced reflect.Value, customEncoder CustomEncoder, typeOptions *TypeOptions) {
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return v, nil, nil
	}

	customEncoder, typeOptions = e.retrieveTypeInfo(v)
	if customEncoder != nil || typeOptions != nil {
		return v, customEncoder, typeOptions
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v, nil, nil
		}

		v = v.Elem()

		customEncoder, typeOptions = e.retrieveTypeInfo(v)
		if customEncoder != nil || typeOptions != nil {
			return v, customEncoder, typeOptions
		}
	}

	return v, nil, nil
}

func (e *Encoder) retrieveTypeInfo(v reflect.Value) (customEncoder CustomEncoder, _ *TypeOptions) {
	vI := v.Interface()

	if customEncoder, ok := e.cfg.CustomEncoders[v.Type()]; ok {
		return customEncoder, nil
	}

	if encodable, ok := vI.(Encodable); ok {
		return func(e *Encoder, v reflect.Value) error {
			return encodable.WBFEncode(e)
		}, nil
	}

	if wbfType, ok := vI.(WBFType); ok {
		o := wbfType.WBFOptions()
		return nil, &o
	}

	return nil, nil
}

func (e *Encoder) encodeInt(v reflect.Value, origSize, customSize ValueBytesCount) {
	size := lo.Ternary(customSize != 0, customSize, origSize)

	switch size {
	case Value1Byte:
		e.w.WriteInt8(int8(v.Int()))
	case Value2Bytes:
		e.w.WriteInt16(int16(v.Int()))
	case Value4Bytes:
		e.w.WriteInt32(int32(v.Int()))
	case Value8Bytes:
		e.w.WriteInt64(v.Int())
	default:
		panic(fmt.Errorf("invalid value size: %v", size))
	}
}

func (e *Encoder) encodeUint(v reflect.Value, origSize, customSize ValueBytesCount) {
	size := lo.Ternary(customSize != 0, customSize, origSize)

	switch size {
	case Value1Byte:
		e.w.WriteUint8(uint8(v.Uint()))
	case Value2Bytes:
		e.w.WriteUint16(uint16(v.Uint()))
	case Value4Bytes:
		e.w.WriteUint32(uint32(v.Uint()))
	case Value8Bytes:
		e.w.WriteUint64(v.Uint())
	default:
		panic(fmt.Errorf("invalid value size: %v", size))
	}
}

func (e *Encoder) encodeArray(v reflect.Value, typOpts *TypeOptions) error {
	switch typOpts.LenBytes {
	case Len2Bytes:
		e.w.WriteSize16(v.Len())
	case Len4Bytes:
		e.w.WriteSize32(v.Len())
	default:
		return fmt.Errorf("invalid collection size type: %v", typOpts.LenBytes)
	}

	for i := 0; i < v.Len(); i++ {
		if err := e.encodeValue(v.Index(i), nil); err != nil {
			return fmt.Errorf("[%v]: %w", i, err)
		}
	}

	return nil
}

func (e *Encoder) encodeMap(v reflect.Value, typOpts *TypeOptions) error {
	switch typOpts.LenBytes {
	case Len2Bytes:
		e.w.WriteSize16(v.Len())
	case Len4Bytes:
		e.w.WriteSize32(v.Len())
	default:
		return fmt.Errorf("invalid collection size type: %v", typOpts.LenBytes)
	}

	for elem := v.MapRange(); elem.Next(); {
		if err := e.encodeValue(elem.Key(), nil); err != nil {
			return fmt.Errorf("key: %w", err)
		}

		if err := e.encodeValue(elem.Value(), nil); err != nil {
			return fmt.Errorf("value: %w", err)
		}
	}

	return nil
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)

		fieldOpts, err := FieldOptionsFromAnnotation(fieldType.Tag.Get(e.cfg.TagName), *e.cfg.DefaultTypeOptions)
		if err != nil {
			return fmt.Errorf("%v: parsing annotation: %w", fieldType.Name, err)
		}

		fieldVal := v.Field(i)

		if fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface {
			isNil := fieldVal.IsNil()

			if isNil && !fieldOpts.Optional {
				return fmt.Errorf("%v: non-optional nil value", fieldType.Name)
			}

			if fieldOpts.Optional {
				e.w.WriteByte(lo.Ternary[byte](isNil, 0, 1))
			}

			if isNil {
				continue
			}

			//fieldVal = fieldVal.Elem()
		}

		if err := e.encodeValue(fieldVal, &fieldOpts.TypeOptions); err != nil {
			return fmt.Errorf("%v: %w", fieldType.Name, err)
		}
	}

	return nil
}

// func (e *Encoder) Writer() *rwutil.Writer {
// 	return &e.w
// }

func (e *Encoder) Write(b []byte) {
	e.w.WriteN(b)
}

func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer

	if err := NewEncoder(&buf, EncoderConfig{}).Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type CustomEncoder func(e *Encoder, v reflect.Value) error

var CustomEncoders = make(map[reflect.Type]CustomEncoder)

func MakeCustomEncoder[V any](f func(e *Encoder, v V) error) func(e *Encoder, v reflect.Value) error {
	return func(e *Encoder, v reflect.Value) error {
		return f(e, v.Interface().(V))
	}
}

func AddCustomEncoder[V any](f func(e *Encoder, v V) error) {
	CustomEncoders[reflect.TypeOf(lo.Empty[V]())] = MakeCustomEncoder(f)
}
