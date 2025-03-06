package bcs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/constraints"
)

func UnmarshalStream[T any](r io.Reader) (T, error) {
	var v T
	_, err := UnmarshalStreamInto[T](r, &v)
	return v, err
}

func MustUnmarshalStream[T any](r io.Reader) T {
	v, err := UnmarshalStream[T](r)
	if err != nil {
		panic(err)
	}

	return v
}

func Unmarshal[T any](b []byte) (T, error) {
	var v T
	_, err := UnmarshalInto(b, &v)
	return v, err
}

func MustUnmarshal[T any](b []byte) T {
	v, err := Unmarshal[T](b)
	if err != nil {
		panic(err)
	}

	return v
}

func UnmarshalStreamInto[V any](r io.Reader, v *V) (*V, error) {
	d := NewDecoder(r)
	d.Decode(v)
	if d.err != nil {
		return nil, d.err
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
	r := bytes.NewReader(b)
	v, err := UnmarshalStreamInto(r, v)
	if err != nil {
		return nil, err
	}

	if r.Len() > 0 {
		return nil, fmt.Errorf("excess bytes: %v", r.Len())
	}

	return v, nil
}

func MustUnmarshalInto[V any](b []byte, v *V) *V {
	if _, err := UnmarshalInto(b, v); err != nil {
		panic(err)
	}

	return v
}

// Decode() is a helper function to decode single value from stream.
// It is convenient to use when you don't yet have a variable to store decoded value.

// If error occurs, it will be stored inside of decoder and can be checked using dec.Err().
// Further calls to Decode() will just fail with same error and do nothing.
// So no need to check error every time.
// Example:
//
//	v1 := Decode[int](dec)
//	v2 := Decode[string](dec)
//	v3 := Decode[bool](dec)
//
//	if dec.Err() != nil {
//	    return dec.Err()
//	}
//
// If Decode() is called inside of UnmarshalBCS() method, you can even skip checking dec.Err(),
// because decoder itself will do it for you anyway.
// Example:
//
//	func (p *MyStruct) UnmarshalBCS(d *bcs.Decoder) error {
//	    p.Field1 = Decode[int](d)
//	    p.Field2 = Decode[string](d)
//	    return nil
//	}
func Decode[V any](dec *Decoder) V {
	var v V
	dec.Decode(&v)

	return v
}

func MustDecode[V any](dec *Decoder) V {
	v := Decode[V](dec)
	if dec.err != nil {
		panic(fmt.Errorf("failed to decode object of type %T: %w", v, dec.err))
	}

	return v
}

type Decodable interface {
	UnmarshalBCS(e *Decoder) error
}

type Readable interface {
	Read(r io.Reader) error
}

type Initializeable interface {
	BCSInit() error
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
		r:             src,
		typeInfoCache: decoderGlobalTypeInfoCache.Get(),
	}
}

type Decoder struct {
	cfg           DecoderConfig
	r             io.Reader
	err           error
	typeInfoCache localTypeInfoCache
}

func (d *Decoder) Err() error {
	return d.err
}

func (d *Decoder) MustDecode(v any) {
	d.Decode(v)
	if d.err != nil {
		panic(d.err)
	}
}

// If error occurs, it will be stored inside of decoder and can be checked using dec.Err().
// After error further calls to Decode() will just do nothing.
// So no need to check error every time.
// Example:
//
//	dec.Decode(&v1)
//	dec.Decode(&v2)
//	dec.Decode(&v3)
//
//	if err := dec.Err(); err != nil {
//	    return err
//	}
//
// If Decode() is called inside of UnmarshalBCS() method, you can even skip checking dec.Err(),
// because decoder itself will do it for you anyway.
// Example:
//
//	func (p *MyStruct) UnmarshalBCS(d *bcs.Decoder) error {
//	    d.Decode(&p.Field1)
//	    d.Decode(&p.Field2)
//	    return nil
//	}
func (d *Decoder) Decode(v any) {
	if d.err != nil {
		return
	}

	vR := reflect.ValueOf(v)

	if vR.Kind() != reflect.Ptr {
		_ = d.handleErrorf("Decode destination must be a pointer")
		return
	}
	if vR.IsNil() {
		_ = d.handleErrorf("Decode destination cannot be nil")
		return
	}

	defer d.typeInfoCache.Save()

	if err := d.decodeValue(vR, nil, nil); err != nil {
		_ = d.handleErrorf("decoding %T: %w", v, err)
		return
	}
}

