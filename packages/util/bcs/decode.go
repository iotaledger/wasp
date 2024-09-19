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

type Readable interface {
	Read(r io.Reader) error
}

var readableT = reflect.TypeOf((*Readable)(nil)).Elem()

type Initializeable interface {
	BCSInit() error
}

var initializeableT = reflect.TypeOf((*Initializeable)(nil)).Elem()

type DecoderConfig struct {
	TagName string
	//CustomDecoders map[reflect.Type]CustomDecoder
}

func (c *DecoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "bcs"
	}
}

func NewDecoder(src io.Reader) *Decoder {
	return NewDecoderWithOpts(src, DecoderConfig{})
}

func NewDecoderWithOpts(src io.Reader, cfg DecoderConfig) *Decoder {
	cfg.InitializeDefaults()

	return &Decoder{
		cfg:           cfg,
		r:             rwutil.NewReader(src),
		typeInfoCache: decoderTypeInfoCache.Get(),
	}
}

type Decoder struct {
	cfg           DecoderConfig
	r             *rwutil.Reader
	typeInfoCache localTypeInfoCache
}

var decoderTypeInfoCache = newGlobalTypeInfoCache()

func (d *Decoder) Err() error {
	return d.r.Err
}

func (d *Decoder) MustDecode(v any) {
	if err := d.Decode(v); err != nil {
		panic(err)
	}
}

func (d *Decoder) Decode(v any) error {
	vR := reflect.ValueOf(v)

	if vR.Kind() != reflect.Ptr {
		return d.handleErrorf("Decode destination must be a pointer")
	}
	if vR.IsNil() {
		return d.handleErrorf("Decode destination cannot be nil")
	}

	defer d.typeInfoCache.Save()

	if err := d.decodeValue(vR, nil, nil); err != nil {
		return d.handleErrorf("decoding %T: %w", v, err)
	}

	return nil
}

func (d *Decoder) DecodeOptional(v any) (bool, error) {
	if hasValue := d.r.ReadByte() != 0; !hasValue {
		return false, d.r.Err
	}

	return true, d.Decode(v)
}

func (d *Decoder) ReadOptionalFlag() bool {
	return d.r.ReadByte() != 0
}

// Enum index is an index of variant in enum type.
func (d *Decoder) ReadEnumIdx() int {
	return int(d.ReadCompactUint())
}

func (d *Decoder) ReadLen() int {
	return int(d.ReadCompactUint())
}

// ULEB - unsigned little-endian base-128 - variable-length integer value.
func (d *Decoder) ReadCompactUint() uint64 {
	return uint64(d.r.ReadSize32())
}

func (d *Decoder) ReadBool() bool {
	return d.r.ReadBool()
}

func (d *Decoder) ReadByte() byte {
	return d.r.ReadByte()
}

func (d *Decoder) ReadInt8() int8 {
	return d.r.ReadInt8()
}

func (d *Decoder) ReadInt16() int16 {
	return d.r.ReadInt16()
}

func (d *Decoder) ReadInt32() int32 {
	return d.r.ReadInt32()
}

func (d *Decoder) ReadInt64() int64 {
	return d.r.ReadInt64()
}

func (d *Decoder) ReadInt() int {
	return int(d.r.ReadInt64())
}

func (d *Decoder) ReadUint8() uint8 {
	return d.r.ReadUint8()
}

func (d *Decoder) ReadUint16() uint16 {
	return d.r.ReadUint16()
}

func (d *Decoder) ReadUint32() uint32 {
	return d.r.ReadUint32()
}

func (d *Decoder) ReadUint64() uint64 {
	return d.r.ReadUint64()
}

func (d *Decoder) ReadUint() uint {
	return uint(d.r.ReadUint64())
}

func (d *Decoder) ReadString() string {
	return d.r.ReadString()
}

func (d *Decoder) Read(b []byte) (n int, err error) {
	d.r.ReadN(b)
	return len(b), d.r.Err
}

