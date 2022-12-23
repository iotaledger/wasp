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
	Array      bool
	FldComment string
	MapKey     string
	Optional   bool
	Type       string
	BaseType   bool
	Comment    string
	Line       int // the line number originally in yaml file
}

//nolint:gocyclo
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
		return fmt.Errorf("invalid field name: %s at %d", f.Name, fldNameDef.Line)
	}
	if !fldAliasRegexp.MatchString(f.Alias) {
		return fmt.Errorf("invalid field alias: %s at %d", f.Alias, fldNameDef.Line)
	}

	fldType := strings.TrimSpace(fldTypeDef.Val)

	// remove // comment
	index = strings.Index(fldType, "//")
	if index >= 0 {
		f.FldComment = " " + fldType[index:]
		fldType = strings.TrimSpace(fldType[:index])
	}

	// remove optional indicator
	n := len(fldType)
	if n > 1 && fldType[n-1:] == "?" {
		f.Optional = true
		fldType = strings.TrimSpace(fldType[:n-1])
	}

	n = len(fldType)
	if n > 2 && fldType[n-2:] == "[]" {
		// must be array
		f.Array = true
		fldType = strings.TrimSpace(fldType[:n-2])
	} else if n > 4 && fldType[:4] == "map[" {
		// must be map
		index = strings.Index(fldType, "]")
		if index > 5 {
			f.MapKey = strings.TrimSpace(fldType[4:index])
			if !fldTypeRegexp.MatchString(f.MapKey) || f.MapKey == "Bool" {
				return fmt.Errorf("invalid key field type: %s at %d", f.MapKey, fldTypeDef.Line)
			}
			fldType = strings.TrimSpace(fldType[index+1:])
		}
	}
	f.Type = fldType
	if !fldTypeRegexp.MatchString(f.Type) {
		return fmt.Errorf("invalid field type: %s at %d", f.Type, fldTypeDef.Line)
	}
	f.BaseType = FieldTypes[f.Type]
	if f.BaseType {
		return nil
	}
	for _, typeDef := range s.Structs {
		if f.Type == typeDef.Name.Val {
			return nil
		}
	}
	for _, subtype := range s.Typedefs {
		if f.Type == subtype.Name {
			return nil
		}
	}
	return fmt.Errorf("invalid field type: %s at %d", f.Type, fldTypeDef.Line)
}
