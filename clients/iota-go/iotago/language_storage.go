package iotago

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
)

// refer BCS doc https://github.com/diem/bcs/blob/master/README.md#externally-tagged-enumerations
// IMPORTANT! The order of the fields MATTERS! DON'T CHANGE!
// this is enum `TypeTag` in `external-crates/move/crates/move-core-types/src/language_storage.rs`
// each field should be the same as enum `TypeTag` there
type TypeTag struct {
	Bool    *serialization.EmptyEnum
	U8      *serialization.EmptyEnum
	U64     *serialization.EmptyEnum
	U128    *serialization.EmptyEnum
	Address *serialization.EmptyEnum
	Signer  *serialization.EmptyEnum
	Vector  *TypeTag
	Struct  *StructTag

	U16  *serialization.EmptyEnum
	U32  *serialization.EmptyEnum
	U256 *serialization.EmptyEnum
}

func (t TypeTag) IsBcsEnum() {}

func (t *TypeTag) String() string {
	if t.Address != nil {
		return "address"
	} else if t.Signer != nil {
		return "signer"
	} else if t.Bool != nil {
		return "bool"
	} else if t.U8 != nil {
		return "u8"
	} else if t.U16 != nil {
		return "u16"
	} else if t.U32 != nil {
		return "u32"
	} else if t.U64 != nil {
		return "u64"
	} else if t.U128 != nil {
		return "u128"
	} else if t.U256 != nil {
		return "u256"
	} else if t.Vector != nil {
		return fmt.Sprintf("vector<%s>", t.Vector.String())
	} else if t.Struct != nil {
		return t.Struct.String()
	} else {
		panic("unknown type")
	}
}

func MustTypeTagFromString(data string) *TypeTag {
	tag, err := TypeTagFromString(data)
	if err != nil {
		panic(err)
	}
	return tag
}

// refer TypeTagSerializer.parseFromStr() at 'sdk/typescript/src/bcs/type-tag-serializer.ts'
func TypeTagFromString(data string) (*TypeTag, error) {
	switch data {
	case "address":
		return &TypeTag{Address: &serialization.EmptyEnum{}}, nil
	case "signer":
		return &TypeTag{Signer: &serialization.EmptyEnum{}}, nil
	case "bool":
		return &TypeTag{Bool: &serialization.EmptyEnum{}}, nil
	case "u8":
		return &TypeTag{U8: &serialization.EmptyEnum{}}, nil
	case "u16":
		return &TypeTag{U16: &serialization.EmptyEnum{}}, nil
	case "u32":
		return &TypeTag{U32: &serialization.EmptyEnum{}}, nil
	case "u64":
		return &TypeTag{U64: &serialization.EmptyEnum{}}, nil
	case "u128":
		return &TypeTag{U128: &serialization.EmptyEnum{}}, nil
	case "u256":
		return &TypeTag{U256: &serialization.EmptyEnum{}}, nil
	}

	vectorRegex := regexp.MustCompile(`^vector<(.+)>$`)
	structRegex := regexp.MustCompile(`^([^:]+)::([^:]+)::([^<]+)(<(.+)>)?$`)

	vectorMatches := vectorRegex.FindStringSubmatch(data)
	if len(vectorMatches) != 0 {
		typeTag, err := TypeTagFromString(vectorMatches[1])
		if err != nil {
			return nil, fmt.Errorf("can't parse %s to TypeTag: %w", data, err)
		}
		return &TypeTag{Vector: typeTag}, nil
	}

	structMatches := structRegex.FindStringSubmatch(data)
	if len(structMatches) != 0 {
		structTag := &StructTag{
			Address: MustAddressFromHex(structMatches[1]),
			Module:  Identifier(structMatches[2]),
			Name:    Identifier(structMatches[3]),
		}
		if len(structMatches) > 5 && structMatches[4] != "" {
			typeTag, err := parseStructTypeArgs(structMatches[5])
			if err != nil {
				return nil, fmt.Errorf("can't parse TypeParams of a Struct in TypeParams: %w", err)
			}
			structTag.TypeParams = typeTag
		}
		return &TypeTag{Struct: structTag}, nil
	}
	return nil, fmt.Errorf("not found")
}

