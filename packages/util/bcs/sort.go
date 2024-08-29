package bcs

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
)

func sortSlice(s []reflect.Value) {
	if len(s) < 2 {
		return
	}

	elemType := s[0].Type()

	switch elemType.Kind() {
	case reflect.Int:
		sortInts[int](s)
	case reflect.Int8:
		sortInts[int8](s)
	case reflect.Int16:
		sortInts[int16](s)
	case reflect.Int32:
		sortInts[int32](s)
	case reflect.Int64:
		sortInts[int64](s)
	case reflect.Uint:
		sortUints[uint](s)
	case reflect.Uint8:
		sortUints[uint8](s)
	case reflect.Uint16:
		sortUints[uint16](s)
	case reflect.Uint32:
		sortUints[uint32](s)
	case reflect.Uint64:
		sortUints[uint64](s)
	case reflect.String:
		sortStrings(s)
	default:
		panic(fmt.Errorf("unsupported slice elem type: %v", elemType))
	}
}

func sortInts[Int constraints.Integer](s []reflect.Value) {
	sort.Slice(s, func(i, j int) bool {
		// TODO: Maybe it is more optimal to create new slice of real type + orig idx, sort it, and then rearrange original slice?
		return Int(s[i].Int()) < Int(s[j].Int())
	})
}

func sortUints[Uint constraints.Unsigned](s []reflect.Value) {
	sort.Slice(s, func(i, j int) bool {
		return Uint(s[i].Uint()) < Uint(s[j].Uint())
	})
}

func sortStrings(s []reflect.Value) {
	vals := lo.Map(s, func(v reflect.Value, _ int) string {
		return v.String()
	})

	sort.Slice(s, func(i, j int) bool {
		return vals[i] < vals[j]
	})
}
