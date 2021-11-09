// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
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
	Name          string
	FullName      string
	Description   string
	KeyID         int
	ConstLen      int
	ConstNames    []string
	ConstValues   []string
	CoreContracts bool
	SchemaTime    time.Time
	Funcs         []*Func
	Params        []*Field
	Results       []*Field
	StateVars     []*Field
	Structs       []*Struct
	Typedefs      []*Field
	Views         []*Func
}

func NewSchema() *Schema {
	return &Schema{}
}

func (s *Schema) Compile(schemaDef *SchemaDef) error {
	s.FullName = strings.TrimSpace(schemaDef.Name)
	if s.FullName == "" {
		return fmt.Errorf("missing contract name")
	}
	s.Name = lower(s.FullName)
	s.Description = strings.TrimSpace(schemaDef.Description)

	err := s.compileStructs(schemaDef)
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
	s.KeyID = 0
	for _, name := range sortedFields(params) {
		param := params[name]
		param.KeyID = s.KeyID
		s.KeyID++
		s.Params = append(s.Params, param)
	}
	for _, name := range sortedFields(results) {
		result := results[name]
		result.KeyID = s.KeyID
		s.KeyID++
		s.Results = append(s.Results, result)
	}
	return s.compileStateVars(schemaDef)
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
			return nil, fmt.Errorf("redefined %s type", what)
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
		varDef.KeyID = s.KeyID
		s.KeyID++
		s.StateVars = append(s.StateVars, varDef)
	}
	return nil
}

func (s *Schema) compileStructs(schemaDef *SchemaDef) error {
	for _, typeName := range sortedMaps(schemaDef.Structs) {
		fieldMap := schemaDef.Structs[typeName]
		typeDef := &Struct{}
		typeDef.Name = typeName
		fieldNames := make(StringMap)
		fieldAliases := make(StringMap)
		for _, fldName := range sortedKeys(fieldMap) {
			fldType := fieldMap[fldName]
			field, err := s.compileField(fldName, fldType)
			if err != nil {
				return err
			}
			if field.Optional {
				return fmt.Errorf("type field cannot be optional")
			}
			if _, ok := fieldNames[field.Name]; ok {
				return fmt.Errorf("duplicate field name")
			}
			fieldNames[field.Name] = field.Name
			if _, ok := fieldAliases[field.Alias]; ok {
				return fmt.Errorf("duplicate field alias")
			}
			fieldAliases[field.Alias] = field.Alias
			typeDef.Fields = append(typeDef.Fields, field)
		}
		s.Structs = append(s.Structs, typeDef)
	}
	return nil
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
			return fmt.Errorf("duplicate sybtype name")
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
