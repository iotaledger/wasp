package model

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
)

// TODO describe schema details in docs
type (
	FieldMap      map[string]*Field
	FieldMapMap   map[string]FieldMap
	StringMap     map[string]string
	StringMapMap  map[string]StringMap
	DefMap        map[DefElt]*DefElt
	DefNameMap    map[string]*DefElt
	DefMapMap     map[DefElt]*DefMap
	DefNameMapMap map[string]*DefMap
)

type FuncDef struct {
	Access  DefElt
	Params  DefMap
	Results DefMap
	Line    int
	Comment string
}
type FuncDefMap map[string]*FuncDef

type DefElt struct {
	Val     string
	Comment string
	Line    int
}

type SchemaDef struct {
	Name        DefElt
	Description DefElt
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

type Func struct {
	Name    string
	Access  DefElt
	Kind    string
	Hname   iscp.Hname
	Params  []*Field
	Results []*Field
	Line    int
	Comment string
}

type Struct struct {
	Name   DefElt
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
