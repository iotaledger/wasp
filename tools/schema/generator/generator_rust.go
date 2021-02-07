// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bufio"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"os"
	"regexp"
)

var rustFuncRegexp = regexp.MustCompile("^pub fn (\\w+).+$")

var rustTypes = StringMap{
	"Address":    "ScAddress",
	"AgentId":    "ScAgentId",
	"ChainId":    "ScChainId",
	"Color":      "ScColor",
	"ContractId": "ScContractId",
	"Hash":       "ScHash",
	"Hname":      "ScHname",
	"Int":        "i64",
	"String":     "String",
}

func (s *Schema) GenerateRust() error {
	err := os.MkdirAll("src", 0755)
	if err != nil {
		return err
	}
	err = os.Chdir("src")
	if err != nil {
		return err
	}
	defer os.Chdir("..")

	err = s.GenerateRustLib()
	if err != nil {
		return err
	}
	err = s.GenerateRustSchema()
	if err != nil {
		return err
	}
	err = s.GenerateRustTypes()
	if err != nil {
		return err
	}
	return s.GenerateRustFuncs()
}

func (s *Schema) GenerateRustFunc(file *os.File, funcDef *FuncDef, isView bool) error {
	funcName := "func_"
	funcKind := "Call"
	if isView {
		funcName = "view_"
		funcKind = "View"
	}
	funcName += snake(funcDef.Name)
	fmt.Fprintf(file, "\npub fn %s(ctx: &Sc%sContext) {\n", funcName, funcKind)
	fmt.Fprintf(file, "    ctx.log(\"calling %s\");\n", funcDef.Name)
	fmt.Fprintf(file, "}\n")
	return nil
}

func (s *Schema) GenerateRustFuncs() error {
	scFileName := s.Name + ".rs"
	file, err := os.Open(scFileName)
	if err != nil {
		return s.GenerateRustFuncsNew(scFileName)
	}
	lines, existing, err := s.GenerateRustFuncScanner(file)
	if err != nil {
		return err
	}

	// save old one from overwrite
	scOriginal := s.Name + ".bak"
	err = os.Rename(scFileName, scOriginal)
	if err != nil {
		return err
	}
	file, err = os.Create(scFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// make copy of file
	for _, line := range lines {
		fmt.Fprintln(file, line)
	}

	// append any new funcs
	for _, funcDef := range s.Funcs {
		name := snake(funcDef.Name)
		if existing[name] == "" {
			err = s.GenerateRustFunc(file, funcDef, false)
			if err != nil {
				return err
			}
		}
	}

	// append any new views
	for _, viewDef := range s.Views {
		name := snake(viewDef.Name)
		if existing[name] == "" {
			err = s.GenerateRustFunc(file, viewDef, true)
			if err != nil {
				return err
			}
		}
	}

	return os.Remove(scOriginal)
}

func (s *Schema) GenerateRustFuncScanner(file *os.File) ([]string, StringMap, error) {
	defer file.Close()
	existing := make(StringMap)
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := rustFuncRegexp.FindStringSubmatch(line)
		if matches != nil {
			existing[matches[1]] = line
		}
		lines = append(lines, line)
	}
	err := scanner.Err()
	if err != nil {
		return nil, nil, err
	}
	return lines, existing, nil
}

