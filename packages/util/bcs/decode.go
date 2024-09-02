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

	return d.decodeValue(vR, nil, nil)
}

func (d *Decoder) decodeValue(v reflect.Value, typeOptionsFromTag *TypeOptions, typeParsingHint *typeInfo) error {
	var t typeInfo

	if typeParsingHint != nil {
		// Hint about type customization is provided by caller when decoding collections.
		// This is done to avoid parsing type for each element of collection.
		// This is an optimization for decoding of large amount of small elements.
		// Otherwise even elements of collection of custom int8-based type each would require parsing of type.
		t = *typeParsingHint
	} else {
		t = d.getEncodedType(v.Type())
	}

	v = d.getEncodedValue(v, t.RefLevelsCount)

	if t.Customization == typeCustomizationHasCustomCodec {
		if err := t.CustomDecoder(d, v.Addr()); err != nil {
			return fmt.Errorf("%v: custom decoder: %w", v.Type(), err)
		}

		return nil
	}

	var typeOptions TypeOptions
	if t.Customization == typeCustomizationHasTypeOptions {
		typeOptions = v.Interface().(BCSType).BCSOptions()
	}
	if typeOptionsFromTag != nil {
		typeOptions.Update(*typeOptionsFromTag)
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(d.r.ReadBool())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if err := d.decodeInt(v, defaultValueSize(v.Kind()), typeOptions.Bytes); err != nil {
			return fmt.Errorf("%v: %w", v.Type(), err)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if err := d.decodeUint(v, defaultValueSize(v.Kind()), typeOptions.Bytes); err != nil {
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
		if t.Customization == typeCustomizationIsStructEnum {
			if err := d.decodeStructEnum(v); err != nil {
				return fmt.Errorf("%v: %w", v.Type(), err)
			}
		} else {
			if err := d.decodeStruct(v); err != nil {
				return fmt.Errorf("%v: %w", v.Type(), err)
			}
		}
	case reflect.Interface:
		if t.Customization != typeCustomizationIsInterfaceEnum {
			panic(fmt.Errorf("unexpected type customization %v", v.Type()))
		}

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

func (e *Decoder) checkTypeCustomizations(t reflect.Type) (typeCustomization, CustomDecoder) {
	// Detecting enum variant index might return error, so we
	// should first check for existance of custom decoder.
	if customDecoder := e.getCustomDecoder(t); customDecoder != nil {
		return typeCustomizationHasCustomCodec, customDecoder
	}

	kind := t.Kind()

	switch {
	case kind == reflect.Interface:
		return typeCustomizationIsInterfaceEnum, nil
	case kind == reflect.Struct && t.Implements(enumT):
		return typeCustomizationIsStructEnum, nil
	case t.Implements(bcsTypeT):
		return typeCustomizationHasTypeOptions, nil
	}

	return typeCustomizationNone, nil
}

func (e *Decoder) getEncodedType(t reflect.Type) typeInfo {
	// Removing all redundant pointers
	refLevelsCount := 0

	for t.Kind() == reflect.Ptr {
		// Before dereferencing pointer, we should check if maybe current type is already the type we should decode.
		customization, customDecoder := e.checkTypeCustomizations(t)
		if customization != typeCustomizationNone {
			return typeInfo{refLevelsCount, customization, nil, customDecoder}
		}

		refLevelsCount++
		t = t.Elem()
	}

	customization, customDecoder := e.checkTypeCustomizations(t)

	return typeInfo{refLevelsCount, customization, nil, customDecoder}
}

func (d *Decoder) getEncodedValue(v reflect.Value, refLevelsCount int) (dereferenced reflect.Value) {
	// Getting rid of found redundant pointers AND creating a new value to be able to set it.

	for i := 0; i < refLevelsCount; i++ {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
	case reflect.Map:
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
	}

	return v
}

func (d *Decoder) getCustomDecoder(t reflect.Type) CustomDecoder {
	if customDecoder, ok := d.cfg.CustomDecoders[t]; ok {
		return customDecoder
	}

	// We skip interfaces, because although they can have custom decoders set for them as global option,
	// they still can't providate them through methods, because their actual type is unknown.
	if t.Kind() != reflect.Interface && reflect.PointerTo(t).Implements(decodableT) {
		customDecoder := func(e *Decoder, v reflect.Value) error {
			return v.Interface().(Decodable).UnmarshalBCS(e)
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

	v.Set(reflect.MakeSlice(v.Type(), length, length))

	return d.decodeArray(v)
}

func (d *Decoder) decodeArray(v reflect.Value) error {
	elemType := v.Type().Elem()

	// We take pointer because elements of array are addressable and
	// decoder requires addressable value.
	t := d.getEncodedType(reflect.PointerTo(elemType))

	if t.Customization == typeCustomizationNone {
		// Optimizations for decoding of basic types
		switch elemType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if err := d.decodeIntArray(v, defaultValueSize(elemType.Kind())); err != nil {
				return fmt.Errorf("%v: %w", elemType, err)
			}

			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if err := d.decodeUintArray(v, defaultValueSize(elemType.Kind())); err != nil {
				return fmt.Errorf("%v: %w", elemType, err)
			}

			return nil
		}
	}

	for i := 0; i < v.Len(); i++ {
		if err := d.decodeValue(v.Index(i).Addr(), nil, &t); err != nil {
			return fmt.Errorf("[%v]: %w", i, err)
		}
	}

	return nil
}

func (d *Decoder) decodeIntArray(v reflect.Value, bytesPerElem ValueBytesCount) error {
	switch bytesPerElem {
	case Value1Byte:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetInt(int64(d.r.ReadInt8()))
		}
	case Value2Bytes:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetInt(int64(d.r.ReadInt16()))
		}
	case Value4Bytes:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetInt(int64(d.r.ReadInt32()))
		}
	case Value8Bytes:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetInt(d.r.ReadInt64())
		}
	default:
		panic(fmt.Errorf("invalid value size: %v", bytesPerElem))
	}

	return d.r.Err
}

func (d *Decoder) decodeUintArray(v reflect.Value, bytesPerElem ValueBytesCount) error {
	switch bytesPerElem {
	case Value1Byte:
		// Optimization for decoding of byte/uint8 slices
		b := make([]byte, v.Len())
		d.r.ReadN(b)
		if v.Kind() == reflect.Slice {
			v.SetBytes(b)
		} else {
			v.Set(reflect.ValueOf(b).Convert(v.Type()))
		}
	case Value2Bytes:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetUint(uint64(d.r.ReadUint16()))
		}
	case Value4Bytes:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetUint(uint64(d.r.ReadUint32()))
		}
	case Value8Bytes:
		for i := 0; i < v.Len(); i++ {
			v.Index(i).SetUint(d.r.ReadUint64())
		}
	default:
		panic(fmt.Errorf("invalid value size: %v", bytesPerElem))
	}

	return d.r.Err
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

	keyTypeDeref := d.getEncodedType(keyType)
	valueTypeDeref := d.getEncodedType(valueType)

	for i := 0; i < length; i++ {
		key := reflect.New(keyType).Elem()
		value := reflect.New(valueType).Elem()

		if err := d.decodeValue(key, nil, &keyTypeDeref); err != nil {
			return fmt.Errorf("key: %w", err)
		}

		if err := d.decodeValue(value, nil, &valueTypeDeref); err != nil {
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

		if err := d.decodeValue(fieldVal, &fieldOpts.TypeOptions, nil); err != nil {
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

	if err := d.decodeValue(variant, nil, nil); err != nil {
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

	return d.decodeValue(v.Field(int(variantIdx)), nil, nil)
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

type CustomDecoder func(e *Decoder, v reflect.Value) error

var CustomDecoders = make(map[reflect.Type]CustomDecoder)

func MakeCustomDecoder[V any](f func(e *Decoder, v *V) error) func(e *Decoder, v reflect.Value) error {
	return func(e *Decoder, v reflect.Value) error {
		return f(e, v.Interface().(*V))
	}
}

func AddCustomDecoder[V any](f func(e *Decoder, v *V) error) {
	CustomDecoders[reflect.TypeOf(lo.Empty[V]())] = MakeCustomDecoder(f)
}