// func (d *Decoder) Writer() *rwutil.Writer {
// 	return &d.w
// }

func (d *Decoder) decodeValue(v reflect.Value, typeOptionsFromTag *TypeOptions, tInfo *typeInfo) error {
	if tInfo == nil {
		// Hint about type customization could have been provided by caller when decoding collections.
		// This is done to avoid parsing type for each element of collection.
		// This is an optimization for decoding of large amount of small elements.
		t, err := d.getEncodedTypeInfo(v.Type())
		if err != nil {
			return err
		}

		tInfo = &t
	}

	v = d.getDecodedValueStorage(v, tInfo.RefLevelsCount)

	if tInfo.CustomDecoder != nil {
		if err := tInfo.CustomDecoder(d, v.Addr()); err != nil {
			if d.r.Err == nil {
				d.r.Err = err
			}
			return fmt.Errorf("%v: custom decoder: %w", v.Type(), err)
		}
		if d.r.Err != nil {
			return fmt.Errorf("%v: custom decoder: %w", v.Type(), d.r.Err)
		}

		if tInfo.Init != nil {
			if err := tInfo.Init(v.Addr()); err != nil {
				return d.handleErrorf("%v: custom init: %w", v.Type(), err)
			}
		}

		return nil
	}

	var typeOptions TypeOptions
	if tInfo.HasTypeOptions {
		typeOptions = v.Interface().(BCSType).BCSOptions()
	}
	if typeOptionsFromTag != nil {
		typeOptions.Update(*typeOptionsFromTag)
	}

	var err error

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(d.r.ReadBool())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if typeOptions.IsCompactInt {
			v.SetInt(int64(d.ReadCompactUint()))
		} else {
			err = d.decodeInt(v, defaultValueSize(v.Kind()), typeOptions.SizeInBytes)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if typeOptions.IsCompactInt {
			v.SetUint(d.ReadCompactUint())
		} else {
			err = d.decodeUint(v, defaultValueSize(v.Kind()), typeOptions.SizeInBytes)
		}
	case reflect.String:
		v.SetString(d.r.ReadString())
	case reflect.Slice:
		if typeOptions.ArrayElement == nil {
			typeOptions.ArrayElement = &ArrayElemOptions{}
		}
		err = d.decodeSlice(v, typeOptions)
	case reflect.Array:
		if typeOptions.ArrayElement == nil {
			typeOptions.ArrayElement = &ArrayElemOptions{}
		}
		err = d.decodeArray(v, typeOptions)
	case reflect.Map:
		if typeOptions.MapKey == nil {
			typeOptions.MapKey = &TypeOptions{}
		}
		if typeOptions.MapValue == nil {
			typeOptions.MapValue = &TypeOptions{}
		}
		err = d.decodeMap(v, typeOptions)
	case reflect.Struct:
		if tInfo.IsStructEnum {
			err = d.decodeStructEnum(v)
		} else {
			err = d.decodeStruct(v, tInfo)
		}
	case reflect.Interface:
		if typeOptions.InterfaceIsNotEnum {
			err = d.decodeInterface(v)
		} else {
			err = d.decodeInterfaceEnum(v)
		}
	default:
		return d.handleErrorf("%v: cannot decode unknown type", v.Type())
	}

	if err != nil {
		return fmt.Errorf("%v: %w", v.Type(), err)
	}
	if d.r.Err != nil {
		return fmt.Errorf("%v: %w", v.Type(), d.r.Err)
	}

	if tInfo.Init != nil {
		if err := tInfo.Init(v.Addr()); err != nil {
			return d.handleErrorf("%v: custom init: %w", v.Type(), err)
		}
	}

	return nil
}

