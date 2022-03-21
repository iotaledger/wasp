package model

type RawFuncDef struct {
	Access  string    `json:"access,omitempty" yaml:"access,omitempty"`
	Params  StringMap `json:"params,omitempty" yaml:"params,omitempty"`
	Results StringMap `json:"results,omitempty" yaml:"results,omitempty"`
}
type RawFuncDefMap map[string]*RawFuncDef

type RawSchemaDef struct {
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Events      StringMapMap  `json:"events" yaml:"events"`
	Structs     StringMapMap  `json:"structs" yaml:"structs"`
	Typedefs    StringMap     `json:"typedefs" yaml:"typedefs"`
	State       StringMap     `json:"state" yaml:"state"`
	Funcs       RawFuncDefMap `json:"funcs" yaml:"funcs"`
	Views       RawFuncDefMap `json:"views" yaml:"views"`
}

type JSONSchemaDef RawSchemaDef

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

func (m RawFuncDefMap) ToFuncDefMap() FuncDefMap {
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

func (m DefMapMap) ToStringMapMap() StringMapMap {
	ret := make(StringMapMap)
	for key, val := range m {
		ret[key.Val] = val.ToStringMap()
	}
	return ret
}

func (m DefMap) ToStringMap() StringMap {
	ret := make(StringMap)
	for key, val := range m {
		ret[key.Val] = val.Val
	}
	return ret
}

func (m FuncDefMap) ToRawFuncDefMap() RawFuncDefMap {
	ret := make(RawFuncDefMap)
	for key, val := range m {
		ret[key] = &RawFuncDef{
			Access:  val.Access.Val,
			Params:  val.Params.ToStringMap(),
			Results: val.Results.ToStringMap(),
		}
	}
	return ret
}

func (s *SchemaDef) ToRawSchemaDef() *RawSchemaDef {
	def := &RawSchemaDef{}
	def.Structs = s.Structs.ToStringMapMap()
	def.Events = s.Events.ToStringMapMap()
	def.Typedefs = s.Typedefs.ToStringMap()
	def.State = s.State.ToStringMap()
	def.Funcs = s.Funcs.ToRawFuncDefMap()
	def.Views = s.Views.ToRawFuncDefMap()
	return def
}