func parseStructTypeArgs(structTypeParams string) ([]TypeTag, error) {
	var retTypeTag []TypeTag
	tokens := splitGenericParameters(structTypeParams, nil)
	for _, tok := range tokens {
		elt, err := TypeTagFromString(tok)
		if err != nil {
			return nil, fmt.Errorf("can't parse struct tag args: %w", err)
		}
		retTypeTag = append(retTypeTag, *elt)
	}
	return retTypeTag, nil
}

func (s *TypeTag) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	tag, err := TypeTagFromString(str)
	if err != nil {
		return err
	}
	*s = *tag
	return nil
}

type StructTag struct {
	Address    *Address
	Module     Identifier
	Name       Identifier
	TypeParams []TypeTag
}

func (s *StructTag) UnmarshalJSON(data []byte) error {
	str := string(data)
	str, _ = strings.CutPrefix(str, "\"")
	str, _ = strings.CutSuffix(str, "\"")
	parsedStructTag, err := StructTagFromString(str)
	if err != nil {
		return fmt.Errorf("can't unmarshal to %s StructTag: %w", str, err)
	}
	s.Address = parsedStructTag.Address
	s.Module = parsedStructTag.Module
	s.Name = parsedStructTag.Name
	s.TypeParams = parsedStructTag.TypeParams
	return nil
}

func (s *StructTag) MarshalJSON() ([]byte, error) {
	if s.Address == nil || s.Module == "" || s.Name == "" {
		return nil, fmt.Errorf("empty StructTag")
	}
	return []byte(fmt.Sprintf("%q", s.String())), nil
}

func (s *StructTag) String() string {
	if s.Address == nil || s.Module == "" || s.Name == "" {
		panic("empty StructTag")
	}
	typeParams := ""
	if len(s.TypeParams) > 0 {
		tmp := ""
		for i, typeTag := range s.TypeParams {
			typeTagString := ""
			if i != 0 {
				typeTagString = ", "
			}
			typeTagString += typeTag.String()
			tmp += typeTagString
		}
		typeParams = fmt.Sprintf("<%s>", tmp)
	}
	return s.Address.String() + "::" + s.Module + "::" + s.Name + typeParams
}

func StructTagFromString(data string) (*StructTag, error) {
	parts := strings.Split(data, "::")
	address, module := parts[0], parts[1]

	rest := data[len(address)+len(module)+4:]
	name := rest
	if idx := strings.Index(rest, "<"); idx > 0 {
		name = rest[:idx]
	}
	typeParams := []TypeTag{}

	if strings.Contains(rest, "<") {
		typeParamsRawStr := rest[strings.Index(rest, "<")+1 : strings.LastIndex(rest, ">")]
		typeParamsTokens := splitGenericParameters(typeParamsRawStr, []string{"<", ">"})
		typeParams = make([]TypeTag, len(typeParamsTokens))
		for i, token := range typeParamsTokens {
			param := TypeTag{}
			if !strings.Contains(token, "::") {
				typeTag, err := TypeTagFromString(token)
				if err != nil {
					return nil, fmt.Errorf("can't parse TypeParams: %w", err)
				}
				param = *typeTag
			} else {
				typeParam, err := StructTagFromString(token)
				if err != nil {
					return nil, fmt.Errorf("can't parse StructTag TypeParams: %w", err)
				}
				param.Struct = typeParam
			}

			typeParams[i] = param
		}
	}

	if len(typeParams) == 0 {
		typeParams = nil
	}

	return &StructTag{
		Address:    MustAddressFromHex(address),
		Module:     module,
		Name:       name,
		TypeParams: typeParams,
	}, nil
}

func splitGenericParameters(str string, genericSeparators []string) []string {
	var left, right string
	if genericSeparators != nil {
		left, right = genericSeparators[0], genericSeparators[1]
	} else {
		left, right = "", ""
	}

	var tokens []string
	word := ""
	nestedAngleBrackets := 0

	for i := 0; i < len(str); i++ {
		char := string(str[i])
		if char == left {
			nestedAngleBrackets++
		}
		if char == right {
			nestedAngleBrackets--
		}
		if nestedAngleBrackets == 0 && char == "," {
			tokens = append(tokens, strings.TrimSpace(word))
			word = ""
			continue
		}
		word += char
	}

	tokens = append(tokens, strings.TrimSpace(word))
	return tokens
}
