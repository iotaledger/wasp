package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func generateContractFuncs() {
	fset := token.NewFileSet()

	output := "var contractFuncs []CoreContractFunction = []CoreContractFunction{\n"

	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == "interface.go" {
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing %s: %v", path, err)
			}

			// Get package name
			packageName := file.Name.Name

			ast.Inspect(file, func(n ast.Node) bool {
				if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.VAR {
					for _, spec := range decl.Specs {
						if valueSpec, ok := spec.(*ast.ValueSpec); ok {
							for _, name := range valueSpec.Names {
								if strings.HasPrefix(name.Name, "Func") ||
									strings.HasPrefix(name.Name, "View") {
									output += fmt.Sprintf("\t\tconstructCoreContractFunction(&%s.%s),\n",
										packageName, name.Name)
								}
							}
						}
					}
				}
				return true
			})
		}
		return nil
	})

	output += "\t}"

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}

	fmt.Println(output)
}