func (s *Schema) GenerateRustFuncsNew(scFileName string) error {
	file, err := os.Create(scFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(false))
	fmt.Fprintf(file, "use wasmlib::*;\n\n")

	fmt.Fprintf(file, "use schema::*;\n")
	if len(s.Types) != 0 {
		fmt.Fprintf(file, "use types::*;\n")
	}

	for _, funcDef := range s.Funcs {
		err = s.GenerateRustFunc(file, funcDef, false)
		if err != nil {
			return err
		}
	}
	for _, viewDef := range s.Views {
		err = s.GenerateRustFunc(file, viewDef, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Schema) GenerateRustLib() error {
	file, err := os.Create("lib.rs")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintf(file, "use %s::*;\n", s.Name)
	fmt.Fprintf(file, "use schema::*;\n")
	fmt.Fprintf(file, "use wasmlib::*;\n\n")

	fmt.Fprintf(file, "mod %s;\n", s.Name)
	fmt.Fprintf(file, "mod schema;\n")
	if len(s.Types) != 0 {
		fmt.Fprintf(file, "mod types;\n")
	}

	fmt.Fprintf(file, "\n#[no_mangle]\n")
	fmt.Fprintf(file, "fn on_load() {\n")
	fmt.Fprintf(file, "    let exports = ScExports::new();\n")
	for _, funcDef := range s.Funcs {
		name := snake(funcDef.Name)
		fmt.Fprintf(file, "    exports.add_call(FUNC_%s, func_%s);\n", upper(name), name)
	}
	for _, viewDef := range s.Views {
		name := snake(viewDef.Name)
		fmt.Fprintf(file, "    exports.add_view(VIEW_%s, view_%s);\n", upper(name), name)
	}
	fmt.Fprintf(file, "}\n")

	//TODO generate parameter checks

	return nil
}

func (s *Schema) GenerateRustSchema() error {
	file, err := os.Create("schema.rs")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintf(file, "#![allow(dead_code)]\n\n")
	fmt.Fprintf(file, "use wasmlib::*;\n\n")

	fmt.Fprintf(file, "pub const SC_NAME: &str = \"%s\";\n", s.Name)
	if s.Description != "" {
		fmt.Fprintf(file, "pub const SC_DESCRIPTION: &str =  \"%s\";\n", s.Description)
	}
	hName := coretypes.Hn(s.Name)
	fmt.Fprintf(file, "pub const SC_HNAME: ScHname = ScHname(0x%s);\n", hName.String())

	if len(s.Params) != 0 {
		fmt.Fprintln(file)
		for _, name := range sortedFields(s.Params) {
			param := s.Params[name]
			name = upper(snake(name))
			fmt.Fprintf(file, "pub const PARAM_%s: &str = \"%s\";\n", name, param.Alias)
		}
	}

	if len(s.Vars) != 0 {
		fmt.Fprintln(file)
		for _, field := range s.Vars {
			name := upper(snake(field.Name))
			fmt.Fprintf(file, "pub const VAR_%s: &str = \"%s\";\n", name, field.Alias)
		}
	}

	if len(s.Funcs)+len(s.Views) != 0 {
		fmt.Fprintln(file)
		for _, funcDef := range s.Funcs {
			name := upper(snake(funcDef.Name))
			fmt.Fprintf(file, "pub const FUNC_%s: &str = \"%s\";\n", name, funcDef.Name)
		}
		for _, viewDef := range s.Views {
			name := upper(snake(viewDef.Name))
			fmt.Fprintf(file, "pub const VIEW_%s: &str = \"%s\";\n", name, viewDef.Name)
		}

		fmt.Fprintln(file)
		for _, funcDef := range s.Funcs {
			name := upper(snake(funcDef.Name))
			hName = coretypes.Hn(funcDef.Name)
			fmt.Fprintf(file, "pub const HFUNC_%s: ScHname = ScHname(0x%s);\n", name, hName.String())
		}
		for _, viewDef := range s.Views {
			name := upper(snake(viewDef.Name))
			hName = coretypes.Hn(viewDef.Name)
			fmt.Fprintf(file, "pub const HVIEW_%s: ScHname = ScHname(0x%s);\n", name, hName.String())
		}
	}
	return nil
}

func (s *Schema) GenerateRustTypes() error {
	if len(s.Types) == 0 {
		return nil
	}

	file, err := os.Create("types.rs")
	if err != nil {
		return err
	}
	defer file.Close()

	// write file header
	fmt.Fprintln(file, copyright(true))
	fmt.Fprintf(file, "#![allow(dead_code)]\n\n")
	fmt.Fprintf(file, "use wasmlib::*;\n")

	// write structs
	for _, typeDef := range s.Types {
		fmt.Fprintf(file, "\n//@formatter:off\n")
		fmt.Fprintf(file, "pub struct %s {\n", typeDef.Name)
		nameLen := 0
		typeLen := 0
		for _, field := range typeDef.Fields {
			fldName := snake(field.Name)
			if nameLen < len(fldName) { nameLen = len(fldName) }
			rustType := rustTypes[field.Type]
			if typeLen < len(rustType) { typeLen = len(rustType) }
		}
		for _, field := range typeDef.Fields {
			fldName := pad(snake(field.Name) + ":", nameLen+1)
			rfldType := pad(rustTypes[field.Type] + ",", typeLen+1)
			fmt.Fprintf(file, "    pub %s %s%s\n", fldName, rfldType, field.Comment)
		}
		fmt.Fprintf(file, "}\n")
		fmt.Fprintf(file, "//@formatter:on\n")
	}

	// write encoder and decoder for structs
	for _, typeDef := range s.Types {
		funcName := lower(snake(typeDef.Name))
		fmt.Fprintf(file, "\npub fn encode_%s(o: &%s) -> Vec<u8> {\n", funcName, typeDef.Name)
		fmt.Fprintf(file, "    let mut encode = BytesEncoder::new();\n")
		for _, field := range typeDef.Fields {
			name := snake(field.Name)
			ref := "&"
			if field.Type == "Int" || field.Type == "Hname" {
				ref = ""
			}
			fmt.Fprintf(file, "    encode.%s(%so.%s);\n", snake(field.Type), ref, name)
		}
		fmt.Fprintf(file, "    return encode.data();\n}\n")

		fmt.Fprintf(file, "\npub fn decode_%s(bytes: &[u8]) -> %s {\n", funcName, typeDef.Name)
		fmt.Fprintf(file, "    let mut decode = BytesDecoder::new(bytes);\n    %s {\n", typeDef.Name)
		for _, field := range typeDef.Fields {
			name := snake(field.Name)
			fmt.Fprintf(file, "        %s: decode.%s(),\n", name, snake(field.Type))
		}
		fmt.Fprintf(file, "    }\n}\n")
	}

	return nil
}
