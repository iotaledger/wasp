package iotago

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
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
	parser := &structTagParser{
		input: []rune(data),
		pos:   0,
	}
	return parser.parseStructTag()
}

// Token types for the struct tag parser
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenAddr
	tokenIdent
	tokenDoubleColon // ::
	tokenLessThan    // <
	tokenGreaterThan // >
	tokenComma       // ,
	tokenPrimitive
)

type token struct {
	type_ tokenType
	value string
	pos   int
}

type structTagParser struct {
	input []rune
	pos   int
}

func (p *structTagParser) skipWhitespace() {
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ' ' || c == '\t' {
			p.pos++
		} else {
			break
		}
	}
}

func (p *structTagParser) nextToken() (token, error) {
	p.skipWhitespace()

	if p.pos >= len(p.input) {
		return token{tokenEOF, "", p.pos}, nil
	}

	start := p.pos
	c := p.input[p.pos]

	// Check for :: first
	if c == ':' && p.pos+1 < len(p.input) && p.input[p.pos+1] == ':' {
		p.pos += 2
		return token{tokenDoubleColon, "::", start}, nil
	}

	// Single character tokens
	switch c {
	case '<':
		p.pos++
		return token{tokenLessThan, "<", start}, nil
	case '>':
		p.pos++
		return token{tokenGreaterThan, ">", start}, nil
	case ',':
		p.pos++
		return token{tokenComma, ",", start}, nil
	}

	// Address: 0x followed by hex digits
	if c == '0' && p.pos+1 < len(p.input) && (p.input[p.pos+1] == 'x' || p.input[p.pos+1] == 'X') {
		p.pos += 2 // Skip 0x
		hexStart := p.pos

		for p.pos < len(p.input) {
			c := p.input[p.pos]
			if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
				p.pos++
			} else {
				break
			}
		}

		if p.pos == hexStart {
			return token{}, fmt.Errorf("invalid address at position %d: expected hex digits after 0x", start)
		}

		hexLen := p.pos - hexStart
		if hexLen > 64 {
			return token{}, fmt.Errorf("invalid address at position %d: hex part too long (%d digits, max 64)", start, hexLen)
		}

		value := string(p.input[start:p.pos])
		return token{tokenAddr, value, start}, nil
	}

	// Check for primitives and identifiers
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
		for p.pos < len(p.input) {
			c := p.input[p.pos]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
				p.pos++
			} else {
				break
			}
		}

		value := string(p.input[start:p.pos])

		// Check if it's a primitive type
		switch value {
		case "bool", "u8", "u16", "u32", "u64", "u128", "u256", "address", "signer":
			return token{tokenPrimitive, value, start}, nil
		default:
			return token{tokenIdent, value, start}, nil
		}
	}

	return token{}, fmt.Errorf("unexpected character '%c' at position %d", c, start)
}

func (p *structTagParser) parseStructTagCore() (*StructTag, error) {
	// Address
	addrToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if addrToken.type_ != tokenAddr {
		return nil, fmt.Errorf("expected address at position %d, got %s", addrToken.pos, addrToken.value)
	}

	address, err := AddressFromHex(addrToken.value)
	if err != nil {
		return nil, fmt.Errorf("invalid address '%s' at position %d: %w", addrToken.value, addrToken.pos, err)
	}

	// First ::
	dcolonToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if dcolonToken.type_ != tokenDoubleColon {
		return nil, fmt.Errorf("expected '::' at position %d, got '%s'", dcolonToken.pos, dcolonToken.value)
	}

	// Module identifier
	moduleToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if moduleToken.type_ != tokenIdent {
		return nil, fmt.Errorf("expected module identifier at position %d, got '%s'", moduleToken.pos, moduleToken.value)
	}

	// Second ::
	dcolonToken2, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if dcolonToken2.type_ != tokenDoubleColon {
		return nil, fmt.Errorf("expected '::' at position %d, got '%s'", dcolonToken2.pos, dcolonToken2.value)
	}

	// Name identifier
	nameToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if nameToken.type_ != tokenIdent {
		return nil, fmt.Errorf("expected struct name identifier at position %d, got '%s'", nameToken.pos, nameToken.value)
	}

	// Optional type parameters
	var typeParams []TypeTag
	nextTok, err := p.nextToken()
	if err != nil {
		return nil, err
	}

	if nextTok.type_ == tokenLessThan {
		typeParams, err = p.parseTypeParams()
		if err != nil {
			return nil, err
		}

		// final >
		gtToken, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if gtToken.type_ != tokenGreaterThan {
			return nil, fmt.Errorf("expected '>' at position %d, got '%s'", gtToken.pos, gtToken.value)
		}
	} else {
		// Put back the token for the caller to consume
		p.pos -= len([]rune(nextTok.value))
	}

	return &StructTag{
		Address:    address,
		Module:     Identifier(moduleToken.value),
		Name:       Identifier(nameToken.value),
		TypeParams: typeParams,
	}, nil
}

