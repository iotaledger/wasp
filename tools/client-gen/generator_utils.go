package main

import (
	"fmt"
	"reflect"
	"strings"
)

func dereferenceType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

func cleanName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func cleanGenericTypeName(t reflect.Type) string {
	typeName := t.String()
	if idx := strings.Index(typeName, "["); idx == -1 {
		return t.Name()
	}
	return formatGenericType(typeName)
}

func formatGenericType(typeName string) string {
	idx := strings.Index(typeName, "[")
	baseName := typeName[:idx]
	params := typeName[idx+1 : len(typeName)-1]

	cleanParams := cleanTypeParameters(params)
	cleanBaseName := getBaseTypeName(baseName)

	return fmt.Sprintf("%s_%s", cleanBaseName, strings.Join(cleanParams, "_"))
}

func cleanTypeParameters(params string) []string {
	paramTypes := strings.Split(params, ",")
	cleanParams := make([]string, 0, len(paramTypes))

	for _, param := range paramTypes {
		param = strings.TrimSpace(param)
		if lastDot := strings.LastIndex(param, "."); lastDot != -1 {
			param = param[lastDot+1:]
		}
		cleanParams = append(cleanParams, cleanName(param))
	}

	return cleanParams
}

func getBaseTypeName(name string) string {
	if lastDot := strings.LastIndex(name, "."); lastDot != -1 {
		return cleanName(name[lastDot+1:])
	}
	return cleanName(name)
}

func getQualifiedTypeName(t reflect.Type) string {
	if t.PkgPath() != "" {
		pkgName := t.PkgPath()[strings.LastIndex(t.PkgPath(), "/")+1:]
		typeName := cleanGenericTypeName(t)
		return fmt.Sprintf("%s%s", strings.Title(cleanName(pkgName)), typeName)
	}
	return cleanGenericTypeName(t)
}
