// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	fldNameRegexp  = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	fldAliasRegexp = regexp.MustCompile(`^[a-zA-Z0-9_$#@*%\-]+$`)
	fldTypeRegexp  = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]+$`)
)

var FieldTypes = map[string]bool{
	"Address":   true,
	"AgentID":   true,
	"BigInt":    true,
	"Bool":      true,
	"Bytes":     true,
	"ChainID":   true,
	"Hash":      true,
	"Hname":     true,
	"Int8":      true,
	"Int16":     true,
	"Int32":     true,
	"Int64":     true,
	"NftID":     true,
	"RequestID": true,
	"String":    true,
	"TokenID":   true,
	"Uint8":     true,
	"Uint16":    true,
	"Uint32":    true,
	"Uint64":    true,
}

type Field struct {
	Name       string // external name for this field
	Alias      string // internal name alias, can be different from Name
	Comment    string
	FldComment string
	IsArray    bool
	IsBaseType bool
	IsOptional bool
	Line       int // the line number originally in yaml file
	MapKey     string
	Type       string
}

func (f *Field) Compile(s *Schema, fldNameDef, fldTypeDef *DefElt) error {
	fldName := strings.TrimSpace(fldNameDef.Val)
	f.Name = fldName
	f.Alias = fldName
	f.Line = fldNameDef.Line
	f.Comment = fldNameDef.Comment
	index := strings.Index(fldName, "=")
	if index >= 0 {
		f.Name = strings.TrimSpace(fldName[:index])
		f.Alias = strings.TrimSpace(fldName[index+1:])
	}
	if !fldNameRegexp.MatchString(f.Name) {
		return fmt.Errorf("invalid field name: %s at line %d", f.Name, fldNameDef.Line)
	}
	if !fldAliasRegexp.MatchString(f.Alias) {
		return fmt.Errorf("invalid field alias: %s at line %d", f.Alias, fldNameDef.Line)
	}

	fldType, err := f.compileFieldType(fldTypeDef)
	if err != nil {
		return err
	}
	f.Type = fldType
	f.IsBaseType = FieldTypes[fldType]
	if f.IsBaseType {
		return nil
	}
	for _, typeDef := range s.Structs {
		if fldType == typeDef.Name.Val {
			return nil
		}
	}
	for _, subtype := range s.Typedefs {
		if fldType == subtype.Name {
			return nil
		}
	}
	return fmt.Errorf("invalid field type: %s at line %d", fldType, fldTypeDef.Line)
}

func (f *Field) compileFieldType(fldTypeDef *DefElt) (string, error) {
	fldType := strings.TrimSpace(fldTypeDef.Val)

	// strip 'optional' indicator
	if strings.HasSuffix(fldType, "?") {
		f.IsOptional = true
		fldType = strings.TrimSpace(fldType[:len(fldType)-1])
	}

	switch {
	case strings.HasSuffix(fldType, "[]"): // is it an array?
		f.IsArray = true
		fldType = strings.TrimSpace(fldType[:len(fldType)-2])
	case strings.HasPrefix(fldType, "map["): // is it a map?
		parts := strings.Split(fldType[4:], "]")
		if len(parts) != 2 {
			return "", fmt.Errorf("expected map field type: %s at line %d", fldType, fldTypeDef.Line)
		}
		f.MapKey = strings.TrimSpace(parts[0])
		if !fldTypeRegexp.MatchString(f.MapKey) || f.MapKey == "Bool" {
			return "", fmt.Errorf("invalid map key field type: %s at line %d", f.MapKey, fldTypeDef.Line)
		}
		fldType = strings.TrimSpace(parts[1])
	}
	if !fldTypeRegexp.MatchString(fldType) {
		return "", fmt.Errorf("invalid field type: %s at line %d", fldType, fldTypeDef.Line)
	}
	return fldType, nil
}
