// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"errors"
	"strings"

	"github.com/iotaledger/wasp/tools/schema/model"
)

const (
	KeyAccess      string = "access"
	KeyAuthor      string = "author"
	KeyCopyright   string = "copyright"
	KeyDescription string = "description"
	KeyEvents      string = "events"
	KeyFuncs       string = "funcs"
	KeyLicense     string = "license"
	KeyName        string = "name"
	KeyParams      string = "params"
	KeyRepository  string = "repository"
	KeyResults     string = "results"
	KeyState       string = "state"
	KeyStructs     string = "structs"
	KeyTypedefs    string = "typedefs"
	KeyVersion     string = "version"
	KeyViews       string = "views"
)

//nolint:funlen,gocyclo
func Convert(root *Node, def *model.SchemaDef) error {
	var name, description, author, version, license, repository model.DefElt
	var events, structs model.DefMapMap
	var typedefs, state model.DefMap
	var funcs, views model.FuncDefMap
	for _, key := range root.Contents {
		switch key.Val {
		case KeyCopyright:
			if len(key.Contents) > 0 && key.Contents[0].Val != "" {
				def.Copyright = ("// " + strings.TrimSuffix(key.Val, "\n"))
			} else {
				def.Copyright = key.HeadComment
			}
		case KeyLicense:
			license.Val = key.Contents[0].Val
			license.Line = key.Line
		case KeyRepository:
			repository.Val = key.Contents[0].Val
			repository.Line = key.Line
		case KeyName:
			name.Val = key.Contents[0].Val
			name.Line = key.Line
		case KeyDescription:
			description.Val = key.Contents[0].Val
			description.Line = key.Line
		case KeyAuthor:
			author.Val = key.Contents[0].Val
			author.Line = key.Line
		case KeyEvents:
			events = key.ToDefMapMap()
		case KeyStructs:
			structs = key.ToDefMapMap()
		case KeyTypedefs:
			typedefs = key.ToDefMap()
		case KeyState:
			state = key.ToDefMap()
		case KeyFuncs:
			funcs = key.ToFuncDefMap()
		case KeyVersion:
			version.Val = key.Contents[0].Val
			version.Line = key.Line
		case KeyViews:
			views = key.ToFuncDefMap()
		default:
			return errors.New("unsupported key")
		}
	}
	def.Name = name
	def.Author = author
	def.License = license
	def.Repository = repository
	def.Description = description
	def.Version = version
	def.Events = events
	def.State = state
	def.Structs = structs
	def.Typedefs = typedefs
	def.Funcs = funcs
	def.Views = views
	return nil
}

func (n *Node) ToDefElt() *model.DefElt {
	comment := ""
	if len(n.HeadComment) > 0 {
		// remove trailing '\n'
		comment = n.HeadComment[:len(n.HeadComment)-1]
	} else if len(n.LineComment) > 0 {
		// remove trailing '\n'
		comment = n.LineComment[:len(n.LineComment)-1]
	}
	return &model.DefElt{
		Val:     n.Val,
		Comment: comment,
		Line:    n.Line,
	}
}

func (n *Node) ToDefMap() model.DefMap {
	defs := make(model.DefMap)
	for _, yamlKey := range n.Contents {
		if strings.ReplaceAll(yamlKey.Val, " ", "") == "{}" {
			// treat "{}" as empty
			continue
		}
		key := *yamlKey.ToDefElt()
		val := yamlKey.Contents[0].ToDefElt()
		if val.Comment != "" && key.Comment == "" {
			key.Comment = val.Comment
		}
		val.Comment = ""
		defs[key] = val
	}
	return defs
}

func (n *Node) ToDefMapMap() model.DefMapMap {
	defs := make(model.DefMapMap)
	for _, yamlKey := range n.Contents {
		// TODO better parsing
		if strings.ReplaceAll(yamlKey.Val, " ", "") == "{}" {
			// treat "{}" as empty
			continue
		}
		comment := ""
		if len(yamlKey.HeadComment) > 0 {
			comment = yamlKey.HeadComment[:len(yamlKey.HeadComment)-1] // remove trailing '\n'
		} else if len(yamlKey.LineComment) > 0 {
			comment = yamlKey.LineComment[:len(yamlKey.LineComment)-1] // remove trailing '\n'
		}

		key := model.DefElt{
			Val:     yamlKey.Val,
			Comment: comment,
			Line:    yamlKey.Line,
		}
		val := yamlKey.ToDefMap()
		defs[key] = &val
	}
	return defs
}

func (n *Node) ToFuncDef() model.FuncDef {
	def := model.FuncDef{}
	def.Line = n.Line
	if len(n.HeadComment) > 0 {
		def.Comment = n.HeadComment[:len(n.HeadComment)-1] // remove trailing '\n'
	} else if len(n.LineComment) > 0 {
		def.Comment = n.LineComment[:len(n.LineComment)-1] // remove trailing '\n'
	}

	for _, yamlKey := range n.Contents {
		if strings.ReplaceAll(yamlKey.Val, " ", "") == "{}" {
			// treat "{}" as empty
			continue
		}
		switch yamlKey.Val {
		case KeyAccess:
			def.Access = *yamlKey.Contents[0].ToDefElt()
			if len(yamlKey.HeadComment) > 0 {
				def.Access.Comment = yamlKey.HeadComment[:len(yamlKey.HeadComment)-1] // remove trailing '\n'
			} else if len(yamlKey.LineComment) > 0 {
				def.Access.Comment = yamlKey.LineComment[:len(yamlKey.LineComment)-1] // remove trailing '\n'
			}
		case KeyParams:
			def.Params = yamlKey.ToDefMap()
		case KeyResults:
			def.Results = yamlKey.ToDefMap()
		default:
			return model.FuncDef{}
		}
	}
	return def
}

func (n *Node) ToFuncDefMap() model.FuncDefMap {
	defs := make(model.FuncDefMap)
	for _, yamlKey := range n.Contents {
		if strings.ReplaceAll(yamlKey.Val, " ", "") == "{}" {
			// treat "{}" as empty
			continue
		}
		comment := ""
		if len(yamlKey.HeadComment) > 0 {
			comment = yamlKey.HeadComment[:len(yamlKey.HeadComment)-1] // remove trailing '\n'
		} else if len(yamlKey.LineComment) > 0 {
			comment = yamlKey.LineComment[:len(yamlKey.LineComment)-1] // remove trailing '\n'
		}
		key := model.DefElt{
			Val:     yamlKey.Val,
			Comment: comment,
			Line:    yamlKey.Line,
		}
		val := yamlKey.ToFuncDef()
		defs[key] = &val
	}
	return defs
}
