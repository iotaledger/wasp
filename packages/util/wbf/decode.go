package wbf

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/samber/lo"
)

type Decodable interface {
	WBFDecode(e *Decoder) error
}

type DecoderConfig struct {
	TagName            string
	DefaultTypeOptions *TypeOptions
	CustomDecoders     map[reflect.Type]CustomDecoder
}

func (c *DecoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "wbf"
	}
	if c.DefaultTypeOptions == nil {
		c.DefaultTypeOptions = &DefaultTypeOptions
	}
	if c.CustomDecoders == nil {
		c.CustomDecoders = CustomDecoders
	}
}

func (c *DecoderConfig) Validate() error {
	if err := c.DefaultTypeOptions.Validate(); err != nil {
		return fmt.Errorf("default array len size: %w", err)
	}

	return nil
}

func NewDecoder(dest io.Writer, cfg DecoderConfig) *Decoder {
	cfg.InitializeDefaults()

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return &Decoder{
		cfg: cfg,
		w:   *rwutil.NewWriter(dest),
	}
}

type Decoder struct {
	cfg DecoderConfig
	w   rwutil.Writer
}

func (e *Decoder) Decode(v any) error {
	if v == nil {
		return fmt.Errorf("cannot Decode a nil value")
	}

	vR := reflect.ValueOf(v)

	return e.DecodeValue(vR, nil)
}

func (e *Decoder) DecodeValue(v reflect.Value, customTypeOptions *TypeOptions) error {
	v, customDecoder, typeOptions := e.dereferenceValue(v)

	if customDecoder != nil {
		if err := customDecoder(e, v); err != nil {
			return fmt.Errorf("%v: custom Decoder: %w", v.Type(), err)
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
		e.DecodeInt(v, Value1Byte, typeOptions.Bytes)
	case reflect.Uint8:
		e.DecodeUint(v, Value1Byte, typeOptions.Bytes)
	case reflect.Int16:
		e.DecodeInt(v, Value2Bytes, typeOptions.Bytes)
	case reflect.Uint16:
		e.DecodeUint(v, Value2Bytes, typeOptions.Bytes)
	case reflect.Int32:
		e.DecodeInt(v, Value4Bytes, typeOptions.Bytes)
	case reflect.Uint32:
		e.DecodeUint(v, Value4Bytes, typeOptions.Bytes)
	case reflect.Int64:
		e.DecodeInt(v, Value8Bytes, typeOptions.Bytes)
	case reflect.Uint64:
		e.DecodeUint(v, Value8Bytes, typeOptions.Bytes)
	case reflect.Int:
		e.DecodeInt(v, Value8Bytes, typeOptions.Bytes)
	case reflect.String:
		e.w.WriteString(v.String())
	case reflect.Array, reflect.Slice:
		if err := e.DecodeArray(v, typeOptions); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Map:
		if err := e.DecodeMap(v, typeOptions); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Struct:
		if err := e.DecodeStruct(v); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	default:
		return fmt.Errorf("%v: cannot Decode unknown type type", v.Type())
	}

	if e.w.Err != nil {
		return fmt.Errorf("%v: %w", v.Type(), e.w.Err)
	}

	return nil
}

func (e *Decoder) dereferenceValue(v reflect.Value) (dereferenced reflect.Value, customDecoder CustomDecoder, typeOptions *TypeOptions) {
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return v, nil, nil
	}

	customDecoder, typeOptions = e.retrieveTypeInfo(v)
	if customDecoder != nil || typeOptions != nil {
		return v, customDecoder, typeOptions
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v, nil, nil
		}

		v = v.Elem()

		customDecoder, typeOptions = e.retrieveTypeInfo(v)
		if customDecoder != nil || typeOptions != nil {
			return v, customDecoder, typeOptions
		}
	}

	return v, nil, nil
}

func (e *Decoder) retrieveTypeInfo(v reflect.Value) (customDecoder CustomDecoder, _ *TypeOptions) {
	vI := v.Interface()

	if customDecoder, ok := e.cfg.CustomDecoders[v.Type()]; ok {
		return customDecoder, nil
	}

	if Decodable, ok := vI.(Decodable); ok {
		return func(e *Decoder, v reflect.Value) error {
			return Decodable.WBFDecode(e)
		}, nil
	}

	if wbfType, ok := vI.(WBFType); ok {
		o := wbfType.WBFOptions()
		return nil, &o
	}

	return nil, nil
}

func (e *Decoder) DecodeInt(v reflect.Value, origSize, customSize ValueBytesCount) {
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

func (e *Decoder) DecodeUint(v reflect.Value, origSize, customSize ValueBytesCount) {
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

func (e *Decoder) DecodeArray(v reflect.Value, typOpts *TypeOptions) error {
	switch typOpts.LenBytes {
	case Len2Bytes:
		e.w.WriteSize16(v.Len())
	case Len4Bytes:
		e.w.WriteSize32(v.Len())
	default:
		return fmt.Errorf("invalid collection size type: %v", typOpts.LenBytes)
	}

	for i := 0; i < v.Len(); i++ {
		if err := e.DecodeValue(v.Index(i), nil); err != nil {
			return fmt.Errorf("[%v]: %w", i, err)
		}
	}

	return nil
}

func (e *Decoder) DecodeMap(v reflect.Value, typOpts *TypeOptions) error {
	switch typOpts.LenBytes {
	case Len2Bytes:
		e.w.WriteSize16(v.Len())
	case Len4Bytes:
		e.w.WriteSize32(v.Len())
	default:
		return fmt.Errorf("invalid collection size type: %v", typOpts.LenBytes)
	}

	for elem := v.MapRange(); elem.Next(); {
		if err := e.DecodeValue(elem.Key(), nil); err != nil {
			return fmt.Errorf("key: %w", err)
		}

		if err := e.DecodeValue(elem.Value(), nil); err != nil {
			return fmt.Errorf("value: %w", err)
		}
	}

	return nil
}

func (e *Decoder) DecodeStruct(v reflect.Value) error {
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

		if err := e.DecodeValue(fieldVal, &fieldOpts.TypeOptions); err != nil {
			return fmt.Errorf("%v: %w", fieldType.Name, err)
		}
	}

	return nil
}

// func (e *Decoder) Writer() *rwutil.Writer {
// 	return &e.w
// }

func (e *Decoder) Write(b []byte) {
	e.w.WriteN(b)
}

func Decode(v any) ([]byte, error) {
	var buf bytes.Buffer

	if err := NewDecoder(&buf, DecoderConfig{}).Decode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type CustomDecoder func(e *Decoder, v reflect.Value) error

var CustomDecoders = make(map[reflect.Type]CustomDecoder)

func MakeCustomDecoder[V any](f func(e *Decoder, v V) error) func(e *Decoder, v reflect.Value) error {
	return func(e *Decoder, v reflect.Value) error {
		return f(e, v.Interface().(V))
	}
}

func AddCustomDecoder[V any](f func(e *Decoder, v V) error) {
	CustomDecoders[reflect.TypeOf(lo.Empty[V]())] = MakeCustomDecoder(f)
}