func (e *Decoder) checkTypeCustomizations(t reflect.Type) typeCustomization {
	customDecoder := e.getCustomDecoder(t)
	customInitFunc := e.getCustomInitFunc(t)

	if customDecoder != nil || customInitFunc != nil {
		return typeCustomization{
			CustomDecoder: customDecoder,
			Init:          customInitFunc,
		}
	}

	kind := t.Kind()

	switch {
	case kind == reflect.Interface:
		return typeCustomization{}
	case kind == reflect.Struct && t.Implements(enumT):
		return typeCustomization{IsStructEnum: true}
	case t.Implements(bcsTypeT):
		return typeCustomization{HasTypeOptions: true}
	}

	return typeCustomization{}
}

func (e *Decoder) getEncodedTypeInfo(t reflect.Type) (typeInfo, error) {
	initialT := t

	if cached, isCached := e.typeInfoCache.Get(initialT); isCached {
		return cached, nil
	}

	// Removing all redundant pointers
	refLevelsCount := 0

	for t.Kind() == reflect.Ptr {
		// Before dereferencing pointer, we should check if maybe current type is already the type we should decode.
		customization := e.checkTypeCustomizations(t)
		if customization.HasCustomizations() {
			res := typeInfo{RefLevelsCount: refLevelsCount, typeCustomization: customization}
			e.typeInfoCache.Add(initialT, res)

			return res, nil
		}

		refLevelsCount++
		t = t.Elem()
	}

	customization := e.checkTypeCustomizations(t)

	res := typeInfo{RefLevelsCount: refLevelsCount, typeCustomization: customization}

	if t.Kind() == reflect.Struct {
		var err error
		res.FieldOptions, res.FieldHasTag, err = FieldOptionsFromStruct(t, e.cfg.TagName)
		if err != nil {
			return typeInfo{}, fmt.Errorf("parsing struct fields options: %w", err)
		}
	}

	e.typeInfoCache.Add(initialT, res)

	return res, nil
}

func (d *Decoder) getDecodedValueStorage(v reflect.Value, refLevelsCount int) (dereferenced reflect.Value) {
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
		// NOTE: Always creating new map even if it is not nil.
		// Collection should have exactly those elements that are encoded.
		v.Set(reflect.MakeMap(v.Type()))
	}

	return v
}

func (d *Decoder) getCustomDecoder(t reflect.Type) CustomDecoder {
	if customDecoder, ok := CustomDecoders[t]; ok {
		return customDecoder
	}

	if t.Kind() == reflect.Interface {
		return nil
	}

	if reflect.PointerTo(t).Implements(decodableT) {
		return func(e *Decoder, v reflect.Value) error {
			return v.Interface().(Decodable).UnmarshalBCS(e)
		}
	}

	if reflect.PointerTo(t).Implements(readableT) {
		return func(e *Decoder, v reflect.Value) error {
			return v.Interface().(Readable).Read(e)
		}
	}

	return nil
}

func (d *Decoder) getCustomInitFunc(t reflect.Type) InitFunc {
	if t.Kind() != reflect.Interface && reflect.PointerTo(t).Implements(initializeableT) {
		initFunc := func(v reflect.Value) error {
			return v.Interface().(Initializeable).BCSInit()
		}

		return initFunc
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
		return d.handleErrorf("invalid value size: %v", size)
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
		return d.handleErrorf("invalid value size: %v", size)
	}

	return nil
}

func (d *Decoder) decodeSlice(v reflect.Value, typOpts TypeOptions) error {
	var length int

	switch typOpts.LenSizeInBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes, 0:
		length = int(d.r.ReadSize32())
	default:
		return d.handleErrorf("invalid array size type: %v", typOpts.LenSizeInBytes)
	}

	if length == 0 {
		return nil
	}

	v.Set(reflect.MakeSlice(v.Type(), length, length))

	return d.decodeArray(v, typOpts)
}

