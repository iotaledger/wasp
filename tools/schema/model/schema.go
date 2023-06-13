// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
)

const (
	DefaultVersion = "0.0.1"
)

type Func struct {
	Name    string
	Alias   string
	Access  DefElt
	Comment string
	Hname   isc.Hname
	Kind    string
	Line    int
	Params  []*Field
	Results []*Field
}

type Struct struct {
	Name   DefElt
	Fields []*Field
}

type Schema struct {
	ContractName  string
	PackageName   string
	Author        string
	Copyright     string
	Description   string
	License       string
	Repository    string
	Version       string
	CoreContracts bool
	SchemaTime    time.Time
	Events        []*Struct
	Funcs         []*Func
	Params        []*Field
	Results       []*Field
	StateVars     []*Field
	Structs       []*Struct
	Typedefs      []*Field
}

func NewSchema() *Schema {
	return &Schema{}
}

func (s *Schema) Compile(schemaDef *SchemaDef) error {
	s.ContractName = schemaDef.Name.Val
	if s.ContractName == "" {
		return errors.New("missing contract name")
	}
	s.PackageName = strings.ToLower(s.ContractName)
	s.Author = schemaDef.Author.Val
	s.Copyright = schemaDef.Copyright
	s.Description = schemaDef.Description.Val
	s.License = schemaDef.License.Val
	s.Repository = schemaDef.Repository.Val
	s.Version = schemaDef.Version.Val
	if s.Version == "" {
		s.Version = DefaultVersion
	}

	err := s.compileEvents(schemaDef)
	if err != nil {
		return err
	}
	err = s.compileStructs(schemaDef)
	if err != nil {
		return err
	}
	err = s.compileTypeDefs(schemaDef)
	if err != nil {
		return err
	}
	params := make(FieldMap)
	results := make(FieldMap)
	err = s.compileFuncs(schemaDef, &params, &results, false)
	if err != nil {
		return err
	}
	err = s.compileFuncs(schemaDef, &params, &results, true)
	if err != nil {
		return err
	}
	for _, name := range sortedFields(params) {
		s.Params = append(s.Params, params[name])
	}
	for _, name := range sortedFields(results) {
		s.Results = append(s.Results, results[name])
	}
	return s.compileStateVars(schemaDef)
}

func (s *Schema) compileEvents(schemaDef *SchemaDef) error {
	for _, eventName := range sortedMaps(schemaDef.Events) {
		event, err := s.compileStruct("event", eventName, *schemaDef.Events[eventName])
		if err != nil {
			return err
		}
		s.Events = append(s.Events, event)
	}
	return nil
}

func (s *Schema) compileField(fldName, fldType *DefElt) (*Field, error) {
	field := &Field{}
	err := field.Compile(s, fldName, fldType)
	if err != nil {
		return nil, err
	}
	return field, nil
}

func (s *Schema) compileFuncs(schemaDef *SchemaDef, params, results *FieldMap, views bool) (err error) {
	funcKind := "func"
	templateFuncs := schemaDef.Funcs
	if views {
		funcKind = "view"
		templateFuncs = schemaDef.Views
	}
	for _, funcName := range sortedFuncDescs(templateFuncs) {
		if views && schemaDef.Funcs[funcName] != nil {
			return fmt.Errorf("duplicate func/view name: %s at line %d", funcName.Val, templateFuncs[funcName].Line)
		}
		funcDesc := templateFuncs[funcName]
		if funcDesc == nil {
			funcDesc = &FuncDef{}
		}

		f := &Func{}
		fName := strings.TrimSpace(funcName.Val)
		f.Name = fName
		f.Alias = fName
		index := strings.Index(fName, "=")
		if index >= 0 {
			f.Name = strings.TrimSpace(fName[:index])
			f.Alias = strings.TrimSpace(fName[index+1:])
		}
		f.Kind = funcKind
		f.Hname = isc.Hn(f.Alias)
		f.Line = funcName.Line
		f.Comment = funcName.Comment

		// check for Hname collision
		for _, other := range s.Funcs {
			if other.Hname == f.Hname {
				return fmt.Errorf("hname collision: %d (%s and %s) at line %d and %d",
					f.Hname, f.Name, other.Name, f.Line, other.Line)
			}
		}

		f.Access = funcDesc.Access
		f.Params, err = s.compileFuncFields(funcDesc.Params, params, "param")
		if err != nil {
			return err
		}
		f.Results, err = s.compileFuncFields(funcDesc.Results, results, "result")
		if err != nil {
			return err
		}
		s.Funcs = append(s.Funcs, f)
	}
	return nil
}

