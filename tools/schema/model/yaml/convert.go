package yaml

import (
	"github.com/iotaledger/wasp/tools/schema/model"
)

const (
	KeyName        string = "name"
	KeyDescription string = "description"
	KeyEvents      string = "events"
	KeyStructs     string = "structs"
	KeyTypedefs    string = "typedefs"
	KeyState       string = "state"
	KeyFuncs       string = "funcs"
	KeyViews       string = "views"
	KeyAccess      string = "access"
	KeyParams      string = "params"
	KeyResults     string = "results"
)

func Convert(root *Node, def *model.SchemaDef) error {
	var name, description model.DefElt
	var events, structs model.DefMapMap
	var typedefs, state model.DefMap
	var funcs, views model.FuncDefMap

	for _, key := range root.Contents {
		switch key.Val {
		case KeyName:
			name.Val = key.Contents[0].Val
			name.Line = key.Line
		case KeyDescription:
			description.Val = key.Contents[0].Val
			description.Line = key.Line
		case KeyEvents:
			events = *key.ToDefMapMap()
		case KeyStructs:
			structs = *key.ToDefMapMap()
		case KeyTypedefs:
			typedefs = *key.ToDefMap()
		case KeyState:
			state = *key.ToDefMap()
		case KeyFuncs:
			funcs = *key.ToFuncDefMap()
		case KeyViews:
			views = *key.ToFuncDefMap()
		default:
		}
	}
	def.Name = name
	def.Description = description
	def.Events = events
	def.Structs = structs
	def.Typedefs = typedefs
	def.State = state
	def.Funcs = funcs
	def.Views = views

	return nil
}

func (n *Node) ToDefElt() *model.DefElt {
	return &model.DefElt{
		Val:     n.Val,
		Comment: n.Comment,
		Line:    n.Line,
	}
}

func (n *Node) ToDefMap() *model.DefMap {
	defs := make(model.DefMap)
	for _, yamlKey := range n.Contents {
		key := *yamlKey.ToDefElt()
		defs[key] = yamlKey.Contents[0].ToDefElt()
	}
	return &defs
}

func (n *Node) ToDefMapMap() *model.DefMapMap {
	defs := make(model.DefMapMap)
	for _, yamlKey := range n.Contents {
		key := model.DefElt{
			Val:     yamlKey.Val,
			Comment: yamlKey.Comment,
			Line:    yamlKey.Line,
		}

		defs[key] = yamlKey.ToDefMap()
	}
	return &defs
}

func (n *Node) ToFuncDef() *model.FuncDef {
	def := model.FuncDef{}
	for _, yamlKey := range n.Contents {
		switch yamlKey.Val {
		case KeyAccess:
			def.Access = *yamlKey.Contents[0].ToDefElt()
		case KeyParams:
			def.Params = *yamlKey.ToDefMap()
		case KeyResults:
			def.Results = *yamlKey.ToDefMap()
		default:
		}

	}
	return &def
}

func (n *Node) ToFuncDefMap() *model.FuncDefMap {
	defs := make(model.FuncDefMap)
	for _, yamlKey := range n.Contents {
		key := model.DefElt{
			Val:     yamlKey.Val,
			Comment: yamlKey.Comment,
			Line:    yamlKey.Line,
		}

		defs[key.Val] = yamlKey.ToFuncDef()
	}
	return &defs
}