func (d *Decoder) decodeArray(v reflect.Value, typOpts TypeOptions) error {
	elemType := v.Type().Elem()

	// We take pointer because elements of array are addressable and
	// decoder requires addressable value.
	tInfo, err := d.getEncodedTypeInfo(reflect.PointerTo(elemType))
	if err != nil {
		return fmt.Errorf("element: %w", err)
	}

	if !tInfo.HasCustomizations() {
		// Optimizations for decoding of basic types
		switch elemType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if err := d.decodeIntArray(v, defaultValueSize(elemType.Kind()), typOpts); err != nil {
				return fmt.Errorf("%v: %w", elemType, err)
			}

			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if err := d.decodeUintArray(v, defaultValueSize(elemType.Kind()), typOpts); err != nil {
				return fmt.Errorf("%v: %w", elemType, err)
			}

			return nil
		}
	}

	if typOpts.ArrayElement.AsByteArray {
		// Elements are encoded as byte arrays.
		for i := 0; i < v.Len(); i++ {
			err := d.decodeAsByteArray(func() error {
				return d.decodeValue(v.Index(i).Addr(), nil, &tInfo)
			})
			if err != nil {
				return fmt.Errorf("[%v]: %w", i, err)
			}
		}
	} else {
		for i := 0; i < v.Len(); i++ {
			if err := d.decodeValue(v.Index(i).Addr(), nil, &tInfo); err != nil {
				return fmt.Errorf("[%v]: %w", i, err)
			}
		}
	}

	return nil
}

func (d *Decoder) decodeIntArray(v reflect.Value, bytesPerElem ValueBytesCount, typOpts TypeOptions) error {
	if typOpts.ArrayElement.IsCompactInt {
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				d.decodeAsByteArray(func() error {
					v.Index(i).SetInt(int64(d.r.ReadSize32()))
					return nil
				})
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetInt(int64(d.r.ReadSize32()))
			}
		}

		return d.r.Err
	}

	switch bytesPerElem {
	case Value1Byte:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(1); err != nil {
					return err
				}
				v.Index(i).SetInt(int64(d.r.ReadInt8()))
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetInt(int64(d.r.ReadInt8()))
			}
		}
	case Value2Bytes:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(2); err != nil {
					return err
				}
				v.Index(i).SetInt(int64(d.r.ReadInt16()))
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetInt(int64(d.r.ReadInt16()))
			}
		}
	case Value4Bytes:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(4); err != nil {
					return err
				}
				v.Index(i).SetInt(int64(d.r.ReadInt32()))
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetInt(int64(d.r.ReadInt32()))
			}
		}
	case Value8Bytes:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(8); err != nil {
					return err
				}
				v.Index(i).SetInt(d.r.ReadInt64())
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetInt(d.r.ReadInt64())
			}
		}
	default:
		panic(fmt.Errorf("invalid value size: %v", bytesPerElem))
	}

	return d.r.Err
}

func (d *Decoder) decodeUintArray(v reflect.Value, bytesPerElem ValueBytesCount, typOpts TypeOptions) error {
	if typOpts.ArrayElement.IsCompactInt {
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				d.decodeAsByteArray(func() error {
					v.Index(i).SetUint(uint64(d.r.ReadSize32()))
					return nil
				})
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetUint(uint64(d.r.ReadSize32()))
			}
		}

		return d.r.Err
	}

	switch bytesPerElem {
	case Value1Byte:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(1); err != nil {
					return err
				}
				v.Index(i).SetUint(uint64(d.r.ReadUint8()))
			}
		} else {
			// Optimization for decoding of byte/uint8 slices
			b := make([]byte, v.Len())
			d.r.ReadN(b)

			if v.Kind() == reflect.Slice {
				v.SetBytes(b)
			} else {
				v.Set(reflect.ValueOf(b).Convert(v.Type()))
			}
		}
	case Value2Bytes:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(2); err != nil {
					return err
				}
				v.Index(i).SetUint(uint64(d.r.ReadUint16()))
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetUint(uint64(d.r.ReadUint16()))
			}
		}
	case Value4Bytes:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(4); err != nil {
					return err
				}
				v.Index(i).SetUint(uint64(d.r.ReadUint32()))
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetUint(uint64(d.r.ReadUint32()))
			}
		}
	case Value8Bytes:
		if typOpts.ArrayElement.AsByteArray {
			for i := 0; i < v.Len(); i++ {
				if err := d.readAndCheckByteArraySize(8); err != nil {
					return err
				}
				v.Index(i).SetUint(d.r.ReadUint64())
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetUint(d.r.ReadUint64())
			}
		}
	default:
		panic(fmt.Errorf("invalid value size: %v", bytesPerElem))
	}

	return d.r.Err
}

