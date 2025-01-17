package bcs

import "reflect"

type Enum interface {
	IsBcsEnum()
}

var structEnumT = reflect.TypeOf((*Enum)(nil)).Elem()