func (d *Decoder) DecodeOptional(v any) bool {
	hasValue := d.ReadOptionalFlag()
	if d.err != nil || !hasValue {
		return false
	}

	d.Decode(v)

	return true
}

func (d *Decoder) ReadOptionalFlag() bool {
	if d.err != nil {
		return false
	}

	f := d.ReadByte()
	switch f {
	case 0:
		return false
	case 1:
		return true
	default:
		_ = d.handleErrorf("invalid optional flag value: %v", f)
		return false
	}
}

// Enum index is an index of variant in enum type.
func (d *Decoder) ReadEnumIdx() int {
	return int(d.ReadCompactUint64())
}

func (d *Decoder) ReadLen() int {
	return int(d.ReadCompactUint64())
}

func (d *Decoder) ReadCompactUint64() uint64 {
	// ULEB - unsigned little-endian base-128 - variable-length integer value.

	b, err := d.readByte()
	if err != nil {
		return 0
	}
	if b < 0x80 {
		return uint64(b)
	}
	value := uint64(b & 0x7f)

	for shift := 7; shift < 63; shift += 7 {
		b, err = d.readByte()
		if err != nil {
			return 0
		}
		if b < 0x80 {
			return value | (uint64(b) << shift)
		}
		value |= uint64(b&0x7f) << shift
	}

	b, err = d.readByte()
	if err != nil {
		return 0
	}

	// must be the final bit (since we already encoded 63 bits)
	if b > 0x01 {
		_ = d.handleErrorf("compact uint64 overflow")
		return 0
	}

	return value | (uint64(b) << 63)
}

func (d *Decoder) readByte() (byte, error) {
	var b [1]byte
	_, err := d.Read(b[:])
	return b[0], err
}

func (d *Decoder) ReadBool() bool {
	b := d.ReadByte()
	switch b {
	case 0:
		return false
	case 1:
		return true
	default:
		_ = d.handleErrorf("invalid bool value: %v", b)
		return false
	}
}

//nolint:govet
func (d *Decoder) ReadByte() byte {
	b, _ := d.readByte()
	return b
}

func (d *Decoder) ReadInt8() int8 {
	return int8(d.ReadByte())
}

func (d *Decoder) ReadUint8() uint8 {
	return uint8(d.ReadByte())
}

func (d *Decoder) ReadInt16() int16 {
	return int16(d.ReadUint16())
}

func (d *Decoder) ReadUint16() uint16 {
	var b [2]byte
	if _, err := d.Read(b[:]); err != nil {
		return 0
	}

	return uint16(b[0]) | (uint16(b[1]) << 8)
}

func (d *Decoder) ReadInt32() int32 {
	return int32(d.ReadUint32())
}

func (d *Decoder) ReadUint32() uint32 {
	var b [4]byte
	if _, err := d.Read(b[:]); err != nil {
		return 0
	}

	return binary.LittleEndian.Uint32(b[:])
}

func (d *Decoder) ReadInt64() int64 {
	return int64(d.ReadUint64())
}

func (d *Decoder) ReadUint64() uint64 {
	var b [8]byte
	if _, err := d.Read(b[:]); err != nil {
		return 0
	}

	return binary.LittleEndian.Uint64(b[:])
}

func (d *Decoder) ReadInt() int {
	return int(d.ReadInt64())
}

func (d *Decoder) ReadUint() uint {
	return uint(d.ReadUint64())
}

func (d *Decoder) ReadString() string {
	length := d.ReadLen()
	if length == 0 {
		return ""
	}

	b := make([]byte, length)
	if _, err := d.Read(b); err != nil {
		return ""
	}

	return string(b)
}

func (d *Decoder) Read(b []byte) (n int, _ error) {
	if d.err != nil {
		return 0, d.err
	}

	n, d.err = d.r.Read(b)

	return n, d.err
}

