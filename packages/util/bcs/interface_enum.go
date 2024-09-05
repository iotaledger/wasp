package bcs

import (
	"fmt"
	"reflect"
)

type EnumVariantID = int

var EnumTypes = make(map[reflect.Type]map[EnumVariantID]reflect.Type)

// NOTE: Registration is not thread-safe for now as it is assumed that all types are registered upon initialization.
func RegisterEnumTypeVariant[EnumType any](id EnumVariantID, newVariant any) struct{} {
	enumT := reflect.TypeOf((*EnumType)(nil)).Elem()
	newVariantT := reflect.TypeOf(newVariant)

	if id < 0 {
		panic(fmt.Errorf("RegisterEnumType: attempt to register variant type %v of enum %v with negative id %v", newVariantT, enumT, id))
	}

	if newVariantT.Kind() == reflect.Interface {
		panic(fmt.Errorf("RegisterEnumType: variant type %v of enum %v is an interface", newVariantT, enumT))
	}

	if !newVariantT.Implements(enumT) {
		panic(fmt.Errorf("RegisterEnumType: variant type %v does not implement enum %v", newVariantT, enumT))
	}

	registeredVariants, enumRegistered := EnumTypes[enumT]

	if !enumRegistered {
		panic(fmt.Errorf("RegisterEnumTypeVariant: enum type %v is not registered", enumT))
	}

	for existingID, registeredVariant := range registeredVariants {
		if newVariantT == registeredVariant {
			panic(fmt.Errorf("RegisterEnumType: variant type %v of enum %v is already registered under id %v instead of %v",
				newVariantT, enumT, existingID, id))
		}
	}

	registeredVariants[id] = newVariantT

	// Returnign something just for a conveniece of using this function in a single line in global scope like:
	// var _ = RegisterEnumTypeVariant[EnumType](id, newVariant)
	return struct{}{}
}

func RegisterEnumTypeWithIDs[EnumType any](variants map[EnumVariantID]any) struct{} {
	enumT := reflect.TypeOf((*EnumType)(nil)).Elem()

	if enumT.Kind() != reflect.Interface {
		panic(fmt.Errorf("RegisterEnumType: enum type %v is not an interface", enumT))
	}

	alreadyRegisteredVariants := EnumTypes[enumT]
	if alreadyRegisteredVariants != nil {
		panic(fmt.Errorf("RegisterEnumType: enum type %v is already registered with variants %v", enumT, alreadyRegisteredVariants))
	}

	registeredSuccessufly := false
	EnumTypes[enumT] = make(map[EnumVariantID]reflect.Type, len(variants))

	defer func() {
		if !registeredSuccessufly {
			delete(EnumTypes, enumT)
		}
	}()

	for id, v := range variants {
		RegisterEnumTypeVariant[EnumType](id, v)
	}

	registeredSuccessufly = true

	return struct{}{}
}

func RegisterEnumType[EnumType any](variants ...any) struct{} {
	variantsMap := make(map[EnumVariantID]any, len(variants))

	for i, v := range variants {
		variantsMap[EnumVariantID(i)] = v
	}

	return RegisterEnumTypeWithIDs[EnumType](variantsMap)
}

func RegisterEnumType1[EnumType any, Variant1 any]() struct{} {
	var variant1 Variant1
	return RegisterEnumType[EnumType](variant1)
}

func RegisterEnumType2[EnumType any, Variant1 any, Variant2 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	return RegisterEnumType[EnumType](variant1, variant2)
}

func RegisterEnumType3[EnumType any, Variant1 any, Variant2 any, Variant3 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	return RegisterEnumType[EnumType](variant1, variant2, variant3)
}

func RegisterEnumType4[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4)
}

func RegisterEnumType5[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5)
}

func RegisterEnumType6[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6)
}

func RegisterEnumType7[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any, Variant7 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	var variant7 Variant7
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6, variant7)
}

func RegisterEnumType8[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any, Variant7 any, Variant8 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	var variant7 Variant7
	var variant8 Variant8
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6, variant7, variant8)
}

func RegisterEnumType9[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any, Variant7 any, Variant8 any, Variant9 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	var variant7 Variant7
	var variant8 Variant8
	var variant9 Variant9
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6, variant7, variant8, variant9)
}

func RegisterEnumType10[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any, Variant7 any, Variant8 any, Variant9 any, Variant10 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	var variant7 Variant7
	var variant8 Variant8
	var variant9 Variant9
	var variant10 Variant10
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6, variant7, variant8, variant9, variant10)
}

func RegisterEnumType11[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any, Variant7 any, Variant8 any, Variant9 any, Variant10 any, Variant11 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	var variant7 Variant7
	var variant8 Variant8
	var variant9 Variant9
	var variant10 Variant10
	var variant11 Variant11
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6, variant7, variant8, variant9, variant10, variant11)
}

func RegisterEnumType12[EnumType any, Variant1 any, Variant2 any, Variant3 any, Variant4 any, Variant5 any, Variant6 any, Variant7 any, Variant8 any, Variant9 any, Variant10 any, Variant11 any, Variant12 any]() struct{} {
	var variant1 Variant1
	var variant2 Variant2
	var variant3 Variant3
	var variant4 Variant4
	var variant5 Variant5
	var variant6 Variant6
	var variant7 Variant7
	var variant8 Variant8
	var variant9 Variant9
	var variant10 Variant10
	var variant11 Variant11
	var variant12 Variant12
	return RegisterEnumType[EnumType](variant1, variant2, variant3, variant4, variant5, variant6, variant7, variant8, variant9, variant10, variant11, variant12)
}
