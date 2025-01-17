package bcs

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"unsafe"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Encodable interface {
	MarshalBCS(e *Encoder) error
}

var encodableT = reflect.TypeOf((*Encodable)(nil)).Elem()

type Writable interface {
	Write(w io.Writer) error
}

var writableT = reflect.TypeOf((*Writable)(nil)).Elem()

type EncoderConfig struct {
	TagName                  string
	InterfaceIsEnumByDefault bool
	// IncludeUnexported bool
	// IncludeUntaggedUnexported bool
	// ExcludeUntagged           bool
	// CustomEncoders map[reflect.Type]CustomEncoder
}

func (c *EncoderConfig) InitializeDefaults() {
	if c.TagName == "" {
		c.TagName = "bcs"
	}
}

func NewBytesEncoder() *BytesEncoder {
	var buf bytes.Buffer
	return &BytesEncoder{Encoder: *NewEncoder(&buf), buf: &buf}
}

type BytesEncoder struct {
	Encoder
	buf *bytes.Buffer
}

func (e *BytesEncoder) Bytes() []byte {
	return e.buf.Bytes()
}

func NewEncoder(dest io.Writer) *Encoder {
	return NewEncoderWithOpts(dest, EncoderConfig{})
}

func NewEncoderWithOpts(dest io.Writer, cfg EncoderConfig) *Encoder {
	cfg.InitializeDefaults()

	return &Encoder{
		cfg:           cfg,
		w:             rwutil.NewWriter(dest),
		typeInfoCache: encoderGlobalTypeInfoCache.Get(),
	}
}

type Encoder struct {
	cfg           EncoderConfig
	w             *rwutil.Writer
	typeInfoCache localTypeInfoCache
}

var encoderGlobalTypeInfoCache = newSharedTypeInfoCache()

func (e *Encoder) Err() error {
	return e.w.Err
}

func (e *Encoder) MustEncode(val any) {
	e.Encode(val)
	if e.w.Err != nil {
		panic(e.w.Err)
	}
}

// If error occurs, it will be stored inside of encoder and can be checked using enc.Err().
// After error further calls to Encode() will just do nothing.
// So no need to check error every time.
// Example:
//
//	enc.Encode(&v1)
//	enc.Encode(&v2)
//	enc.Encode(&v3)
//
//	if err := enc.Err(); err != nil {
//	    return err
//	}
//
// If Encode() is called inside of MarshalBCS() method, you can even skip checking enc.Err(),
// because decoder itself will do it for you anyway.
// Example:
//
//	func (p *MyStruct) MarshalBCS(e *bcs.Encoder) error {
//	    e.Encode(&p.Field1)
//	    e.Encode(&p.Field2)
//	    return nil
//	}
func (e *Encoder) Encode(val any) {
	if e.w.Err != nil {
		return
	}

	if val == nil {
		_ = e.handleErrorf("cannot encode a nil value")
		return
	}

	defer e.typeInfoCache.Save()

	if err := e.encodeValue(reflect.ValueOf(val), nil, nil); err != nil {
		_ = fmt.Errorf("encoding %T: %w", val, err)
		return
	}
}

func (e *Encoder) EncodeOptional(val any) {
	if e.w.Err != nil {
		return
	}

	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map:
	default:
		_ = e.handleErrorf("optional value must be a pointer, interface or map, got %v", v.Type())
		return
	}

	if v.IsNil() {
		e.w.WriteByte(0)
		return
	}

	e.w.WriteByte(1)
	e.Encode(val)
}

func (e *Encoder) WriteOptionalFlag(hasValue bool) {
	if hasValue {
		e.w.WriteByte(1)
	} else {
		e.w.WriteByte(0)
	}
}

// Enum index is an index of variant in enum type.
func (e *Encoder) WriteEnumIdx(variantIdx int) {
	e.w.WriteSize32(variantIdx)
}

func (e *Encoder) WriteLen(length int) {
	e.w.WriteSize32(length)
}

// ULEB - unsigned little-endian base-128 - variable-length integer value.
func (e *Encoder) WriteCompactUint(v uint64) {
	e.w.WriteAmount64(v)
}

