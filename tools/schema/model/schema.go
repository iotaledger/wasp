// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
)

// TODO describe schema details in docs
type (
	FieldMap     map[string]*Field
	FieldMapMap  map[string]FieldMap
	StringMap    map[string]string
	StringMapMap map[string]StringMap
)

type FuncDef struct {
	Access  string    `json:"access,omitempty" yaml:"access,omitempty"`
	Params  StringMap `json:"params,omitempty" yaml:"params,omitempty"`
	Results StringMap `json:"results,omitempty" yaml:"results,omitempty"`
}
type FuncDefMap map[string]*FuncDef

type SchemaDef struct {
	Name        string       `json:"name" yaml:"name"`
	Description string       `json:"description" yaml:"description"`
	Events      StringMapMap `json:"events" yaml:"events"`
	Structs     StringMapMap `json:"structs" yaml:"structs"`
	Typedefs    StringMap    `json:"typedefs" yaml:"typedefs"`
	State       StringMap    `json:"state" yaml:"state"`
	Funcs       FuncDefMap   `json:"funcs" yaml:"funcs"`
	Views       FuncDefMap   `json:"views" yaml:"views"`
}

type Func struct {
	Name    string
	Access  string
	Kind    string
	Hname   iscp.Hname
	Params  []*Field
	Results []*Field
}

type Struct struct {
	Name   string
	Fields []*Field
}

type Schema struct {
	ContractName  string
	PackageName   string
	Description   string
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
	s.ContractName = strings.TrimSpace(schemaDef.Name)
	if s.ContractName == "" {
		return fmt.Errorf("missing contract name")
	}
	s.PackageName = strings.ToLower(s.ContractName)
	s.Description = strings.TrimSpace(schemaDef.Description)

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
		event, err := s.compileStruct("event", eventName, schemaDef.Events[eventName])
		if err != nil {
			return err
		}
		s.Events = append(s.Events, event)
	}
	return nil
}

func (s *Schema) compileField(fldName, fldType string) (*Field, error) {
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
			return fmt.Errorf("duplicate func/view name: %s", funcName)
		}
		funcDesc := templateFuncs[funcName]
		if funcDesc == nil {
			funcDesc = &FuncDef{}
		}

		f := &Func{}
		f.Name = funcName
		f.Kind = funcKind
		f.Hname = iscp.Hn(funcName)

		// check for Hname collision
		for _, other := range s.Funcs {
			if other.Hname == f.Hname {
				return fmt.Errorf("hname collision: %d (%s and %s)", f.Hname, f.Name, other.Name)
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

func (s *Schema) compileFuncFields(fieldMap StringMap, allFieldMap *FieldMap, what string) ([]*Field, error) {
	fields := make([]*Field, 0, len(fieldMap))
	fieldNames := make(StringMap)
	fieldAliases := make(StringMap)
	for _, fldName := range sortedKeys(fieldMap) {
		fldType := fieldMap[fldName]
		field, err := s.compileField(fldName, fldType)
		if err != nil {
			return nil, err
		}
		if _, ok := fieldNames[field.Name]; ok {
			return nil, fmt.Errorf("duplicate %s name", what)
		}
		fieldNames[field.Name] = field.Name
		if _, ok := fieldAliases[field.Alias]; ok {
			return nil, fmt.Errorf("duplicate %s alias", what)
		}
		fieldAliases[field.Alias] = field.Alias
		existing, ok := (*allFieldMap)[field.Name]
		if !ok {
			(*allFieldMap)[field.Name] = field
			existing = field
		}
		if existing.Alias != field.Alias {
			return nil, fmt.Errorf("redefined %s alias: '%s' != '%s", what, existing.Alias, field.Alias)
		}
		if existing.Type != field.Type {
			return nil, fmt.Errorf("redefined %s type: %s", what, field.Name)
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (s *Schema) compileStateVars(schemaDef *SchemaDef) error {
	varNames := make(StringMap)
	varAliases := make(StringMap)
	for _, varName := range sortedKeys(schemaDef.State) {
		varType := schemaDef.State[varName]
		varDef, err := s.compileField(varName, varType)
		if err != nil {
			return err
		}
		if _, ok := varNames[varDef.Name]; ok {
			return fmt.Errorf("duplicate var name")
		}
		varNames[varDef.Name] = varDef.Name
		if _, ok := varAliases[varDef.Alias]; ok {
			return fmt.Errorf("duplicate var alias")
		}
		varAliases[varDef.Alias] = varDef.Alias
		s.StateVars = append(s.StateVars, varDef)
	}
	return nil
}

func (s *Schema) compileStructs(schemaDef *SchemaDef) error {
	for _, structName := range sortedMaps(schemaDef.Structs) {
		structDef, err := s.compileStruct("struct", structName, schemaDef.Structs[structName])
		if err != nil {
			return err
		}
		s.Structs = append(s.Structs, structDef)
	}
	return nil
}

func (s *Schema) compileStruct(kind, structName string, structFields StringMap) (*Struct, error) {
	structDef := &Struct{Name: structName}
	fieldNames := make(StringMap)
	fieldAliases := make(StringMap)
	for _, fldName := range sortedKeys(structFields) {
		fldType := structFields[fldName]
		field, err := s.compileField(fldName, fldType)
		if err != nil {
			return nil, err
		}
		if field.Optional {
			return nil, fmt.Errorf("%s field cannot be optional", kind)
		}
		if field.Array {
			return nil, fmt.Errorf("%s field cannot be an array", kind)
		}
		if field.MapKey != "" {
			return nil, fmt.Errorf("%s field cannot be a map", kind)
		}
		if _, ok := fieldNames[field.Name]; ok {
			return nil, fmt.Errorf("duplicate %s field name", kind)
		}
		fieldNames[field.Name] = field.Name
		if _, ok := fieldAliases[field.Alias]; ok {
			return nil, fmt.Errorf("duplicate %s field alias", kind)
		}
		fieldAliases[field.Alias] = field.Alias
		structDef.Fields = append(structDef.Fields, field)
	}
	return structDef, nil
}

func (s *Schema) compileTypeDefs(schemaDef *SchemaDef) error {
	varNames := make(StringMap)
	varAliases := make(StringMap)
	for _, varName := range sortedKeys(schemaDef.Typedefs) {
		varType := schemaDef.Typedefs[varName]
		varDef, err := s.compileField(varName, varType)
		if err != nil {
			return err
		}
		if _, ok := varNames[varDef.Name]; ok {
			return fmt.Errorf("duplicate subtype name")
		}
		varNames[varDef.Name] = varDef.Name
		if _, ok := varAliases[varDef.Alias]; ok {
			return fmt.Errorf("duplicate subtype alias")
		}
		varAliases[varDef.Alias] = varDef.Alias
		s.Typedefs = append(s.Typedefs, varDef)
	}
	return nil
}

func sortedFields(dict FieldMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedFuncDescs(dict FuncDefMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(dict StringMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedMaps(dict StringMapMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
