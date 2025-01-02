package bcs

import "reflect"

// These can be used to represent the absence of a value in inteface enums
// Example:
//
//	bcs.RegisterEnumType2[EnumType, None, SomeVariant]()
type (
	None struct{}
	Nil  = None
)

var noneT = reflect.TypeOf(None{})