func (d *Decoder) readAndCheckByteArraySize(expected uint8) error {
	size := d.r.ReadUint8()
	if size != expected {
		return d.handleErrorf("invalid byte array length: expected %v, got %v", expected, size)
	}

	return nil
}

func (d *Decoder) decodeMap(v reflect.Value, typOpts TypeOptions) error {
	var length int

	switch typOpts.LenSizeInBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes, 0:
		length = int(d.r.ReadSize32())
	default:
		return d.handleErrorf("invalid map size type: %v", typOpts.LenSizeInBytes)
	}

	keyType := v.Type().Key()
	valueType := v.Type().Elem()

	keyTypeInfo, err := d.getEncodedTypeInfo(keyType)
	if err != nil {
		return fmt.Errorf("key: %w", err)
	}

	valueTypeInfo, err := d.getEncodedTypeInfo(valueType)
	if err != nil {
		return fmt.Errorf("value: %w", err)
	}

	for i := 0; i < length; i++ {
		key := reflect.New(keyType).Elem()
		value := reflect.New(valueType).Elem()

		if err := d.decodeValue(key, typOpts.MapKey, &keyTypeInfo); err != nil {
			return fmt.Errorf("key: %w", err)
		}

		if err := d.decodeValue(value, typOpts.MapValue, &valueTypeInfo); err != nil {
			return fmt.Errorf("value: %w", err)
		}

		v.SetMapIndex(key, value)
	}

	return nil
}

func (d *Decoder) decodeStruct(v reflect.Value, tInfo *typeInfo) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldOpts, hasTag := tInfo.FieldOptions[i], tInfo.FieldHasTag[i]

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
					// TODO: should we "clean" the field?
					// I'm not doing it to allow presetting it and keeping even if it was missing.
					continue
				}
			}
		}

		var err error

		if fieldOpts.AsByteArray {
			err = d.decodeAsByteArray(func() error {
				return d.decodeValue(fieldVal, &fieldOpts.TypeOptions, nil)
			})
		} else {
			err = d.decodeValue(fieldVal, &fieldOpts.TypeOptions, nil)
		}

		if err != nil {
			return fmt.Errorf("%v: %w", fieldType.Name, err)
		}
	}

	return nil
}

func (d *Decoder) decodeInterface(v reflect.Value) error {
	if v.IsNil() {
		return d.handleErrorf("cannot decode interface which is not enum and has nil value")
	}

	e := v.Elem()

	if e.Kind() == reflect.Ptr {
		return d.decodeValue(e, nil, nil)
	}

	// Interface is not nil and contains non-pointer value.
	// This means the value is not addressable - we cannot decode into it.
	// So we need to create a copy of the value and decode into it.
	eCopy := reflect.New(e.Type()).Elem()
	eCopy.Set(e)
	e = eCopy

	if err := d.decodeValue(e, nil, nil); err != nil {
		return err
	}

	v.Set(e)

	return nil
}

