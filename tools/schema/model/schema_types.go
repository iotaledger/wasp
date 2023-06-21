// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

type (
	FieldMap       map[string]*Field
	StringMap      map[string]string
	StringMapMap   map[string]StringMap
	DefMap         map[DefElt]*DefElt
	DefNameMap     map[string]*DefElt
	DefMapMap      map[DefElt]*DefMap
	DefNameMapMap  map[string]*DefMap
	FuncDefMap     map[DefElt]*FuncDef
	FuncNameDefMap map[string]*FuncDef
)

type FuncDef struct {
	Access  DefElt
	Params  DefMap
	Results DefMap
	Line    int
	Comment string
}

type DefElt struct {
	Val     string
	Comment string
	Line    int
}

type SchemaDef struct {
	Name        DefElt
	Author      DefElt
	Copyright   string
	Description DefElt
	License     DefElt
	Repository  DefElt
	Version     DefElt
	Events      DefMapMap
	Structs     DefMapMap
	Typedefs    DefMap
	State       DefMap
	Funcs       FuncDefMap
	Views       FuncDefMap
}

func NewSchemaDef() *SchemaDef {
	def := &SchemaDef{}
	def.Events = make(DefMapMap)
	def.Structs = make(DefMapMap)
	def.Typedefs = make(DefMap)
	def.State = make(DefMap)
	def.Funcs = make(FuncDefMap)
	def.Views = make(FuncDefMap)
	return def
}
