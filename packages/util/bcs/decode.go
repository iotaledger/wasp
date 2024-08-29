package bcs

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/samber/lo"
)

type Decodable interface {
	UnmarshalBCS(e *Decoder) error
}

type DecoderConfig struct {
	TagName            string
	DefaultTypeOptions *TypeOptions
	CustomDecoders     map[reflect.Type]CustomDecoder
}

func (c *DecoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "bcs"
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

func NewDecoder(src io.Reader, cfg DecoderConfig) *Decoder {
	cfg.InitializeDefaults()

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return &Decoder{
		cfg: cfg,
		r:   *rwutil.NewReader(src),
	}
}

type Decoder struct {
	cfg DecoderConfig
	r   rwutil.Reader
}

func (d *Decoder) Decode(v any) error {
	if v == nil {
		return fmt.Errorf("cannot Decode a nil value")
	}

	vR := reflect.ValueOf(v)

	if vR.Kind() != reflect.Ptr {
		return fmt.Errorf("Decode destination must be a pointer")
	}
	if vR.IsNil() {
		return fmt.Errorf("Decode destination cannot be nil")
	}

	return d.decodeValue(vR.Elem(), nil)
}

func (d *Decoder) decodeValue(v reflect.Value, customTypeOptions *TypeOptions) error {
	v, customDecoder, typeOptions := d.dereferenceValue(v)

	if customDecoder != nil {
		dec, err := customDecoder(d)
		if err != nil {
			return fmt.Errorf("%v: custom Decoder: %w", v.Type(), err)
		}

		v.Set(dec)

		return nil
	}

	if typeOptions == nil {
		o := *d.cfg.DefaultTypeOptions
		typeOptions = &o
	}

	if customTypeOptions != nil {
		typeOptions.Update(*customTypeOptions)
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(d.r.ReadBool())
	case reflect.Int8:
		d.decodeInt(v, Value1Byte, typeOptions.Bytes)
	case reflect.Uint8:
		d.decodeUint(v, Value1Byte, typeOptions.Bytes)
	case reflect.Int16:
		d.decodeInt(v, Value2Bytes, typeOptions.Bytes)
	case reflect.Uint16:
		d.decodeUint(v, Value2Bytes, typeOptions.Bytes)
	case reflect.Int32:
		d.decodeInt(v, Value4Bytes, typeOptions.Bytes)
	case reflect.Uint32:
		d.decodeUint(v, Value4Bytes, typeOptions.Bytes)
	case reflect.Int64:
		d.decodeInt(v, Value8Bytes, typeOptions.Bytes)
	case reflect.Uint64:
		d.decodeUint(v, Value8Bytes, typeOptions.Bytes)
	case reflect.Int:
		d.decodeInt(v, Value8Bytes, typeOptions.Bytes)
	case reflect.String:
		v.SetString(d.r.ReadString())
	case reflect.Slice:
		if err := d.decodeSlice(v, typeOptions); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Array:
		if err := d.decodeArray(v); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Map:
		if err := d.decodeMap(v, typeOptions); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Struct:
		if err := d.decodeStruct(v); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	default:
		return fmt.Errorf("%v: cannot Decode unknown type type", v.Type())
	}

	if d.r.Err != nil {
		return fmt.Errorf("%v: %w", v.Type(), d.r.Err)
	}

	return nil
}

func (d *Decoder) dereferenceValue(v reflect.Value) (dereferenced reflect.Value, customDecoder CustomDecoder, typeOptions *TypeOptions) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Map {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		} else if v.Kind() == reflect.Map && v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}

		customDecoder, typeOptions = d.retrieveTypeInfo(v)
		if customDecoder != nil || typeOptions != nil {
			return v, customDecoder, typeOptions
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		} else {
			return v, nil, nil
		}
	}

	customDecoder, typeOptions = d.retrieveTypeInfo(v)
	if customDecoder != nil || typeOptions != nil {
		return v, customDecoder, typeOptions
	}

	return v, nil, nil
}

func (d *Decoder) retrieveTypeInfo(v reflect.Value) (customDecoder CustomDecoder, _ *TypeOptions) {
	vI := v.Interface()

	if customDecoder, ok := d.cfg.CustomDecoders[v.Type()]; ok {
		return customDecoder, nil
	}

	if decodable, ok := vI.(Decodable); ok {
		return func(e *Decoder) (reflect.Value, error) {
			if err := decodable.UnmarshalBCS(e); err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(decodable), nil
		}, nil
	}

	if bcsType, ok := vI.(BCSType); ok {
		o := bcsType.BCSOptions()
		return nil, &o
	}

	return nil, nil
}

func (d *Decoder) decodeInt(v reflect.Value, origSize, customSize ValueBytesCount) {
	size := lo.Ternary(customSize != 0, customSize, origSize)

	switch size {
	case Value1Byte:
		v.SetInt(int64(d.r.ReadInt8()))
	case Value2Bytes:
		v.SetInt(int64(d.r.ReadInt16()))
	case Value4Bytes:
		v.SetInt(int64(d.r.ReadInt32()))
	case Value8Bytes:
		v.SetInt(d.r.ReadInt64())
	default:
		panic(fmt.Errorf("invalid value size: %v", size))
	}
}

