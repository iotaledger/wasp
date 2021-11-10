package generator

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
)

const (
	KeyArray     = "array"
	KeyBaseType  = "basetype"
	KeyCore      = "core"
	KeyExist     = "exist"
	KeyFunc      = "func"
	KeyInit      = "init"
	KeyMandatory = "mandatory"
	KeyMap       = "map"
	KeyMut       = "mut"
	KeyParam     = "param"
	KeyParams    = "params"
	KeyProxy     = "proxy"
	KeyPtrs      = "ptrs"
	KeyResult    = "result"
	KeyResults   = "results"
	KeyState     = "state"
	KeyStruct    = "struct"
	KeyStructs   = "structs"
	KeyThis      = "this"
	KeyTypeDef   = "typedef"
	KeyTypeDefs  = "typedefs"
	KeyView      = "view"
)

var emitKeyRegExp = regexp.MustCompile(`\$[a-zA-Z_]+`)

// emit processes "$#emit template"
// It processes all lines in the named template
// If the template is non-existent nothing will happen
// Any line starting with a special "$#" directive will recursively be processed
// An unknown directive will result in an error
func (g *GenBase) emit(template string) {
	lines := strings.Split(g.templates[template], "\n")
	for i := 1; i < len(lines)-1; i++ {
		// replace any placeholder keys
		line := emitKeyRegExp.ReplaceAllStringFunc(lines[i], func(key string) string {
			text, ok := g.keys[key[1:]]
			if ok {
				return text
			}
			return "???:" + key
		})

		// remove concatenation markers
		line = strings.ReplaceAll(line, "$+", "")

		// line contains special directive?
		space := strings.Index(line, " ")
		if space <= 2 || line[:2] != "$#" {
			// no special directive, just emit line
			g.println(line)
			continue
		}

		// now process special directive
		switch line[2:space] {
		case "emit":
			g.emit(strings.TrimSpace(line[7:]))
		case "each":
			g.emitEach(line)
		case "func":
			g.emitFunc(line)
		case "if":
			g.emitIf(line)
		case "set":
			g.emitSet(line)
		default:
			g.error(line)
		}
	}
}

// emitEach processes "$#each array template"
// It processes the template for each item in the array
// Produces an error if the array key is unknown
func (g *GenBase) emitEach(line string) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		g.error(line)
		return
	}

	template := parts[2]
	switch parts[1] {
	case KeyFunc:
		g.emitEachFunc(g.s.Funcs, template)
	case KeyMandatory:
		g.emitEachMandatoryField(template)
	case KeyParam:
		g.emitEachField(g.currentFunc.Params, template)
	case KeyParams:
		g.emitEachField(g.s.Params, template)
	case KeyResult:
		g.emitEachField(g.currentFunc.Results, template)
	case KeyResults:
		g.emitEachField(g.s.Results, template)
	case KeyState:
		g.emitEachField(g.s.StateVars, template)
	case KeyStruct:
		g.emitEachField(g.currentStruct.Fields, template)
	case KeyStructs:
		g.emitEachStruct(g.s.Structs, template)
	case KeyTypeDef:
		g.emitEachField(g.s.Typedefs, template)
	default:
		g.error(line)
	}
}

func (g *GenBase) emitEachField(fields []*Field, template string) {
	g.maxCamelFldLen = 0
	g.maxSnakeFldLen = 0
	for _, g.currentField = range fields {
		camelLen := len(g.currentField.Name)
		if g.maxCamelFldLen < camelLen {
			g.maxCamelFldLen = camelLen
		}
		snakeLen := len(snake(g.currentField.Name))
		if g.maxSnakeFldLen < snakeLen {
			g.maxSnakeFldLen = snakeLen
		}
	}

	for _, g.currentField = range fields {
		g.gen.setFieldKeys(true)
		g.emit(template)
	}
}

func (g *GenBase) emitEachFunc(funcs []*Func, template string) {
	g.maxCamelFuncLen = 0
	g.maxSnakeFuncLen = 0
	for _, g.currentFunc = range funcs {
		camelLen := len(g.currentFunc.Name)
		if g.maxCamelFuncLen < camelLen {
			g.maxCamelFuncLen = camelLen
		}
		snakeLen := len(snake(g.currentFunc.Name))
		if g.maxSnakeFuncLen < snakeLen {
			g.maxSnakeFuncLen = snakeLen
		}
	}

	for _, g.currentFunc = range funcs {
		g.gen.setFuncKeys()
		g.emit(template)
	}
}

func (g *GenBase) emitEachMandatoryField(template string) {
	mandatoryFields := make([]*Field, 0)
	for _, g.currentField = range g.currentFunc.Params {
		if !g.currentField.Optional {
			mandatoryFields = append(mandatoryFields, g.currentField)
		}
	}
	g.emitEachField(mandatoryFields, template)
}

func (g *GenBase) emitEachStruct(structs []*Struct, template string) {
	for _, g.currentStruct = range structs {
		g.setMultiKeyValues("strName", g.currentStruct.Name)
		g.emit(template)
	}
}

// emitFunc processes "$#func emitter"
// It can call back into go code to emit more complex stuff
// Produces an error if emitter is unknown
func (g *GenBase) emitFunc(line string) {
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		g.error(line)
		return
	}

	emitter, ok := g.emitters[parts[1]]
	if ok {
		emitter(g)
		return
	}
	g.error(line)
}

