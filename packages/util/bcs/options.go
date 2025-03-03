package bcs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type TypeOptions struct {
	// IncludeUnexported bool

	// TODO: Is this needed? It is present in rwutil as Size16/Size32, but it is more of validation.
	LenSizeInBytes LenBytesCount

	// TODO: Is this really useful? The engineer can just change type of int to indicate its size.
	UnderlyingType reflect.Kind

	IsCompactInt         bool
	InterfaceIsNotEnum   bool
	ExportAnonymousField bool
	NilIfEmpty           bool

	ArrayElement *ArrayElemOptions
	MapKey       *TypeOptions
	MapValue     *TypeOptions
}

func (o *TypeOptions) Validate() error {
	if err := o.LenSizeInBytes.Validate(); err != nil {
		return fmt.Errorf("array len size: %w", err)
	}
	if o.ArrayElement != nil {
		if err := o.ArrayElement.Validate(); err != nil {
			return fmt.Errorf("array element: %w", err)
		}
	}
	if o.MapKey != nil {
		if err := o.MapKey.Validate(); err != nil {
			return fmt.Errorf("map key: %w", err)
		}
	}
	if o.MapValue != nil {
		if err := o.MapValue.Validate(); err != nil {
			return fmt.Errorf("map value: %w", err)
		}
	}

	return nil
}

func (o *TypeOptions) Update(other TypeOptions) {
	if other.LenSizeInBytes != 0 {
		o.LenSizeInBytes = other.LenSizeInBytes
	}
	if other.UnderlyingType != reflect.Invalid {
		o.UnderlyingType = other.UnderlyingType
	}
	if other.IsCompactInt {
		o.IsCompactInt = true
	}
	if other.InterfaceIsNotEnum {
		o.InterfaceIsNotEnum = true
	}
	if other.NilIfEmpty {
		o.NilIfEmpty = true
	}
	if other.ExportAnonymousField {
		o.ExportAnonymousField = true
	}
	if other.ArrayElement != nil {
		if o.ArrayElement == nil {
			o.ArrayElement = other.ArrayElement
		} else {
			o.ArrayElement.Update(*other.ArrayElement)
		}
	}
	if other.MapKey != nil {
		if o.MapKey == nil {
			o.MapKey = other.MapKey
		} else {
			o.MapKey.Update(*other.MapKey)
		}
	}
	if other.MapValue != nil {
		if o.MapValue == nil {
			o.MapValue = other.MapValue
		} else {
			o.MapValue.Update(*other.MapValue)
		}
	}
}

type ArrayElemOptions struct {
	TypeOptions
	AsByteArray bool
}

func (o *ArrayElemOptions) Update(other ArrayElemOptions) {
	o.TypeOptions.Update(other.TypeOptions)
	if other.AsByteArray {
		o.AsByteArray = true
	}
}

type FieldOptions struct {
	TypeOptions
	Skip     bool
	Optional bool
	// OmitEmpty bool
	// ByteOrder    binary.ByteOrder
	AsByteArray bool
}

func (o *FieldOptions) Validate() error {
	if err := o.TypeOptions.Validate(); err != nil {
		return err
	}

	return nil
}

func FieldOptionsFromStruct(structType reflect.Type, tagName string) (_ []FieldOptions, hasTag []bool, err error) {
	fieldOpts := make([]FieldOptions, structType.NumField())
	hasTag = make([]bool, structType.NumField())

	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)

		fieldOpts[i], hasTag[i], err = FieldOptionsFromField(fieldType, tagName)
		if err != nil {
			return nil, nil, fmt.Errorf("field %v: %w", fieldType.Name, err)
		}
	}

	return fieldOpts, hasTag, nil
}