func (d *Decoder) decodeUint(v reflect.Value, origSize, customSize ValueBytesCount) {
	size := lo.Ternary(customSize != 0, customSize, origSize)

	switch size {
	case Value1Byte:
		v.SetUint(uint64(d.r.ReadUint8()))
	case Value2Bytes:
		v.SetUint(uint64(d.r.ReadUint16()))
	case Value4Bytes:
		v.SetUint(uint64(d.r.ReadUint32()))
	case Value8Bytes:
		v.SetUint(d.r.ReadUint64())
	default:
		panic(fmt.Errorf("invalid value size: %v", size))
	}
}

func (d *Decoder) decodeSlice(v reflect.Value, typOpts *TypeOptions) error {
	var length int

	switch typOpts.LenBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes:
		length = int(d.r.ReadSize32())
	default:
		return fmt.Errorf("invalid array size type: %v", typOpts.LenBytes)
	}

	if length == 0 {
		return nil
	}

	elems := reflect.MakeSlice(v.Type(), length, length)

	for i := 0; i < length; i++ {
		elem := reflect.New(v.Type().Elem()).Elem()

		if err := d.decodeValue(elem, nil); err != nil {
			return fmt.Errorf("[%v]: %w", i, err)
		}

		elems.Index(i).Set(elem)
	}

	v.Set(elems)

	return nil
}

func (d *Decoder) decodeArray(v reflect.Value) error {
	elemType := v.Type().Elem()

	for i := 0; i < v.Type().Len(); i++ {
		elem := reflect.New(elemType).Elem()

		if err := d.decodeValue(elem, nil); err != nil {
			return fmt.Errorf("[%v]: %w", i, err)
		}

		v.Index(i).Set(elem)
	}

	return nil
}

func (d *Decoder) decodeMap(v reflect.Value, typOpts *TypeOptions) error {
	var length int

	switch typOpts.LenBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes:
		length = int(d.r.ReadSize32())
	default:
		return fmt.Errorf("invalid map size type: %v", typOpts.LenBytes)
	}

	v.Set(reflect.MakeMap(v.Type()))

	keyType := v.Type().Key()
	valueType := v.Type().Elem()

	for i := 0; i < length; i++ {
		key := reflect.New(keyType).Elem()
		value := reflect.New(valueType).Elem()

		if err := d.decodeValue(key, nil); err != nil {
			return fmt.Errorf("key: %w", err)
		}

		if err := d.decodeValue(value, nil); err != nil {
			return fmt.Errorf("value: %w", err)
		}

		v.SetMapIndex(key, value)
	}

	return nil
}

func (d *Decoder) decodeStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)

		fieldOpts, hasTag, err := d.fieldOptsFromTag(fieldType)
		if err != nil {
			return fmt.Errorf("%v: parsing annotation: %w", fieldType.Name, err)
		}

		if fieldOpts.Skip {
			continue
		}

		fieldVal := v.Field(i)

		if !fieldType.IsExported() {
			if !hasTag {
				continue
			}

			// The field is unexported, but it has a tag, so we need to serialize it.
			// Trick to access unexported fields: https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields/43918797#43918797
			fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		}

		fieldKind := fieldVal.Kind()

		if fieldKind == reflect.Ptr || fieldKind == reflect.Interface || fieldKind == reflect.Map {
			if fieldOpts.Optional {
				present := d.r.ReadByte()

				if present == 0 {
					continue
				}
			}
		}

		if err := d.decodeValue(fieldVal, &fieldOpts.TypeOptions); err != nil {
			return fmt.Errorf("%v: %w", fieldType.Name, err)
		}
	}

	return nil
}

func (d *Decoder) fieldOptsFromTag(fieldType reflect.StructField) (FieldOptions, bool, error) {
	a, hasTag := fieldType.Tag.Lookup(d.cfg.TagName)

	fieldOpts, err := FieldOptionsFromTag(a, *d.cfg.DefaultTypeOptions)
	if err != nil {
		return FieldOptions{}, false, fmt.Errorf("%v: parsing annotation: %w", fieldType.Name, err)
	}

	return fieldOpts, hasTag, nil
}

// func (d *Decoder) Writer() *rwutil.Writer {
// 	return &d.w
// }

func (d *Decoder) Read(n int) ([]byte, error) {
	b := make([]byte, n)
	d.r.ReadN(b)

	return b, d.r.Err
}

func Unmarshal[T any](b []byte) (T, error) {
	var t T
	if err := NewDecoder(bytes.NewReader(b), DecoderConfig{}).Decode(&t); err != nil {
		return t, err
	}

	return t, nil
}

type CustomDecoder func(e *Decoder) (reflect.Value, error)

var CustomDecoders = make(map[reflect.Type]CustomDecoder)

func MakeCustomDecoder[V any](f func(e *Decoder) (V, error)) func(e *Decoder) (reflect.Value, error) {
	return func(e *Decoder) (reflect.Value, error) {
		r, err := f(e)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(r), nil
	}
}

func AddCustomDecoder[V any](f func(e *Decoder) (V, error)) {
	CustomDecoders[reflect.TypeOf(lo.Empty[V]())] = MakeCustomDecoder(f)
}
