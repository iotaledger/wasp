package bcs

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
)

func sortMap(entries []*lo.Entry[reflect.Value, reflect.Value]) error {
	// NOTE: Map is sorted to ensure deterministic encoding.
	// It is sorted by the actual value, although by BCS spec it should be sorted by encoded bytes of key.
	// This should not be an issue, because sorting of ints and strings in Go already happen by its bytes.

	if len(entries) < 2 {
		return nil
	}

	keyType := entries[0].Key.Type()

	var cmp func(i, j int) bool

	//NOTE: We use unsigned types fo signed values to ensure that they are sorted by its byte representation.

	switch keyType.Kind() {
	case reflect.Int:
		cmp = cmpIntKeys[uint](entries)
	case reflect.Int8:
		cmp = cmpIntKeys[uint8](entries)
	case reflect.Int16:
		cmp = cmpIntKeys[uint16](entries)
	case reflect.Int32:
		cmp = cmpIntKeys[uint32](entries)
	case reflect.Int64:
		cmp = cmpIntKeys[uint64](entries)
	case reflect.Uint:
		cmp = cmpUintKeys[uint](entries)
	case reflect.Uint8:
		cmp = cmpUintKeys[uint8](entries)
	case reflect.Uint16:
		cmp = cmpUintKeys[uint16](entries)
	case reflect.Uint32:
		cmp = cmpUintKeys[uint32](entries)
	case reflect.Uint64:
		cmp = cmpUintKeys[uint64](entries)
	case reflect.String:
		cmp = cmpStringKeys(entries)
	default:
		return fmt.Errorf("unsupported map key type: %v", keyType)
	}

	sort.Slice(entries, cmp)

	return nil
}

// NOTE: Uint constraints.Unsigned is not a typo. See comment in sortMap.
func cmpIntKeys[Uint constraints.Unsigned](entries []*lo.Entry[reflect.Value, reflect.Value]) func(i, j int) bool {
	return func(i, j int) bool {
		return Uint(entries[i].Key.Int()) < Uint(entries[j].Key.Int())
	}
}

func cmpUintKeys[Uint constraints.Unsigned](entries []*lo.Entry[reflect.Value, reflect.Value]) func(i, j int) bool {
	return func(i, j int) bool {
		return Uint(entries[i].Key.Uint()) < Uint(entries[j].Key.Uint())
	}
}

func cmpStringKeys(entries []*lo.Entry[reflect.Value, reflect.Value]) func(i, j int) bool {
	return func(i, j int) bool {
		return entries[i].Key.String() < entries[j].Key.String()
	}
}