func (e *Encoder) WriteBool(v bool) {
	e.w.WriteBool(v)
}

//nolint:govet
func (e *Encoder) WriteByte(v byte) {
	e.w.WriteByte(v)
}

func (e *Encoder) WriteInt8(v int8) {
	e.w.WriteInt8(v)
}

func (e *Encoder) WriteInt16(v int16) {
	e.w.WriteInt16(v)
}

func (e *Encoder) WriteInt32(v int32) {
	e.w.WriteInt32(v)
}

func (e *Encoder) WriteInt64(v int64) {
	e.w.WriteInt64(v)
}

func (e *Encoder) WriteInt(v int) {
	e.w.WriteInt64(int64(v))
}

func (e *Encoder) WriteUint8(v uint8) {
	e.w.WriteUint8(v)
}

func (e *Encoder) WriteUint16(v uint16) {
	e.w.WriteUint16(v)
}

func (e *Encoder) WriteUint32(v uint32) {
	e.w.WriteUint32(v)
}

func (e *Encoder) WriteUint64(v uint64) {
	e.w.WriteUint64(v)
}

func (e *Encoder) WriteUint(v uint) {
	e.w.WriteUint64(uint64(v))
}

func (e *Encoder) WriteString(v string) {
	e.w.WriteString(v)
}

// For support of io.Writer interface
func (e *Encoder) Write(b []byte) (n int, err error) {
	e.w.WriteFromFunc(func(w io.Writer) (int, error) {
		n, err = w.Write(b)
		return n, err
	})

	return n, e.w.Err
}

