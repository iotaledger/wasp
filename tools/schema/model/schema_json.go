package model

type JSONFuncDef struct {
	Access  string    `json:"access,omitempty"`
	Params  StringMap `json:"params,omitempty"`
	Results StringMap `json:"results,omitempty"`
}
type JSONFuncDefMap map[string]*JSONFuncDef

type JSONSchemaDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Events      StringMapMap   `json:"events"`
	Structs     StringMapMap   `json:"structs"`
	Typedefs    StringMap      `json:"typedefs"`
	State       StringMap      `json:"state"`
	Funcs       JSONFuncDefMap `json:"funcs"`
	Views       JSONFuncDefMap `json:"views"`
}

func (s *JSONSchemaDef) ToSchemaDef() *SchemaDef {
	def := NewSchemaDef()
	def.Name = DefElt{Val: s.Name}
	def.Description = DefElt{Val: s.Description}
	def.Events = s.Events.ToDefMapMap()
	def.Structs = s.Structs.ToDefMapMap()
	def.State = s.State.ToDefMap()
	def.Typedefs = s.Typedefs.ToDefMap()
	def.Funcs = s.Funcs.ToFuncDefMap()
	def.Views = s.Views.ToFuncDefMap()
	return def
}

func (mm StringMapMap) ToDefMapMap() DefMapMap {
	defs := make(DefMapMap, len(mm))
	for key, valmap := range mm {
		m := valmap.ToDefMap()
		defs[DefElt{Val: key}] = &m
	}
	return defs
}

func (mm StringMap) ToDefMap() DefMap {
	defs := make(DefMap, len(mm))
	for key, val := range mm {
		defs[DefElt{Val: key}] = &DefElt{Val: val}
	}
	return defs
}

func (m JSONFuncDefMap) ToFuncDefMap() FuncDefMap {
	defs := make(FuncDefMap)
	for key, val := range m {
		defs[key] = &FuncDef{
			Access:  DefElt{Val: val.Access},
			Params:  val.Params.ToDefMap(),
			Results: val.Results.ToDefMap(),
		}
	}
	return defs
}