func (d *Decoder) decodeInterfaceEnum(v reflect.Value) error {
	variants, registered := EnumTypes[v.Type()]
	if !registered {
		return d.handleErrorf("interface type %v is not registered as enum", v.Type())
	}

	variantIdx := d.r.ReadSize32()
	if d.r.Err != nil {
		return d.r.Err
	}

	if int(variantIdx) >= len(variants) {
		return d.handleErrorf("invalid variant index %v for enum %v - enum has only %v variants", variantIdx, v.Type(), len(variants))
	}

	variantT := variants[variantIdx]
	if variantT == noneT {
		return nil
	}

	variant := reflect.New(variantT).Elem()

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
		return d.handleErrorf("invalid variant index %v for enum %v - enum has only %v variants", variantIdx, t, t.NumField())
	}

	return d.decodeValue(v.Field(int(variantIdx)), nil, nil)
}

func (d *Decoder) decodeAsByteArray(dec func() error) error {
	// This value was written as variable array of bytes.
	// Bytes of array are same as of value but they also have length prepended to them. So in theory we could just
	// skip length and continue reading. But that may result in confusing decoding errors in case of corrupted data.
	// So more reliable way is to separate those bytes and decode from them.

	b := make([]byte, int(d.r.ReadSize32()))
	d.r.ReadN(b)

	if d.r.Err != nil {
		return fmt.Errorf("bytearr: %w", d.r.Err)
	}

	origStream := d.r
	defer func() { d.r = origStream }() // for case of panic/error

	d.r = rwutil.NewBytesReader(b)

	return dec()
}

func (d *Decoder) handleErrorf(format string, args ...interface{}) error {
	d.r.Err = fmt.Errorf(format, args...)
	return d.r.Err
}

func Decode[V any](dec *Decoder) (V, error) {
	var v V
	err := dec.Decode(&v)

	return v, err
}

func MustDecode[V any](dec *Decoder) V {
	v, err := Decode[V](dec)
	if err != nil {
		panic(fmt.Errorf("failed to decode object of type %T: %w", v, err))
	}

	return v
}

func UnmarshalStream[T any](r io.Reader) (T, error) {
	dec := NewDecoder(r)
	return Decode[T](dec)
}

func MustUnmarshalStream[T any](r io.Reader) T {
	v, err := UnmarshalStream[T](r)
	if err != nil {
		panic(err)
	}

	return v
}

func Unmarshal[T any](b []byte) (T, error) {
	return UnmarshalStream[T](bytes.NewReader(b))
}

func MustUnmarshal[T any](b []byte) T {
	v, err := Unmarshal[T](b)
	if err != nil {
		panic(err)
	}

	return v
}

func UnmarshalStreamInto[V any](r io.Reader, v *V) (*V, error) {
	if err := NewDecoder(r).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

func MustUnmarshalStreamInto[V any](r io.Reader, v *V) *V {
	if _, err := UnmarshalStreamInto(r, v); err != nil {
		panic(err)
	}

	return v
}

func UnmarshalInto[V any](b []byte, v *V) (*V, error) {
	return UnmarshalStreamInto(bytes.NewReader(b), v)
}

func MustUnmarshalInto[V any](b []byte, v *V) *V {
	if _, err := UnmarshalInto(b, v); err != nil {
		panic(err)
	}

	return v
}

type CustomDecoder func(e *Decoder, v reflect.Value) error
type InitFunc func(v reflect.Value) error

var CustomDecoders = make(map[reflect.Type]CustomDecoder)

func MakeCustomDecoder[V any](f func(e *Decoder, v *V) error) func(e *Decoder, v reflect.Value) error {
	return func(e *Decoder, v reflect.Value) error {
		return f(e, v.Interface().(*V))
	}
}

func AddCustomDecoder[V any](f func(e *Decoder, v *V) error) {
	CustomDecoders[reflect.TypeOf((*V)(nil)).Elem()] = MakeCustomDecoder(f)
}

func RemoveCustomDecoder[V any]() {
	delete(CustomDecoders, reflect.TypeOf((*V)(nil)).Elem())
}