func FieldOptionsFromField(fieldType reflect.StructField, tagName string) (FieldOptions, bool, error) {
	a, hasTag := fieldType.Tag.Lookup(tagName)

	fieldOpts, err := FieldOptionsFromTag(a)
	if err != nil {
		return FieldOptions{}, false, fmt.Errorf("parsing annotation: %w", err)
	}

	switch fieldType.Type.Kind() {
	case reflect.Slice, reflect.Array:
		a, hasElemTag := fieldType.Tag.Lookup(tagName + "_elem")
		elemOpts, err := FieldOptionsFromTag(a)
		if err != nil {
			return FieldOptions{}, false, fmt.Errorf("parsing elem annotation: %w", err)
		}

		fieldOpts.ArrayElement = &ArrayElemOptions{
			TypeOptions: elemOpts.TypeOptions,
			AsByteArray: elemOpts.AsByteArray,
		}

		hasTag = hasTag || hasElemTag
	case reflect.Map:
		a, hasKeyTag := fieldType.Tag.Lookup(tagName + "_key")
		keyOpts, err := FieldOptionsFromTag(a)
		if err != nil {
			return FieldOptions{}, false, fmt.Errorf("parsing key annotation: %w", err)
		}

		fieldOpts.MapKey = &keyOpts.TypeOptions

		a, hasValueTag := fieldType.Tag.Lookup(tagName + "_value")
		valueOpts, err := FieldOptionsFromTag(a)
		if err != nil {
			return FieldOptions{}, false, fmt.Errorf("parsing value annotation: %w", err)
		}

		fieldOpts.MapValue = &valueOpts.TypeOptions

		hasTag = hasTag || hasKeyTag || hasValueTag
	}

	return fieldOpts, hasTag, nil
}

func FieldOptionsFromTag(a string) (_ FieldOptions, _ error) {
	if a == "" {
		return FieldOptions{}, nil
	}
	if a == "-" {
		return FieldOptions{Skip: true}, nil
	}

	opts := FieldOptions{}

	parts := strings.Split(a, ",")

	for _, part := range parts {
		subparts := strings.Split(part, "=")

		if len(subparts) > 2 {
			return FieldOptions{}, fmt.Errorf("invalid field tag: %s", part)
		}

		key := subparts[0]
		val := ""
		if len(subparts) == 2 {
			val = subparts[1]
		}

		switch key {
		case "compact":
			opts.IsCompactInt = true
		case "type":
			var err error
			opts.UnderlyingType, err = UnderlayingTypeFromString(val)
			if err != nil {
				return FieldOptions{}, fmt.Errorf("invalid undelaying type tag: %s", val)
			}
		case "len_bytes":
			bytes, err := strconv.Atoi(val)
			if err != nil {
				return FieldOptions{}, fmt.Errorf("invalid len_bytes tag: %s", val)
			}

			opts.LenSizeInBytes = LenBytesCount(bytes) //nolint:gosec
		case "optional":
			opts.Optional = true
		case "nil_if_empty":
			opts.NilIfEmpty = true
		case "bytearr":
			opts.AsByteArray = true
		case "not_enum":
			opts.InterfaceIsNotEnum = true
		case "export":
			opts.ExportAnonymousField = true
		case "":
			return FieldOptions{}, fmt.Errorf("empty field tag entry")
		default:
			return FieldOptions{}, fmt.Errorf("unknown field tag: %s", key)
		}
	}

	return opts, nil
}

type LenBytesCount uint8

const (
	Len2Bytes LenBytesCount = 2
	Len4Bytes LenBytesCount = 4
)

func (s LenBytesCount) Validate() error {
	switch s {
	case Len2Bytes, Len4Bytes:
		return nil
	default:
		return fmt.Errorf("invalid collection len size: %v", s)
	}
}

func UnderlayingTypeFromString(s string) (reflect.Kind, error) {
	switch s {
	case "i8", "int8":
		return reflect.Int8, nil
	case "i16", "int16":
		return reflect.Int16, nil
	case "i32", "int32":
		return reflect.Int32, nil
	case "i64", "int64":
		return reflect.Int64, nil
	case "u8", "uint8":
		return reflect.Uint8, nil
	case "u16", "uint16":
		return reflect.Uint16, nil
	case "u32", "uint32":
		return reflect.Uint32, nil
	case "u64", "uint64":
		return reflect.Uint64, nil
	default:
		return 0, fmt.Errorf("invalid value size: %s", s)
	}
}

type BCSType interface {
	BCSOptions() TypeOptions
}

var bcsTypeT = reflect.TypeOf((*BCSType)(nil)).Elem()
