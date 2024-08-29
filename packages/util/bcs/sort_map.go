package bcs

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
)

func sortMap(entries []*lo.Entry[reflect.Value, reflect.Value]) {
	if len(entries) < 2 {
		return
	}

	keyType := entries[0].Key.Type()

	var cmp func(i, j int) bool

	switch keyType.Kind() {
	case reflect.Int:
		cmp = cmpIntKeys[int](entries)
	case reflect.Int8:
		cmp = cmpIntKeys[int8](entries)
	case reflect.Int16:
		cmp = cmpIntKeys[int16](entries)
	case reflect.Int32:
		cmp = cmpIntKeys[int32](entries)
	case reflect.Int64:
		cmp = cmpIntKeys[int64](entries)
	case reflect.Uint:
		cmp = cmpIntKeys[uint](entries)
	case reflect.Uint8:
		cmp = cmpIntKeys[uint8](entries)
	case reflect.Uint16:
		cmp = cmpIntKeys[uint16](entries)
	case reflect.Uint32:
		cmp = cmpIntKeys[uint32](entries)
	case reflect.Uint64:
		cmp = cmpIntKeys[uint64](entries)
	case reflect.String:
		cmp = cmpStringKeys(entries)
	default:
		panic(fmt.Errorf("unsupported slice elem type: %v", keyType))
	}

	sort.Slice(entries, cmp)
}

func cmpIntKeys[Int constraints.Integer](entries []*lo.Entry[reflect.Value, reflect.Value]) func(i, j int) bool {
	return func(i, j int) bool {
		// TODO: Maybe it is more optimal to create new slice of real type + orig idx, sort it, and then rearrange original slice?
		return Int(entries[i].Key.Int()) < Int(entries[j].Key.Int())
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
