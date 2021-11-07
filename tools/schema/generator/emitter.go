package generator

import (
	"regexp"
	"strings"
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

func (g *GenBase) emit(template string) {
	lines := strings.Split(g.templates[template], "\n")
	for i := 1; i < len(lines)-1; i++ {
		line := lines[i]

		// replace any placeholder keys
		line = emitKeyRegExp.ReplaceAllStringFunc(line, func(key string) string {
			text, ok := g.keys[key[1:]]
			if ok {
				return text
			}
			return "???:" + key
		})

		// remove concatenation markers
		line = strings.ReplaceAll(line, "$+", "")

		// now process special commands
		if strings.HasPrefix(line, "$#") {
			if strings.HasPrefix(line, "$#emit ") {
				g.emit(strings.TrimSpace(line[7:]))
				continue
			}
			if strings.HasPrefix(line, "$#each ") {
				g.emitEach(line)
				continue
			}
			if strings.HasPrefix(line, "$#func ") {
				g.emitFunc(line)
				continue
			}
			if strings.HasPrefix(line, "$#if ") {
				g.emitIf(line)
				continue
			}
			if strings.HasPrefix(line, "$#set ") {
				g.emitSet(line)
				continue
			}
			g.println("???:" + line)
			continue
		}

		g.println(line)
	}
}

func (g *GenBase) emitEach(key string) {
	parts := strings.Split(key, " ")
	if len(parts) != 3 {
		g.println("???:" + key)
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
		g.println("???:" + key)
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

func (g *GenBase) emitFunc(key string) {
	parts := strings.Split(key, " ")
	if len(parts) != 2 {
		g.println("???:" + key)
		return
	}

	emitter, ok := g.emitters[parts[1]]
	if ok {
		emitter(g)
		return
	}
	g.println("???:" + key)
}

//nolint:funlen
func (g *GenBase) emitIf(key string) {
	parts := strings.Split(key, " ")
	if len(parts) < 3 || len(parts) > 4 {
		g.println("???:" + key)
		return
	}

	conditionKey := parts[1]
	template := parts[2]

	condition := false
	switch conditionKey {
	case KeyArray:
		condition = g.currentField.Array
	case KeyBaseType:
		condition = g.currentField.TypeID != 0
	case KeyExist:
		proxy := g.keys[KeyProxy]
		condition = g.newTypes[proxy]
	case KeyInit:
		condition = g.currentFunc.Name == KeyInit
	case KeyMap:
		condition = g.currentField.MapKey != ""
	case KeyMut:
		condition = g.keys[KeyMut] == "Mutable"
	case KeyCore:
		condition = g.s.CoreContracts
	case KeyFunc, KeyView:
		condition = g.keys["kind"] == conditionKey
	case KeyParam:
		condition = len(g.currentFunc.Params) != 0
	case KeyParams:
		condition = len(g.s.Params) != 0
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
	case KeyPtrs:
		condition = len(g.currentFunc.Params) != 0 || len(g.currentFunc.Results) != 0
	default:
		g.println("???:" + key)
		return
	}

	if condition {
		g.emit(template)
		return
	}

	// else branch?
	if len(parts) == 4 {
		template = parts[3]
		g.emit(template)
	}
}

func (g *GenBase) emitSet(line string) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		g.println("???:" + line)
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
