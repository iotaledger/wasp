package bcs

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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
	TagName                  string
	InterfaceIsEnumByDefault bool
	// CustomDecoders map[reflect.Type]CustomDecoder
}

func (c *DecoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "bcs"
	}
}

func NewBytesDecoder(b []byte) *BytesDecoder {
	r := bytes.NewReader(b)
	return &BytesDecoder{
		Decoder: *NewDecoder(r),
		buf:     r,
		b:       b,
	}
}

type BytesDecoder struct {
	Decoder
	buf *bytes.Reader
	b   []byte
}

// Size returns the original length of the underlying byte slice.
// The result is unaffected by any method calls.
func (d *BytesDecoder) Size() int {
	return int(d.buf.Size())
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (d *BytesDecoder) Len() int {
	return d.buf.Len()
}

// Pos returns the current position in the underlying slice.
// Unread portion of the slice starts at this position.
func (d *BytesDecoder) Pos() int {
	return int(d.buf.Size()) - d.Len()
}

// Leftovers returns the unread portion of the slice.
func (d *BytesDecoder) Leftovers() []byte {
	return d.b[d.Pos():len(d.b):len(d.b)]
}

func NewDecoder(src io.Reader) *Decoder {
	return NewDecoderWithOpts(src, DecoderConfig{})
}

func NewDecoderWithOpts(src io.Reader, cfg DecoderConfig) *Decoder {
	cfg.InitializeDefaults()

	return &Decoder{
		cfg:           cfg,
		r:             rwutil.NewReader(src),
		typeInfoCache: decoderGlobalTypeInfoCache.Get(),
	}
}

type Decoder struct {
	cfg           DecoderConfig
	r             *rwutil.Reader
	typeInfoCache localTypeInfoCache
}

var decoderGlobalTypeInfoCache = newSharedTypeInfoCache()

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
	hasValue := d.ReadOptionalFlag()
	if d.r.Err != nil || !hasValue {
		return false, d.r.Err
	}

	return true, d.Decode(v)
}

func (d *Decoder) ReadOptionalFlag() bool {
	f := d.r.ReadByte()
	switch f {
	case 0:
		return false
	case 1:
		return true
	default:
		d.r.Err = fmt.Errorf("invalid optional flag value: %v", f)
		return false
	}
}

// Enum index is an index of variant in enum type.
func (d *Decoder) ReadEnumIdx() int {
	return int(d.r.ReadSize32())
}

func (d *Decoder) ReadLen() int {
	return int(d.r.ReadSize32())
}

