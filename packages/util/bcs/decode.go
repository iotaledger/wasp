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

var decodableT = reflect.TypeOf((*Decodable)(nil)).Elem()

type DecoderConfig struct {
	TagName        string
	CustomDecoders map[reflect.Type]CustomDecoder
}

func (c *DecoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "bcs"
	}
	if c.CustomDecoders == nil {
		c.CustomDecoders = CustomDecoders
	}
}

func NewDecoder(src io.Reader, cfg DecoderConfig) *Decoder {
	cfg.InitializeDefaults()

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

func (d *Decoder) decodeValue(v reflect.Value, typeOptionsFromTag *TypeOptions) error {
	v, typeOptions, isEnum, customDecoder := d.dereferenceValue(v)

	if customDecoder != nil {
		dec, err := customDecoder(d)
		if err != nil {
			return fmt.Errorf("%v: custom decoder: %w", v.Type(), err)
		}

		v.Set(dec)

		return nil
	}

	if typeOptionsFromTag != nil {
		typeOptions.Update(*typeOptionsFromTag)
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(d.r.ReadBool())
	case reflect.Int8:
		if err := d.decodeInt(v, Value1Byte, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Uint8:
		if err := d.decodeUint(v, Value1Byte, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Int16:
		if err := d.decodeInt(v, Value2Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Uint16:
		if err := d.decodeUint(v, Value2Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Int32:
		if err := d.decodeInt(v, Value4Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Uint32:
		if err := d.decodeUint(v, Value4Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Int64:
		if err := d.decodeInt(v, Value8Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Uint64:
		if err := d.decodeUint(v, Value8Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Int:
		if err := d.decodeInt(v, Value8Bytes, typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
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
		if isEnum {
			if err := d.decodeStructEnum(v); err != nil {
				return fmt.Errorf("%v: %w", v.Type(), err)
			}
		} else {
			if err := d.decodeStruct(v); err != nil {
				return fmt.Errorf("%v: %w", v.Type(), err)
			}
		}
	case reflect.Interface:
		if err := d.decodeInterfaceEnum(v); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	default:
		return fmt.Errorf("%v: cannot decode unknown type", v.Type())
	}

	if d.r.Err != nil {
		return fmt.Errorf("%v: %w", v.Type(), d.r.Err)
	}

	return nil
}

func (d *Decoder) dereferenceValue(v reflect.Value) (dereferenced reflect.Value, _ TypeOptions, isEnum bool, _ CustomDecoder) {
	// Getting rid of redundant pointers AND creating a new value to be able to set it.
loop:
	for {
		switch v.Kind() {
		case reflect.Ptr:
			v.Set(reflect.New(v.Type().Elem()))
		case reflect.Map:
			v.Set(reflect.MakeMap(v.Type()))
			break loop
		default:
			break loop
		}

		typeOptions, isEnum, customDecoder := d.retrieveTypeInfo(v)
		if typeOptions != nil || isEnum || customDecoder != nil {
			if typeOptions == nil {
				typeOptions = &TypeOptions{}
			}

			return v, *typeOptions, isEnum, customDecoder
		}

		v = v.Elem()
	}

	typeOptions, isEnum, customDecoder := d.retrieveTypeInfo(v)
	if typeOptions == nil {
		typeOptions = &TypeOptions{}
	}

	return v, *typeOptions, isEnum, customDecoder
}

func (d *Decoder) retrieveTypeInfo(v reflect.Value) (_ *TypeOptions, isEnum bool, _ CustomDecoder) {
	if customDecoder := d.getCustomDecoder(v); customDecoder != nil {
		return nil, false, customDecoder
	}

	switch v.Kind() {
	case reflect.Interface:
		// This is enum - we can't retrieve any type info, because we don't know enum variant type.
		return nil, true, nil
	case reflect.Struct:
		if v.Type().Implements(enumT) {
			return nil, true, nil
		}
	}

	vI := v.Interface()

	// Check if this type provides decoding options
	if bcsType, ok := vI.(BCSType); ok {
		typeOptions := bcsType.BCSOptions()

		return &typeOptions, false, nil
	}

	return nil, false, nil
}

func (d *Decoder) getCustomDecoder(v reflect.Value) CustomDecoder {
	t := v.Type()

	if customDecoder, ok := d.cfg.CustomDecoders[t]; ok {
		return customDecoder
	}

	// We skip interfaces, because alhgouth they can have custom decoders set for them as global option,
	// they still can't providate them through methods, because their actual type is unknown.
	if t.Kind() != reflect.Interface && t.Implements(decodableT) {
		decodable := v.Interface().(Decodable)

		customDecoder := func(e *Decoder) (reflect.Value, error) {
			err := decodable.UnmarshalBCS(e)
			return reflect.ValueOf(decodable), err
		}

		return customDecoder
	}

	return nil
}

func (d *Decoder) decodeInt(v reflect.Value, origSize, customSize ValueBytesCount) error {
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
		return fmt.Errorf("invalid value size: %v", size)
	}

	return nil
}

func (d *Decoder) decodeUint(v reflect.Value, origSize, customSize ValueBytesCount) error {
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
		return fmt.Errorf("invalid value size: %v", size)
	}

	return nil
}

func (d *Decoder) decodeSlice(v reflect.Value, typOpts TypeOptions) error {
	var length int

	switch typOpts.LenBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes, 0:
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

func (d *Decoder) decodeMap(v reflect.Value, typOpts TypeOptions) error {
	var length int

	switch typOpts.LenBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes, 0:
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

	fieldOpts, err := FieldOptionsFromTag(a)
	if err != nil {
		return FieldOptions{}, false, fmt.Errorf("%v: parsing annotation: %w", fieldType.Name, err)
	}

	return fieldOpts, hasTag, nil
}

func (d *Decoder) decodeInterfaceEnum(v reflect.Value) error {
	variants, registered := EnumTypes[v.Type()]
	if !registered {
		return fmt.Errorf("interface type %v is not registered as enum", v.Type())
	}

	variantIdx := d.r.ReadSize32()
	if d.r.Err != nil {
		return d.r.Err
	}

	if int(variantIdx) >= len(variants) {
		return fmt.Errorf("invalid variant index %v for enum %v", variantIdx, v.Type())
	}

	variant := reflect.New(variants[variantIdx]).Elem()

	if err := d.decodeValue(variant, nil); err != nil {
		return fmt.Errorf("%v: %w", variants[variantIdx], err)
	}

	v.Set(variant)

	return nil
}

func (d *Decoder) decodeStructEnum(v reflect.Value) error {
	variantIdx := d.r.ReadSize32()

	t := v.Type()

	if t.NumField() <= int(variantIdx) {
		return fmt.Errorf("invalid variant index %v for enum %v with %v options", variantIdx, t, t.NumField())
	}

	return d.decodeValue(v.Field(int(variantIdx)), nil)
}

// func (d *Decoder) Writer() *rwutil.Writer {
// 	return &d.w
// }

func (d *Decoder) Read(b []byte) (n int, err error) {
	d.r.ReadN(b)
	return len(b), d.r.Err
}

func Unmarshal[T any](b []byte) (T, error) {
	var t T
	if err := NewDecoder(bytes.NewReader(b), DecoderConfig{}).Decode(&t); err != nil {
		return t, err
	}

	return t, nil
}

func MustUnmarshal[T any](b []byte) T {
	v, err := Unmarshal[T](b)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal object of type %T from BCS: %w", v, err))
	}

	return v
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