//nolint:gocyclo,funlen
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
			if d.err == nil {
				d.err = err
			}
			return d.handleErrorf("%v: custom decoder: %w", v.Type(), err)
		}
		if d.err != nil {
			return d.handleErrorf("%v: custom decoder: %w", v.Type(), d.err)
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
		v.SetBool(d.ReadBool())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if typeOptions.IsCompactInt {
			v.SetInt(int64(d.ReadCompactUint64())) //nolint:gosec
		} else {
			err = d.decodeInt(v, typeOptions.UnderlyingType)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if typeOptions.IsCompactInt {
			v.SetUint(d.ReadCompactUint64())
		} else {
			err = d.decodeUint(v, typeOptions.UnderlyingType)
		}
	case reflect.String:
		v.SetString(d.ReadString())
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
		return d.handleErrorf("%v: %w", v.Type(), err)
	}
	if d.err != nil {
		return d.handleErrorf("%v: %w", v.Type(), d.err)
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
	case kind == reflect.Struct && t.Implements(structEnumT):
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
			return typeInfo{}, d.handleErrorf("parsing struct fields options: %w", err)
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
		v.SetInt(int64(d.ReadInt8()))
	case reflect.Int16:
		v.SetInt(int64(d.ReadInt16()))
	case reflect.Int32:
		v.SetInt(int64(d.ReadInt32()))
	case reflect.Int64, reflect.Int:
		v.SetInt(d.ReadInt64())
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
		v.SetUint(uint64(d.ReadUint8()))
	case reflect.Uint16:
		v.SetUint(uint64(d.ReadUint16()))
	case reflect.Uint32:
		v.SetUint(uint64(d.ReadUint32()))
	case reflect.Uint64, reflect.Uint:
		v.SetUint(d.ReadUint64())
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
		return decodeConvertNumber3(d, d.ReadInt8, set)
	case reflect.Int16:
		return decodeConvertNumber3(d, d.ReadInt16, set)
	case reflect.Int32:
		return decodeConvertNumber3(d, d.ReadInt32, set)
	case reflect.Int64, reflect.Int:
		return decodeConvertNumber3(d, d.ReadInt64, set)
	case reflect.Uint8:
		return decodeConvertNumber3(d, d.ReadUint8, set)
	case reflect.Uint16:
		return decodeConvertNumber3(d, d.ReadUint16, set)
	case reflect.Uint32:
		return decodeConvertNumber3(d, d.ReadUint32, set)
	case reflect.Uint64, reflect.Uint:
		return decodeConvertNumber3(d, d.ReadUint64, set)
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
	length := d.ReadLen()

	switch typeOpts.LenSizeInBytes {
	case 0:
	case Len2Bytes:
		if length > 0xFFFF {
			return d.handleErrorf("array size exceeds 2 bytes: %v", length)
		}
	case Len4Bytes:
		if length > 0xFFFFFFFF {
			return d.handleErrorf("array size exceeds 4 bytes: %v", length)
		}
	default:
		return d.handleErrorf("invalid array size type: %v", typeOpts.LenSizeInBytes)
	}

	if length == 0 {
		if !typeOpts.NilIfEmpty {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}

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
		return d.handleErrorf("element: %w", err)
	}

	if !tInfo.HasCustomizations() {
		// The type does not have any customizations. So we can use  some optimizations for encoding of basic types
		if elemType.Kind() == reflect.Uint8 && (v.Kind() == reflect.Slice || v.CanAddr()) && !typeOpts.ArrayElement.AsByteArray {
			// Optimization for []byte and [N]byte.
			_, _ = d.Read(v.Bytes())
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
				return d.handleErrorf("[%v]: %w", i, err)
			}
		}
	} else {
		for i := 0; i < v.Len(); i++ {
			if err := d.decodeValue(v.Index(i).Addr(), nil, &tInfo); err != nil {
				return d.handleErrorf("[%v]: %w", i, err)
			}
		}
	}

	return nil
}