// emitIf processes "$#if condition template [elseTemplate]"
// It processes template when the named condition is true
// It processes the optional elseTemplate when the named condition is false
// Produces an error if named condition is unknown
//nolint:funlen
func (g *GenBase) emitIf(line string) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 || len(parts) > 4 {
		g.error(line)
		return
	}

	condition := false
	switch parts[1] {
	case KeyArray:
		condition = g.currentField.Array
	case KeyBaseType:
		condition = g.currentField.TypeID != 0
	case KeyCore:
		condition = g.s.CoreContracts
	case KeyExist:
		condition = g.newTypes[g.keys[KeyProxy]]
	case KeyFunc:
		condition = g.keys["kind"] == KeyFunc
	case KeyInit:
		condition = g.currentFunc.Name == KeyInit
	case KeyMap:
		condition = g.currentField.MapKey != ""
	case KeyMut:
		condition = g.keys[KeyMut] == "Mutable"
	case KeyParam:
		condition = len(g.currentFunc.Params) != 0
	case KeyParams:
		condition = len(g.s.Params) != 0
	case KeyPtrs:
		condition = len(g.currentFunc.Params) != 0 || len(g.currentFunc.Results) != 0
	case KeyResult:
		condition = len(g.currentFunc.Results) != 0
	case KeyResults:
		condition = len(g.s.Results) != 0
	case KeyState:
		condition = len(g.s.StateVars) != 0
	case KeyStructs:
		condition = len(g.s.Structs) != 0
	case KeyThis:
		condition = g.currentField.Alias == KeyThis
	case KeyTypeDef:
		condition = g.fieldIsTypeDef()
	case KeyTypeDefs:
		condition = len(g.s.Typedefs) != 0
	case KeyView:
		condition = g.keys["kind"] == KeyView
	default:
		key, ok := g.keys[parts[1]]
		if !ok {
			g.error(line)
			return
		}
		condition = key != ""
	}

	if condition {
		g.emit(parts[2])
		return
	}

	// else branch?
	if len(parts) == 4 {
		g.emit(parts[3])
	}
}

// emitSet processes "$#set key value"
// It sets the specified key to value, which can be anything
// Just make sure there is a space after the key name
// The special key "exist" is used to add a newly generated type
// It can be used to prevent duplicate types from being generated
func (g *GenBase) emitSet(line string) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		g.error(line)
		return
	}

	key := parts[1]
	value := line[len(parts[0])+len(key)+2:]
	g.keys[key] = value

	if key == KeyExist {
		g.newTypes[value] = true
	}
}

func (g *GenBase) fieldIsTypeDef() bool {
	for _, typeDef := range g.s.Typedefs {
		if typeDef.Name == g.currentField.Type {
			g.currentField = typeDef
			g.gen.setFieldKeys(false)
			return true
		}
	}
	return false
}

func (g *GenBase) setCommonKeys() {
	g.keys["false"] = ""
	g.keys["true"] = "true"
	g.keys["empty"] = ""
	g.keys["space"] = " "
	g.keys["package"] = g.s.Name
	g.keys["Package"] = g.s.FullName
	g.keys["module"] = moduleName + strings.Replace(moduleCwd[len(modulePath):], "\\", "/", -1)
	scName := g.s.Name
	if g.s.CoreContracts {
		// strip off "core" prefix
		scName = scName[4:]
	}
	g.keys["scName"] = scName
	g.keys["hscName"] = iscp.Hn(scName).String()
	g.keys["scDesc"] = g.s.Description
	g.keys["maxIndex"] = strconv.Itoa(g.s.KeyID)
}

func (g *GenBase) setFieldKeys(pad bool) {
	g.setMultiKeyValues("fldName", g.currentField.Name)
	g.setMultiKeyValues("fldType", g.currentField.Type)

	g.keys["fldAlias"] = g.currentField.Alias
	g.keys["fldComment"] = g.currentField.Comment
	g.keys["fldMapKey"] = g.currentField.MapKey
	g.keys["fldIndex"] = strconv.Itoa(g.currentField.KeyID)

	if pad {
		g.keys["fldPad"] = spaces[:g.maxCamelFldLen-len(g.keys["fldName"])]
		g.keys["fld_pad"] = spaces[:g.maxSnakeFldLen-len(g.keys["fld_name"])]
	}
}

func (g *GenBase) setFuncKeys() {
	g.setMultiKeyValues("funcName", g.currentFunc.Name)
	g.setMultiKeyValues("kind", g.currentFunc.Kind)
	g.keys["funcHName"] = iscp.Hn(g.keys["funcName"]).String()
	grant := g.currentFunc.Access
	comment := ""
	index := strings.Index(grant, "//")
	if index >= 0 {
		comment = grant[index:]
		grant = strings.TrimSpace(grant[:index])
	}
	g.keys["funcAccess"] = grant
	g.keys["funcAccessComment"] = comment

	g.keys["funcPad"] = spaces[:g.maxCamelFuncLen-len(g.keys["funcName"])]
	g.keys["func_pad"] = spaces[:g.maxSnakeFuncLen-len(g.keys["func_name"])]
}

func (g *GenBase) setMultiKeyValues(key, value string) {
	value = uncapitalize(value)
	g.keys[key] = value
	g.keys[capitalize(key)] = capitalize(value)
	g.keys[snake(key)] = snake(value)
	g.keys[upper(snake(key))] = upper(snake(value))
}
