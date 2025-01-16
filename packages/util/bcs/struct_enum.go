package bcs

import "reflect"

type Enum interface {
	IsBcsEnum()
}

var enumT = reflect.TypeOf((*Enum)(nil)).Elem()