func (d *Decoder) decodeMap(v reflect.Value, typeOpts TypeOptions) error {
	length := d.ReadLen()

	switch typeOpts.LenSizeInBytes {
	case 0:
	case Len2Bytes:
		if length > 0xFFFF {
			return d.handleErrorf("map size exceeds 2 bytes: %v", length)
		}
	case Len4Bytes:
		if length > 0xFFFFFFFF {
			return d.handleErrorf("map size exceeds 4 bytes: %v", length)
		}
	default:
		return d.handleErrorf("invalid map size type: %v", typeOpts.LenSizeInBytes)
	}

	keyType := v.Type().Key()
	valueType := v.Type().Elem()

	keyTypeInfo, err := d.getEncodedTypeInfo(keyType)
	if err != nil {
		return d.handleErrorf("key: %w", err)
	}

	valueTypeInfo, err := d.getEncodedTypeInfo(valueType)
	if err != nil {
		return d.handleErrorf("value: %w", err)
	}

	for i := 0; i < length; i++ {
		key := reflect.New(keyType).Elem()
		value := reflect.New(valueType).Elem()

		if err := d.decodeValue(key, typeOpts.MapKey, &keyTypeInfo); err != nil {
			return d.handleErrorf("key: %w", err)
		}

		if err := d.decodeValue(value, typeOpts.MapValue, &valueTypeInfo); err != nil {
			return d.handleErrorf("value: %w", err)
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
			fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		} else if fieldOpts.ExportAnonymousField {
			return d.handleErrorf("%v: field %v is already exported, but is marked for export", t.Name(), fieldType.Name)
		}

		fieldKind := fieldVal.Kind()

		if fieldKind == reflect.Ptr || fieldKind == reflect.Interface || fieldKind == reflect.Map || fieldKind == reflect.Slice {
			if fieldOpts.Optional {
				hasValue := d.ReadOptionalFlag()
				if d.err != nil {
					return d.handleErrorf("%v: %w", fieldType.Name, d.err)
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
			return d.handleErrorf("%v: %w", fieldType.Name, err)
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
	variantIdx := d.ReadEnumIdx()
	if d.err != nil {
		return d.err
	}

	if variantIdx >= len(variants) {
		return d.handleErrorf("invalid variant index %v for enum %v - enum has only %v variants", variantIdx, v.Type(), len(variants))
	}

	variantT := variants[variantIdx]
	if variantT == noneT {
		return nil
	}

	variant := reflect.New(variantT).Elem()

	if err := d.decodeValue(variant, nil, nil); err != nil {
		return d.handleErrorf("%v: %w", variants[variantIdx], err)
	}

	v.Set(variant)

	return nil
}

func (d *Decoder) decodeStructEnum(v reflect.Value) error {
	variantIdx := d.ReadEnumIdx()

	t := v.Type()

	if t.NumField() <= variantIdx {
		return d.handleErrorf("invalid variant index %v for enum %v - enum has only %v variants", variantIdx, t, t.NumField())
	}

	return d.decodeValue(v.Field(variantIdx), nil, nil)
}

func (d *Decoder) decodeAsByteArray(dec func() error) error {
	// This value was written as variable array of bytes.
	// Bytes of array are same as of value but they also have length prepended to them. So in theory we could just
	// skip length and continue reading. But that may result in confusing decoding errors in case of corrupted data.
	// So more reliable way is to separate those bytes and decode from them.

	b := make([]byte, d.ReadLen())
	d.Read(b)

	if d.err != nil {
		return d.handleErrorf("bytearr: %w", d.err)
	}

	origStream := d.r
	defer func() { d.r = origStream }() // for case of panic/error

	buff := bytes.NewBuffer(b)
	d.r = buff

	if err := dec(); err != nil {
		return err
	}
	if d.err != nil {
		return d.err
	}

	if avail := buff.Len(); avail > 0 {
		return d.handleErrorf("bytearr: excess bytes: %v", avail)
	}

	return nil
}

func (d *Decoder) handleErrorf(format string, args ...interface{}) error {
	d.err = fmt.Errorf(format, args...)
	return d.err
}

var (
	decodableT                 = reflect.TypeOf((*Decodable)(nil)).Elem()
	readableT                  = reflect.TypeOf((*Readable)(nil)).Elem()
	initializeableT            = reflect.TypeOf((*Initializeable)(nil)).Elem()
	decoderGlobalTypeInfoCache = newSharedTypeInfoCache()
)
