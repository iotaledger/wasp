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
	"Bool":      true,
	"Bytes":     true,
	"ChainID":   true,
	"Color":     true,
	"Hash":      true,
	"Hname":     true,
	"Int8":      true,
	"Int16":     true,
	"Int32":     true,
	"Int64":     true,
	"RequestID": true,
	"String":    true,
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
	Line       int // the line number originally in yaml/json file
}

func (f *Field) Compile(s *Schema, fldName, fldType *DefElt) error {
	fldName.Val = strings.TrimSpace(fldName.Val)
	f.Name = fldName.Val
	f.Alias = fldName.Val
	f.Line = fldName.Line
	f.Comment = fldName.Comment
	index := strings.Index(fldName.Val, "=")
	if index >= 0 {
		f.Name = strings.TrimSpace(fldName.Val[:index])
		f.Alias = strings.TrimSpace(fldName.Val[index+1:])
	}
	if !fldNameRegexp.MatchString(f.Name) {
		return fmt.Errorf("invalid field name: %s at %d", f.Name, fldName.Line)
	}
	if !fldAliasRegexp.MatchString(f.Alias) {
		return fmt.Errorf("invalid field alias: %s at %d", f.Alias, fldName.Line)
	}

	fldType.Val = strings.TrimSpace(fldType.Val)

	// remove // comment
	index = strings.Index(fldType.Val, "//")
	if index >= 0 {
		f.FldComment = " " + fldType.Val[index:]
		fldType.Val = strings.TrimSpace(fldType.Val[:index])
	}

	// remove optional indicator
	n := len(fldType.Val)
	if n > 1 && fldType.Val[n-1:] == "?" {
		f.Optional = true
		fldType.Val = strings.TrimSpace(fldType.Val[:n-1])
	}

	n = len(fldType.Val)
	if n > 2 && fldType.Val[n-2:] == "[]" {
		// must be array
		f.Array = true
		fldType.Val = strings.TrimSpace(fldType.Val[:n-2])
	} else if n > 4 && fldType.Val[:4] == "map[" {
		// must be map
		index = strings.Index(fldType.Val, "]")
		if index > 5 {
			f.MapKey = strings.TrimSpace(fldType.Val[4:index])
			if !fldTypeRegexp.MatchString(f.MapKey) || f.MapKey == "Bool" {
				return fmt.Errorf("invalid key field type: %s at %d", f.MapKey, fldType.Line)
			}
			fldType.Val = strings.TrimSpace(fldType.Val[index+1:])
		}
	}
	f.Type = fldType.Val
	if !fldTypeRegexp.MatchString(f.Type) {
		return fmt.Errorf("invalid field type: %s at %d", f.Type, fldType.Line)
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
	return fmt.Errorf("invalid field type: %s at %d", f.Type, fldType.Line)
}