//nolint:gocyclo,funlen
func (e *Encoder) encodeValue(v reflect.Value, typeOptionsFromTag *TypeOptions, tInfo *typeInfo) error {
	if tInfo == nil {
		// Hint about type customization could have been provided by caller when encoding collections.
		// This is done to avoid parsing type for each element of collection.
		// This is an optimization for encoding of large amount of simple elements.

		t, err := e.getEncodedTypeInfo(v.Type())
		if err != nil {
			return err
		}

		tInfo = &t
	}

	v, err := e.getEncodedValue(v, tInfo.RefLevelsCount)
	if err != nil {
		return fmt.Errorf("%v: %w", v.Type(), err)
	}

	if tInfo.CustomEncoder != nil {
		if err := tInfo.CustomEncoder(e, v); err != nil { //nolint:govet
			if e.w.Err == nil {
				e.w.Err = err
			}
			return fmt.Errorf("%v: custom encoder: %w", v.Type(), err)
		}
		if e.w.Err != nil {
			return fmt.Errorf("%v: custom encoder: %w", v.Type(), e.w.Err)
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

	switch v.Kind() {
	case reflect.Bool:
		e.w.WriteBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typeOptions.IsCompactInt {
			e.WriteCompactUint(uint64(v.Int())) //nolint:gosec
		} else {
			err = e.encodeInt(v, typeOptions.UnderlyingType)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if typeOptions.IsCompactInt {
			e.WriteCompactUint(v.Uint())
		} else {
			err = e.encodeUint(v, typeOptions.UnderlyingType)
		}
	case reflect.String:
		e.w.WriteString(v.String())
	case reflect.Slice:
		if typeOptions.ArrayElement == nil {
			typeOptions.ArrayElement = &ArrayElemOptions{}
		}
		err = e.encodeSlice(v, typeOptions)
	case reflect.Array:
		if typeOptions.ArrayElement == nil {
			typeOptions.ArrayElement = &ArrayElemOptions{}
		}
		err = e.encodeArray(v, typeOptions)
	case reflect.Map:
		if typeOptions.MapKey == nil {
			typeOptions.MapKey = &TypeOptions{}
		}
		if typeOptions.MapValue == nil {
			typeOptions.MapValue = &TypeOptions{}
		}
		err = e.encodeMap(v, typeOptions)
	case reflect.Struct:
		if tInfo.IsStructEnum {
			err = e.encodeStructEnum(v)
		} else {
			err = e.encodeStruct(v, tInfo)
		}
	case reflect.Interface:
		err = e.encodeInterface(v, !typeOptions.InterfaceIsNotEnum)
	default:
		return e.handleErrorf("%v: cannot encode unknown type", v.Type())
	}

	if err != nil {
		return fmt.Errorf("%v: %w", v.Type(), err)
	}
	if e.w.Err != nil {
		return fmt.Errorf("%v: %w", v.Type(), e.w.Err)
	}

	return nil
}

// This structure is used to store result of parsing type to reuse it for each of element of collection.
type typeInfo struct {
	RefLevelsCount int
	typeCustomization
	FieldOptions []FieldOptions
	FieldHasTag  []bool
}

// Finds actual type we want to encode from the current type of value.
// Possible cases:
// 1. Type has multiple layers of pointers. We need to remove them all or until first type with custom encoder.
// 2. Type is not a pointer but its pointer type has custom encoder. In this case we need to use pointer to value instead of value itself.
func (e *Encoder) getEncodedTypeInfo(t reflect.Type) (typeInfo, error) {
	initialT := t

	if cached, isCached := e.typeInfoCache.Get(initialT); isCached {
		return cached, nil
	}

	refLevelsCount := 0

	if t.Kind() != reflect.Ptr {
		// Type is not a pointer but value. But there could be custom encoder for
		// its pointer type, so need to check it. And if there is, we need to use
		// pointer to value instead of value itself.
		// If value is not addressable, we need to copy it to make it addressable.

		customEncoder := e.getCustomEncoder(reflect.PointerTo(t))
		if customEncoder != nil {
			res := typeInfo{RefLevelsCount: -1, typeCustomization: typeCustomization{CustomEncoder: customEncoder}}
			e.typeInfoCache.Add(initialT, res)

			return res, nil
		}
	} else {
		// Value is a pointer

		// Removing all redundant pointers
		for t.Kind() == reflect.Ptr {
			// Before removing pointer, we need to check if maybe current type is already the type we should encode.
			customEncoder := e.getCustomEncoder(t)
			if customEncoder != nil {
				res := typeInfo{RefLevelsCount: refLevelsCount, typeCustomization: typeCustomization{CustomEncoder: customEncoder}}
				e.typeInfoCache.Add(initialT, res)

				return res, nil
			}

			refLevelsCount++
			t = t.Elem()
		}
	}

	customization := e.checkTypeCustomizations(t)

	res := typeInfo{RefLevelsCount: refLevelsCount, typeCustomization: customization}

	if t.Kind() == reflect.Struct {
		// Value type is struct - parsing tags of its fields
		var err error
		res.FieldOptions, res.FieldHasTag, err = FieldOptionsFromStruct(t, e.cfg.TagName)
		if err != nil {
			return typeInfo{}, fmt.Errorf("parsing struct fields options: %v: %w", t, err)
		}
	}

	e.typeInfoCache.Add(initialT, res)

	return res, nil
}

func (e *Encoder) getEncodedValue(v reflect.Value, refsCount int) (valToEncode reflect.Value, _ error) {
	if refsCount == -1 {
		// Custom encoder for pointer type is found, so we need to encode pointer to value instead of value itself.
		if v.CanAddr() {
			return v.Addr(), nil
		}

		// Value is not addressable - copying it to make it addressable
		copied := reflect.New(v.Type())
		copied.Elem().Set(v)

		return copied, nil
	}

	// Removing all found redundant pointers
	for i := 0; i < refsCount; i++ {
		if v.IsNil() {
			return v, e.handleErrorf("attempt to encode non-optinal nil value of type %v", v.Type())
		}

		v = v.Elem()
	}

	return v, nil
}

type typeCustomization struct {
	CustomEncoder  CustomEncoder
	CustomDecoder  CustomDecoder
	Init           InitFunc
	IsStructEnum   bool
	HasTypeOptions bool
}

func (c *typeCustomization) HasCustomizations() bool {
	return c.CustomEncoder != nil || c.CustomDecoder != nil || c.Init != nil || c.IsStructEnum || c.HasTypeOptions
}

func (e *Encoder) checkTypeCustomizations(t reflect.Type) typeCustomization {
	// Detecting enum variant index might return error, so we
	// should first check for existence of custom encoder.
	if customEncoder := e.getCustomEncoder(t); customEncoder != nil {
		return typeCustomization{CustomEncoder: customEncoder}
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

func (e *Encoder) getCustomEncoder(t reflect.Type) CustomEncoder {
	// Check if this type has custom encoder func
	if customEncoder, ok := CustomEncoders[t]; ok {
		return customEncoder
	}

	// Check if this type implements custom encoding interface.
	// Although we could allow encoding of interfaces, which implement Encodable, still
	// we exclude them here to ensure symetric behavior with decoding.
	if t.Kind() == reflect.Interface {
		return nil
	}

	if t.Implements(encodableT) {
		return func(e *Encoder, v reflect.Value) error {
			return v.Interface().(Encodable).MarshalBCS(e)
		}
	}

	if t.Implements(writableT) {
		return func(e *Encoder, v reflect.Value) error {
			return v.Interface().(Writable).Write(e)
		}
	}

	return nil
}

func (e *Encoder) encodeInt(v reflect.Value, encodedType reflect.Kind) error {
	k := v.Kind()

	if encodedType != reflect.Invalid && encodedType != k {
		return convertEncodeNumber(e, v.Int(), encodedType)
	}

	switch k {
	case reflect.Int8:
		e.w.WriteInt8(int8(v.Int())) //nolint:gosec
	case reflect.Int16:
		e.w.WriteInt16(int16(v.Int())) //nolint:gosec
	case reflect.Int32:
		e.w.WriteInt32(int32(v.Int())) //nolint:gosec
	case reflect.Int64, reflect.Int:
		e.w.WriteInt64(v.Int())
	default:
		panic(fmt.Sprintf("unexpected int kind: %v", k))
	}

	return nil
}

func (e *Encoder) encodeUint(v reflect.Value, encodedType reflect.Kind) error {
	k := v.Kind()

	if encodedType != reflect.Invalid && encodedType != k {
		return convertEncodeNumber(e, v.Uint(), encodedType)
	}

	switch k {
	case reflect.Uint8:
		e.w.WriteUint8(uint8(v.Uint())) //nolint:gosec
	case reflect.Uint16:
		e.w.WriteUint16(uint16(v.Uint())) //nolint:gosec
	case reflect.Uint32:
		e.w.WriteUint32(uint32(v.Uint())) //nolint:gosec
	case reflect.Uint64, reflect.Uint:
		e.w.WriteUint64(v.Uint())
	default:
		panic(fmt.Sprintf("unexpected uint kind: %v", k))
	}

	return nil
}

func convertEncodeNumber[Value constraints.Numeric](e *Encoder, v Value, encodedType reflect.Kind) error {
	switch encodedType {
	case reflect.Int8:
		return convertEncodeNumber2(e, v, e.w.WriteInt8)
	case reflect.Int16:
		return convertEncodeNumber2(e, v, e.w.WriteInt16)
	case reflect.Int32:
		return convertEncodeNumber2(e, v, e.w.WriteInt32)
	case reflect.Int64, reflect.Int:
		return convertEncodeNumber2(e, v, e.w.WriteInt64)
	case reflect.Uint8:
		return convertEncodeNumber2(e, v, e.w.WriteUint8)
	case reflect.Uint16:
		return convertEncodeNumber2(e, v, e.w.WriteUint16)
	case reflect.Uint32:
		return convertEncodeNumber2(e, v, e.w.WriteUint32)
	case reflect.Uint64, reflect.Uint:
		return convertEncodeNumber2(e, v, e.w.WriteUint64)
	default:
		return e.handleErrorf("invalid underlaying type %v for type %T", encodedType, lo.Empty[Value]())
	}
}

// The name has suffix 2 because it is a helper function for convertEncodeNumber to unwrap type To.
func convertEncodeNumber2[To, From constraints.Numeric, Ret any](e *Encoder, v From, write func(To) Ret) error {
	converted := To(v)

	if From(converted) != v {
		return e.handleErrorf("value %v is out of range of type %T", v, To(0))
	}

	write(converted)

	return nil
}

func (e *Encoder) encodeSlice(v reflect.Value, typeOpts TypeOptions) error {
	switch typeOpts.LenSizeInBytes {
	case Len2Bytes:
		e.w.WriteSize16(v.Len())
	case Len4Bytes, 0:
		e.w.WriteSize32(v.Len())
	default:
		return e.handleErrorf("invalid collection size type: %v", typeOpts.LenSizeInBytes)
	}

	return e.encodeArray(v, typeOpts)
}

func (e *Encoder) encodeArray(v reflect.Value, typeOpts TypeOptions) error {
	elemType := v.Type().Elem()

	tInfo, err := e.getEncodedTypeInfo(elemType)
	if err != nil {
		return fmt.Errorf("element: %w", err)
	}

	if !tInfo.HasCustomizations() {
		// The type does not have any customizations. So we can use  some optimizations for encoding of basic types
		if elemType.Kind() == reflect.Uint8 && (v.Kind() == reflect.Slice || v.CanAddr()) && !typeOpts.ArrayElement.AsByteArray {
			// Optimization for []byte and [N]byte.
			e.w.WriteN(v.Bytes())
			return nil
		}

		// There could be other optimizations for encoding of basic types. But I removed them for now for simplicity.
	}

	if typeOpts.ArrayElement.AsByteArray {
		for i := 0; i < v.Len(); i++ {
			err := e.encodeAsByteArray(func() error {
				return e.encodeValue(v.Index(i), &typeOpts.ArrayElement.TypeOptions, &tInfo)
			})
			if err != nil {
				return fmt.Errorf("[%v]: %v: %w", i, elemType, err)
			}
		}
	} else {
		for i := 0; i < v.Len(); i++ {
			if err := e.encodeValue(v.Index(i), &typeOpts.ArrayElement.TypeOptions, &tInfo); err != nil {
				return fmt.Errorf("[%v]: %v: %w", i, elemType, err)
			}
		}
	}

	return nil
}

func (e *Encoder) encodeMap(v reflect.Value, typeOpts TypeOptions) error {
	if v.IsNil() {
		return e.handleErrorf("attempt to encode non-optional nil-map")
	}

	switch typeOpts.LenSizeInBytes {
	case Len2Bytes:
		e.w.WriteSize16(v.Len())
	case Len4Bytes, 0:
		e.w.WriteSize32(v.Len())
	default:
		return e.handleErrorf("invalid collection size type: %v", typeOpts.LenSizeInBytes)
	}

	t := v.Type()
	keyTypeInfo, err := e.getEncodedTypeInfo(t.Key())
	if err != nil {
		return fmt.Errorf("key: %w", err)
	}

	valTypeInfo, err := e.getEncodedTypeInfo(t.Elem())
	if err != nil {
		return fmt.Errorf("value: %w", err)
	}

	entries := make([]*lo.Tuple2[[]byte, reflect.Value], 0, v.Len())

	for elem := v.MapRange(); elem.Next(); {
		// Encoding keys to be able to sort map entries by key's bytes
		encodedKey, err := e.getBytes(func() error {
			return e.encodeValue(elem.Key(), typeOpts.MapKey, &keyTypeInfo)
		})
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}

		entry := lo.T2[[]byte, reflect.Value](encodedKey, elem.Value())
		entries = append(entries, &entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare(entries[i].A, entries[j].A) < 0
	})

	for i := range entries {
		e.w.WriteN(entries[i].A)

		if err := e.encodeValue(entries[i].B, typeOpts.MapValue, &valTypeInfo); err != nil {
			return fmt.Errorf("value: %w", err)
		}
	}

	return nil
}

func (e *Encoder) encodeStruct(v reflect.Value, tInfo *typeInfo) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldOpts, hasTag := tInfo.FieldOptions[i], tInfo.FieldHasTag[i]
		if fieldOpts.Skip {
			continue
		}

		fieldType := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldType.IsExported() {
			if !fieldOpts.ExportAnonymousField {
				if hasTag {
					return e.handleErrorf("%v: unexported field %v has BCS tag, but is not marked for export", t.Name(), fieldType.Name)
				}

				// Unexported fields are skipped by default if not explicitly marked as exported
				continue
			}

			if !fieldVal.CanAddr() {
				// Field is not addressable yet - making it addressable
				vCopy := reflect.New(t).Elem()
				vCopy.Set(v)
				v = vCopy
				fieldVal = v.Field(i)
			}

			// Accesing unexported field
			// Trick to access unexported fields: https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields/43918797#43918797
			fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		} else if fieldOpts.ExportAnonymousField {
			return e.handleErrorf("%v: field %v is already exported, but is marked for export", t.Name(), fieldType.Name)
		}

		fieldKind := fieldVal.Kind()

		if fieldKind == reflect.Ptr || fieldKind == reflect.Interface || fieldKind == reflect.Map {
			// The field is nullable

			isNil := fieldVal.IsNil()

			if isNil && !fieldOpts.Optional && fieldKind != reflect.Interface {
				return e.handleErrorf("%v: non-optional nil value", fieldType.Name)
			}

			if fieldOpts.Optional {
				e.w.WriteByte(lo.Ternary[byte](isNil, 0, 1))

				if isNil {
					continue
				}
			}
		}

		var err error

		if fieldOpts.AsByteArray {
			err = e.encodeAsByteArray(func() error {
				return e.encodeValue(fieldVal, &fieldOpts.TypeOptions, nil)
			})
		} else {
			err = e.encodeValue(fieldVal, &fieldOpts.TypeOptions, nil)
		}

		if err != nil {
			return e.handleErrorf("%v: %w", fieldType.Name, err)
		}
	}

	return nil
}

func (e *Encoder) encodeStructEnum(v reflect.Value) error {
	enumVariantIdx, err := e.getStructEnumVariantIdx(v)
	if err != nil {
		return err
	}

	if err := e.encodeEnum(v.Field(enumVariantIdx), enumVariantIdx); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) getStructEnumVariantIdx(v reflect.Value) (enumVariantIdx EnumVariantID, _ error) {
	enumVariantIdx = -1

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		k := field.Kind()
		switch k {
		case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
			if field.IsNil() {
				continue
			}

			if enumVariantIdx != -1 {
				prevSetField := v.Type().Field(enumVariantIdx)
				currentField := v.Type().Field(i)
				return -1, e.handleErrorf("multiple options are set in enum struct %v: %v and %v", v.Type(), prevSetField.Name, currentField.Name)
			}

			enumVariantIdx = i
			// We do not break here to check if there are multiple options set
		default:
			fieldType := v.Type().Field(i)
			return -1, e.handleErrorf("field %v of enum %v is of non-nullable type %v", fieldType.Name, v.Type(), fieldType.Type)
		}
	}

	if enumVariantIdx == -1 {
		return -1, e.handleErrorf("no options are set in enum struct %v", v.Type())
	}

	return enumVariantIdx, nil
}

func (e *Encoder) encodeInterface(v reflect.Value, couldBeEnum bool) error {
	if !couldBeEnum {
		if v.IsNil() {
			return e.handleErrorf("cannot encode nil interface, which is not enum and not optional")
		}

		return e.encodeValue(v.Elem(), nil, nil)
	}

	t := v.Type()

	enumVariants, registered := EnumTypes[t]
	if !registered {
		if e.cfg.InterfaceIsEnumByDefault {
			return e.handleErrorf("interface %v is not registered as enum type", t)
		}

		if v.IsNil() {
			return e.handleErrorf("cannot encode nil interface, which is not enum and not optional")
		}

		return e.encodeValue(v.Elem(), nil, nil)
	}

	enumVariantIdx, err := e.getInterfaceEnumVariantIdx(v, enumVariants)
	if err != nil {
		return err
	}

	if err := e.encodeEnum(v.Elem(), enumVariantIdx); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) getInterfaceEnumVariantIdx(v reflect.Value, enumVariants map[int]reflect.Type) (enumVariantIdx EnumVariantID, _ error) {
	isNil := v.IsNil()

	var valT reflect.Type
	if isNil {
		valT = noneT
	} else {
		valT = v.Elem().Type()
	}

	enumVariantIdx = -1

	for id, variant := range enumVariants {
		if valT == variant {
			enumVariantIdx = id
		}
	}

	if enumVariantIdx == -1 {
		if isNil {
			return -1, e.handleErrorf("bcs.None is not registered as part of enum type %v - cannot encode nil interface enum value", v.Type())
		}
		return -1, e.handleErrorf("variant %v is not registered as part of enum type %v", valT, v.Type())
	}

	return enumVariantIdx, nil
}

func (e *Encoder) encodeEnum(v reflect.Value, variantIdx int) error {
	e.w.WriteSize32(variantIdx)

	if !v.IsValid() {
		return nil
	}

	if err := e.encodeValue(v, nil, nil); err != nil {
		return fmt.Errorf("%v: %w", v.Type(), err)
	}

	return nil
}

// Captures bytes written by enc() and prepends them with their count.
func (e *Encoder) encodeAsByteArray(enc func() error) error {
	encodedVal, err := e.getBytes(enc)
	if err != nil {
		return err
	}

	e.w.WriteSize32(len(encodedVal))
	e.w.WriteN(encodedVal)

	if e.w.Err != nil {
		return fmt.Errorf("bytearr: %w", e.w.Err)
	}

	return nil
}

func (e *Encoder) getBytes(enc func() error) ([]byte, error) {
	origStream := e.w
	defer func() { e.w = origStream }() // for case of panic/error

	e.w = rwutil.NewBytesWriter()
	if err := enc(); err != nil {
		return nil, err
	}

	encodedVal := e.w.Bytes()

	return encodedVal, nil
}

func (e *Encoder) handleErrorf(format string, args ...interface{}) error {
	e.w.Err = fmt.Errorf(format, args...)
	return e.w.Err
}

// Pointer is forced here for two reasons:
//   - This allows to avoid copying of value in cases when there is custom encoder exists with pointer receiver
//   - This allow to detect actual type of interface value. Because otherwise the implementation has no way to detect interface.
//
// But because of that encoding a value, which is stored in variable of type "any" would be very inconvenient.
// So to make it more user-friendly, this function treats "*any" as "any".
func MarshalStream[V any](v *V, dest io.Writer) error {
	e := NewEncoder(dest)

	switch v := interface{}(v).(type) {
	case *interface{}:
		// Exception for pointer to "any" just for convenience.
		e.Encode(*v)
	default:
		e.Encode(v)
	}

	return e.w.Err
}

func MustMarshalStream[V any](v *V, dest io.Writer) {
	if err := MarshalStream(v, dest); err != nil {
		panic(fmt.Errorf("failed to marshal object of type %T into BCS: %w", v, err))
	}
}

func Marshal[V any](v *V) ([]byte, error) {
	var buf bytes.Buffer

	if err := MarshalStream(v, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func MustMarshal[V any](v *V) []byte {
	b, err := Marshal(v)
	if err != nil {
		panic(fmt.Errorf("failed to marshal object of type %T into BCS: %w", v, err))
	}

	return b
}

type CustomEncoder func(e *Encoder, v reflect.Value) error

var CustomEncoders = make(map[reflect.Type]CustomEncoder)

func MakeCustomEncoder[V any](f func(e *Encoder, v V) error) func(e *Encoder, v reflect.Value) error {
	return func(e *Encoder, v reflect.Value) error {
		return f(e, v.Interface().(V))
	}
}

func AddCustomEncoder[V any](f func(e *Encoder, v V) error) {
	CustomEncoders[reflect.TypeOf((*V)(nil)).Elem()] = MakeCustomEncoder(f)
}

func RemoveCustomEncoder[V any]() {
	delete(CustomEncoders, reflect.TypeOf((*V)(nil)).Elem())
}
