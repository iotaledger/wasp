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
)

var topLevelFields []string = []string{
	KeyEvents,
	KeyStructs,
	KeyTypedefs,
	KeyState,
	KeyFuncs,
	KeyViews,
}

func Unmarshal(root *Node, def *model.SchemaDef) {
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
		case KeyStructs:
		case KeyTypedefs:
		case KeyState:
		case KeyFuncs:
		case KeyViews:
		default:
		}
	}
}