func (p *structTagParser) parseStructTag() (*StructTag, error) {
	// address
	addrToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if addrToken.type_ != tokenAddr {
		return nil, fmt.Errorf("expected address at position %d, got %s", addrToken.pos, addrToken.value)
	}

	address, err := AddressFromHex(addrToken.value)
	if err != nil {
		return nil, fmt.Errorf("invalid address '%s' at position %d: %w", addrToken.value, addrToken.pos, err)
	}

	// first ::
	dcolonToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if dcolonToken.type_ != tokenDoubleColon {
		return nil, fmt.Errorf("expected '::' at position %d, got '%s'", dcolonToken.pos, dcolonToken.value)
	}

	// module identifier
	moduleToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if moduleToken.type_ != tokenIdent {
		return nil, fmt.Errorf("expected module identifier at position %d, got '%s'", moduleToken.pos, moduleToken.value)
	}

	// second ::
	dcolonToken2, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if dcolonToken2.type_ != tokenDoubleColon {
		return nil, fmt.Errorf("expected '::' at position %d, got '%s'", dcolonToken2.pos, dcolonToken2.value)
	}

	// name identifier
	nameToken, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if nameToken.type_ != tokenIdent {
		return nil, fmt.Errorf("expected struct name identifier at position %d, got '%s'", nameToken.pos, nameToken.value)
	}

	// optional type parameters
	var typeParams []TypeTag
	nextTok, err := p.nextToken()
	if err != nil {
		return nil, err
	}

	switch nextTok.type_ {
	case tokenLessThan:
		typeParams, err = p.parseTypeParams()
		if err != nil {
			return nil, err
		}

		// final >
		gtToken, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if gtToken.type_ != tokenGreaterThan {
			return nil, fmt.Errorf("expected '>' at position %d, got '%s'", gtToken.pos, gtToken.value)
		}

		// check for eof
		eofToken, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if eofToken.type_ != tokenEOF {
			return nil, fmt.Errorf("expected end of input at position %d, got '%s'", eofToken.pos, eofToken.value)
		}
	case tokenEOF:
		// No type parameters, which is fine
	default:
		return nil, fmt.Errorf("expected '<' or end of input at position %d, got '%s'", nextTok.pos, nextTok.value)
	}

	return &StructTag{
		Address:    address,
		Module:     Identifier(moduleToken.value),
		Name:       Identifier(nameToken.value),
		TypeParams: typeParams,
	}, nil
}

func (p *structTagParser) parseTypeParams() ([]TypeTag, error) {
	var typeParams []TypeTag

	for {
		typeTag, err := p.parseTypeTag()
		if err != nil {
			return nil, err
		}
		typeParams = append(typeParams, *typeTag)

		// Check for comma or end
		nextTok, err := p.nextToken()
		if err != nil {
			return nil, err
		}

		switch nextTok.type_ {
		case tokenComma:
			continue // Parse next type parameter
		case tokenGreaterThan:
			// Put back the > for the caller to consume
			p.pos -= len(nextTok.value)
			return typeParams, nil
		default:
			return nil, fmt.Errorf("expected ',' or '>' at position %d, got '%s'", nextTok.pos, nextTok.value)
		}
	}
}

func (p *structTagParser) parseTypeTag() (*TypeTag, error) {
	oldPos := p.pos
	nextTok, err := p.nextToken()
	if err != nil {
		return nil, err
	}

	if nextTok.type_ == tokenPrimitive {
		// It's a primitive type
		switch nextTok.value {
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
		case "address":
			return &TypeTag{Address: &serialization.EmptyEnum{}}, nil
		case "signer":
			return &TypeTag{Signer: &serialization.EmptyEnum{}}, nil
		default:
			return nil, fmt.Errorf("unknown primitive type '%s' at position %d", nextTok.value, nextTok.pos)
		}
	} else if nextTok.type_ == tokenAddr {
		// It's a struct tag - rewind and parse as struct
		p.pos = oldPos
		structTag, err := p.parseStructTagCore()
		if err != nil {
			return nil, err
		}
		return &TypeTag{Struct: structTag}, nil
	} else {
		return nil, fmt.Errorf("expected primitive type or address at position %d, got '%s'", nextTok.pos, nextTok.value)
	}
}

// splitGenericParameters is kept for backward compatibility with other parsing code
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