// ULEB - unsigned little-endian base-128 - variable-length integer value.
func (d *Decoder) ReadCompactUint() uint64 {
	return d.r.ReadAmount64()
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
	d.r.ReadFromFunc(func(r io.Reader) (int, error) {
		n, err = r.Read(b)
		return n, err
	})

	return n, d.r.Err
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
			err = d.decodeInt(v, typeOptions.UnderlayingType)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if typeOptions.IsCompactInt {
			v.SetUint(d.ReadCompactUint())
		} else {
			err = d.decodeUint(v, typeOptions.UnderlayingType)
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
		err = d.decodeInterface(v, !typeOptions.InterfaceIsNotEnum)
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

func (d *Decoder) checkTypeCustomizations(t reflect.Type) typeCustomization {
	customDecoder := d.getCustomDecoder(t)
	customInitFunc := d.getCustomInitFunc(t)

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

func (d *Decoder) getEncodedTypeInfo(t reflect.Type) (typeInfo, error) {
	initialT := t

	if cached, isCached := d.typeInfoCache.Get(initialT); isCached {
		return cached, nil
	}

	// Removing all redundant pointers
	refLevelsCount := 0

	for t.Kind() == reflect.Ptr {
		// Before dereferencing pointer, we should check if maybe current type is already the type we should decode.
		customization := d.checkTypeCustomizations(t)
		if customization.HasCustomizations() {
			res := typeInfo{RefLevelsCount: refLevelsCount, typeCustomization: customization}
			d.typeInfoCache.Add(initialT, res)

			return res, nil
		}

		refLevelsCount++
		t = t.Elem()
	}

	customization := d.checkTypeCustomizations(t)

	res := typeInfo{RefLevelsCount: refLevelsCount, typeCustomization: customization}

	if t.Kind() == reflect.Struct {
		var err error
		res.FieldOptions, res.FieldHasTag, err = FieldOptionsFromStruct(t, d.cfg.TagName)
		if err != nil {
			return typeInfo{}, fmt.Errorf("parsing struct fields options: %w", err)
		}
	}

	d.typeInfoCache.Add(initialT, res)

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

func (d *Decoder) decodeInt(v reflect.Value, encodedType reflect.Kind) error {
	k := v.Kind()

	if encodedType != reflect.Invalid && encodedType != k {
		return decodeConvertNumber(d, v, encodedType)
	}

	switch k {
	case reflect.Int8:
		v.SetInt(int64(d.r.ReadInt8()))
	case reflect.Int16:
		v.SetInt(int64(d.r.ReadInt16()))
	case reflect.Int32:
		v.SetInt(int64(d.r.ReadInt32()))
	case reflect.Int64, reflect.Int:
		v.SetInt(d.r.ReadInt64())
	default:
		panic(fmt.Sprintf("unexpected int kind: %v", k))
	}

	return nil
}

func (d *Decoder) decodeUint(v reflect.Value, encodedType reflect.Kind) error {
	k := v.Kind()

	if encodedType != reflect.Invalid && encodedType != k {
		return decodeConvertNumber(d, v, encodedType)
	}

	switch k {
	case reflect.Uint8:
		v.SetUint(uint64(d.r.ReadUint8()))
	case reflect.Uint16:
		v.SetUint(uint64(d.r.ReadUint16()))
	case reflect.Uint32:
		v.SetUint(uint64(d.r.ReadUint32()))
	case reflect.Uint64, reflect.Uint:
		v.SetUint(d.r.ReadUint64())
	default:
		panic(fmt.Sprintf("unexpected uint kind: %v", k))
	}

	return nil
}

func decodeConvertNumber(d *Decoder, v reflect.Value, encodedType reflect.Kind) error {
	switch v.Kind() {
	case reflect.Int8:
		return decodeConvertNumber2(d, encodedType, func(i int8) { v.SetInt(int64(i)) })
	case reflect.Int16:
		return decodeConvertNumber2(d, encodedType, func(i int16) { v.SetInt(int64(i)) })
	case reflect.Int32:
		return decodeConvertNumber2(d, encodedType, func(i int32) { v.SetInt(int64(i)) })
	case reflect.Int64, reflect.Int:
		return decodeConvertNumber2(d, encodedType, v.SetInt)
	case reflect.Uint8:
		return decodeConvertNumber2(d, encodedType, func(u uint8) { v.SetUint(uint64(u)) })
	case reflect.Uint16:
		return decodeConvertNumber2(d, encodedType, func(u uint16) { v.SetUint(uint64(u)) })
	case reflect.Uint32:
		return decodeConvertNumber2(d, encodedType, func(u uint32) { v.SetUint(uint64(u)) })
	case reflect.Uint64, reflect.Uint:
		return decodeConvertNumber2(d, encodedType, v.SetUint)
	default:
		panic(fmt.Sprintf("unexpected number kind: %v", v.Kind()))
	}
}

// decodeConvertNumber2 is a helper function for decodeConvertNumber, which is used to unwrap type RealType.
func decodeConvertNumber2[RealType constraints.Numeric](d *Decoder, encodedType reflect.Kind, set func(RealType)) error {
	switch encodedType {
	case reflect.Int8:
		return decodeConvertNumber3(d, d.r.ReadInt8, set)
	case reflect.Int16:
		return decodeConvertNumber3(d, d.r.ReadInt16, set)
	case reflect.Int32:
		return decodeConvertNumber3(d, d.r.ReadInt32, set)
	case reflect.Int64, reflect.Int:
		return decodeConvertNumber3(d, d.r.ReadInt64, set)
	case reflect.Uint8:
		return decodeConvertNumber3(d, d.r.ReadUint8, set)
	case reflect.Uint16:
		return decodeConvertNumber3(d, d.r.ReadUint16, set)
	case reflect.Uint32:
		return decodeConvertNumber3(d, d.r.ReadUint32, set)
	case reflect.Uint64, reflect.Uint:
		return decodeConvertNumber3(d, d.r.ReadUint64, set)
	default:
		return d.handleErrorf("invalid underlaying type %v for type %T", encodedType, lo.Empty[RealType]())
	}
}

// decodeConvertNumber3 is a helper function for decodeConvertNumber2, which is used to convert value to From.
func decodeConvertNumber3[To, From constraints.Numeric](d *Decoder, read func() From, set func(To)) error {
	v := read()
	converted := To(v)

	if From(converted) != v {
		return d.handleErrorf("value %v is out of range of type %T", v, To(0))
	}

	set(converted)

	return nil
}

func (d *Decoder) decodeSlice(v reflect.Value, typeOpts TypeOptions) error {
	var length int

	switch typeOpts.LenSizeInBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes, 0:
		length = int(d.r.ReadSize32())
	default:
		return d.handleErrorf("invalid array size type: %v", typeOpts.LenSizeInBytes)
	}

	if length == 0 {
		return nil
	}

	v.Set(reflect.MakeSlice(v.Type(), length, length))

	return d.decodeArray(v, typeOpts)
}

func (d *Decoder) decodeArray(v reflect.Value, typeOpts TypeOptions) error {
	elemType := v.Type().Elem()

	// We take pointer because elements of array are addressable and
	// decoder requires addressable value.
	tInfo, err := d.getEncodedTypeInfo(reflect.PointerTo(elemType))
	if err != nil {
		return fmt.Errorf("element: %w", err)
	}

	if !tInfo.HasCustomizations() {
		// The type does not have any customizations. So we can use  some optimizations for encoding of basic types
		if elemType.Kind() == reflect.Uint8 && (v.Kind() == reflect.Slice || v.CanAddr()) && !typeOpts.ArrayElement.AsByteArray {
			// Optimization for []byte and [N]byte.
			d.r.ReadN(v.Bytes())
			return nil
		}

		// There could be other optimizations for encoding of basic types. But I removed them for now for simplicity.
	}

	if typeOpts.ArrayElement.AsByteArray {
		// Elements were encoded as byte arrays.
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

func (d *Decoder) decodeMap(v reflect.Value, typeOpts TypeOptions) error {
	var length int

	switch typeOpts.LenSizeInBytes {
	case Len2Bytes:
		length = int(d.r.ReadSize16())
	case Len4Bytes, 0:
		length = int(d.r.ReadSize32())
	default:
		return d.handleErrorf("invalid map size type: %v", typeOpts.LenSizeInBytes)
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

		if err := d.decodeValue(key, typeOpts.MapKey, &keyTypeInfo); err != nil {
			return fmt.Errorf("key: %w", err)
		}

		if err := d.decodeValue(value, typeOpts.MapValue, &valueTypeInfo); err != nil {
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
			if !fieldOpts.ExportAnonymousField {
				if hasTag {
					return d.handleErrorf("%v: unexported field %v has tag, but is not marked for export", t.Name(), fieldType.Name)
				}

				// Unexported fields are skipped by default if not explicitly marked as exported
				continue
			}

			// The field is unexported, but it has a tag, so we need to serialize it.
			// Trick to access unexported fields: https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields/43918797#43918797
			fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		} else if fieldOpts.ExportAnonymousField {
			return d.handleErrorf("%v: field %v is already exported, but is marked for export", t.Name(), fieldType.Name)
		}

		fieldKind := fieldVal.Kind()

		if fieldKind == reflect.Ptr || fieldKind == reflect.Interface || fieldKind == reflect.Map {
			if fieldOpts.Optional {
				hasValue := d.ReadOptionalFlag()
				if d.r.Err != nil {
					return fmt.Errorf("%v: %w", fieldType.Name, d.r.Err)
				}

				if !hasValue {
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

func (d *Decoder) decodeInterface(v reflect.Value, couldBeEnum bool) error {
	if couldBeEnum {
		variants, registered := EnumTypes[v.Type()]

		if registered {
			return d.decodeInterfaceEnum(v, variants)
		}
		if d.cfg.InterfaceIsEnumByDefault {
			return d.handleErrorf("interface type %v is not registered as enum", v.Type())
		}
	}

	if v.IsNil() {
		return d.handleErrorf("cannot decode interface which is not enum and has nil value")
	}

	e := v.Elem()

	if e.Kind() == reflect.Ptr {
		if e.IsNil() {
			v.Set(reflect.New(e.Type().Elem()))
			e = v.Elem()
		}

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

func (d *Decoder) decodeInterfaceEnum(v reflect.Value, variants map[int]reflect.Type) error {
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

func Decode[V any](dec *Decoder) V {
	var v V
	err := dec.Decode(&v)
	if err != nil {
		panic(err)
	}
	return v
}

func MustDecode[V any](dec *Decoder) V {
	v := Decode[V](dec)
	if err := dec.Err(); err != nil {
		panic(fmt.Errorf("failed to decode object of type %T: %w", v, err))
	}

	return v
}

func UnmarshalStream[T any](r io.Reader) (T, error) {
	dec := NewDecoder(r)
	v := Decode[T](dec)
	return v, dec.Err()
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

type (
	CustomDecoder func(e *Decoder, v reflect.Value) error
	InitFunc      func(v reflect.Value) error
)

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
