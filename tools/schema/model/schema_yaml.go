package model

import "gopkg.in/yaml.v3"

type YAMLSchemaDef struct {
	Name        yaml.Node
	Description yaml.Node
	Events      yaml.Node
	Structs     yaml.Node
	Typedefs    yaml.Node
	State       yaml.Node
	Funcs       yaml.Node
	Views       yaml.Node
}

func (s *YAMLSchemaDef) ToSchemaDef() *SchemaDef {
	def := NewSchemaDef()
	def.Name.FromYAMLNode(&s.Name)
	def.Description.FromYAMLNode(&s.Description)
	def.Events.FromYAMLNode(&s.Events)
	def.State.FromYAMLNode(&s.State)
	def.Structs.FromYAMLNode(&s.Structs)
	def.Typedefs.FromYAMLNode(&s.Typedefs)
	def.Funcs.FromYAMLNode(&s.Funcs)
	def.Views.FromYAMLNode(&s.Views)
	return def
}

func (e *DefElt) FromYAMLNode(n *yaml.Node) *DefElt {
	e.Val = n.Value
	e.Line = n.Line
	e.Comment = n.LineComment
	return e
}

func (m *DefMap) FromYAMLNode(n *yaml.Node) *DefMap {
	for i := 0; i < len(n.Content); i += 2 {
		key := DefElt{
			Val:  n.Content[i].Value,
			Line: n.Content[i].Line,
			// Take only the comment after value
			Comment: n.Content[i].LineComment,
		}
		(*m)[key] = &DefElt{
			Val:  n.Content[i+1].Value,
			Line: n.Content[i+1].Line,
			// Take only the comment after value
			Comment: n.Content[i+1].LineComment,
		}
	}
	return m
}

func (mm *DefMapMap) FromYAMLNode(n *yaml.Node) *DefMapMap {
	for i := 0; i < len(n.Content); i += 2 {
		m := make(DefMap)
		key := DefElt{
			Val:  n.Content[i].Value,
			Line: n.Content[i].Line,
			// Take only the comment after value
			Comment: n.Content[i].LineComment,
		}
		(*mm)[key] = m.FromYAMLNode(n.Content[i+1])
	}
	return mm
}

func (m *FuncDefMap) FromYAMLNode(n *yaml.Node) *FuncDefMap {
	for i := 0; i < len(n.Content); i += 2 {
		d := &FuncDef{}
		d.Line = n.Content[i].Line
		d.Comment = n.Content[i].LineComment
		d.Params = make(DefMap)
		d.Results = make(DefMap)
		vals := n.Content[i+1]
		for j := 0; j < len(vals.Content); j += 2 {
			switch vals.Content[j].Value {
			case "access":
				d.Access.FromYAMLNode(vals.Content[j+1])
			case "params":
				d.Params.FromYAMLNode(vals.Content[j+1])
			case "results":
				d.Results.FromYAMLNode(vals.Content[j+1])
			default:
				return nil
			}
		}

		(*m)[n.Content[i].Value] = d
	}
	return m
}