func (s *Schema) compileFuncFields(fieldMap DefMap, allFieldMap *FieldMap, what string) ([]*Field, error) {
	fields := make([]*Field, 0, len(fieldMap))
	fieldNames := make(DefNameMap)
	fieldAliases := make(DefNameMap)
	for _, fldKey := range sortedKeys(fieldMap) {
		fldName := fldKey
		fldType := fieldMap[fldName]
		field, err := s.compileField(&fldName, fldType)
		if err != nil {
			return nil, err
		}
		tmpfld, ok := fieldNames[field.Name]
		if ok {
			return nil, fmt.Errorf("duplicate %s name at line %d and %d", what, tmpfld.Line, field.Line)
		}
		fieldNames[field.Name] = &DefElt{
			Val:     field.Name,
			Line:    field.Line,
			Comment: field.Comment,
		}
		tmpfld, ok = fieldAliases[field.Alias]
		if ok {
			return nil, fmt.Errorf("duplicate %s alias at line %d and %d", what, tmpfld.Line, field.Line)
		}
		fieldAliases[field.Alias] = &DefElt{
			Val:     field.Alias,
			Line:    field.Line,
			Comment: field.Comment,
		}
		existing, ok := (*allFieldMap)[field.Name]
		if !ok {
			(*allFieldMap)[field.Name] = field
			existing = field
		}
		if existing.Alias != field.Alias {
			return nil, fmt.Errorf("redefined %s alias: '%s' != '%s' at line %d and %d", what, existing.Alias, field.Alias, existing.Line, field.Line)
		}
		if existing.Type != field.Type {
			return nil, fmt.Errorf("redefined %s type: %s at line %d and %d", what, field.Name, existing.Line, field.Line)
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (s *Schema) compileStateVars(schemaDef *SchemaDef) error {
	varNames := make(DefNameMap)
	varAliases := make(DefNameMap)
	for _, varKey := range sortedKeys(schemaDef.State) {
		varName := varKey
		varType := schemaDef.State[varName]
		varDef, err := s.compileField(&varName, varType)
		if err != nil {
			return err
		}
		varState, ok := varNames[varDef.Name]
		if ok {
			return fmt.Errorf("duplicate var name: %s at line %d and %d", varState.Val, varState.Line, varDef.Line)
		}
		varNames[varDef.Name] = &DefElt{
			Val:     varDef.Name,
			Line:    varDef.Line,
			Comment: varDef.Comment,
		}
		varState, ok = varAliases[varDef.Alias]
		if ok {
			return fmt.Errorf("duplicate var alias: %s at line %d and %d", varState.Val, varState.Line, varDef.Line)
		}
		varAliases[varDef.Alias] = &DefElt{
			Val:     varDef.Alias,
			Line:    varDef.Line,
			Comment: varDef.Comment,
		}
		s.StateVars = append(s.StateVars, varDef)
	}
	return nil
}

func (s *Schema) compileStructs(schemaDef *SchemaDef) error {
	for _, structName := range sortedMaps(schemaDef.Structs) {
		structDef, err := s.compileStruct("struct", structName, *schemaDef.Structs[structName])
		if err != nil {
			return err
		}
		s.Structs = append(s.Structs, structDef)
	}
	return nil
}

func (s *Schema) compileStruct(kind string, structName DefElt, structFields DefMap) (*Struct, error) {
	structDef := &Struct{Name: structName}
	fieldNames := make(DefNameMap)
	fieldAliases := make(DefNameMap)
	for _, fldKey := range sortedKeys(structFields) {
		fldName := fldKey
		fldType := structFields[fldName]
		field, err := s.compileField(&fldName, fldType)
		if err != nil {
			return nil, err
		}
		if field.IsOptional {
			return nil, fmt.Errorf("%s field cannot be optional", kind)
		}
		if field.IsArray {
			return nil, fmt.Errorf("%s field cannot be an array", kind)
		}
		if field.MapKey != "" {
			return nil, fmt.Errorf("%s field cannot be a map", kind)
		}
		tmpfld, ok := fieldNames[field.Name]
		if ok {
			return nil, fmt.Errorf("duplicate %s field name at line %d and %d", kind, tmpfld.Line, field.Line)
		}
		fieldNames[field.Name] = &DefElt{
			Val:  field.Name,
			Line: field.Line,
		}
		tmpfld, ok = fieldAliases[field.Alias]
		if ok {
			return nil, fmt.Errorf("duplicate %s field alias at line %d and %d", kind, tmpfld.Line, field.Line)
		}
		fieldAliases[field.Alias] = &DefElt{
			Val:  field.Alias,
			Line: field.Line,
		}
		structDef.Fields = append(structDef.Fields, field)
	}
	return structDef, nil
}

func (s *Schema) compileTypeDefs(schemaDef *SchemaDef) error {
	varNames := make(DefNameMap)
	varAliases := make(DefNameMap)
	for _, varKey := range sortedKeys(schemaDef.Typedefs) {
		varName := varKey
		varType := schemaDef.Typedefs[varName]
		varDef, err := s.compileField(&varName, varType)
		if err != nil {
			return err
		}
		tmpvar, ok := varNames[varDef.Name]
		if ok {
			return fmt.Errorf("duplicate subtype name at line %d and %d", tmpvar.Line, varDef.Line)
		}
		varNames[varDef.Name] = &DefElt{
			Val:  varDef.Name,
			Line: varDef.Line,
		}
		tmpvar, ok = varAliases[varDef.Alias]
		if ok {
			return fmt.Errorf("duplicate subtype alias at line %d and %d", tmpvar.Line, varDef.Line)
		}
		varAliases[varDef.Alias] = &DefElt{
			Val:  varDef.Alias,
			Line: varDef.Line,
		}
		s.Typedefs = append(s.Typedefs, varDef)
	}
	return nil
}

func sortedFields(dict FieldMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return strings.ToLower(keys[i]) < strings.ToLower(keys[j]) })
	return keys
}

type DefEltList []DefElt

func (l DefEltList) Len() int           { return len(l) }
func (l DefEltList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l DefEltList) Less(i, j int) bool { return strings.ToLower(l[i].Val) < strings.ToLower(l[j].Val) }

func sortedKeys(dict DefMap) []DefElt {
	keys := make(DefEltList, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Sort(keys)
	return keys
}

func sortedMaps(dict DefMapMap) []DefElt {
	keys := make(DefEltList, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Sort(keys)
	return keys
}

func sortedFuncDescs(dict FuncDefMap) []DefElt {
	keys := make(DefEltList, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Sort(keys)
	return keys
}
