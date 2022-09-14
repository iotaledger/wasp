// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	camelPart       = regexp.MustCompile(`[a-z0-9][A-Z]`)
	camelPartWithID = regexp.MustCompile(`[A-Z][A-Z]+[a-z]`)
	moduleCwd       = "???"
	moduleName      = "???"
	modulePath      = "???"
)

// capitalize first letter
func capitalize(name string) string {
	if name == "" {
		return ""
	}
	return upper(name[:1]) + name[1:]
}

func filterIDorVM(value string) string {
	n := strings.Index(value, "Id")
	for n >= 0 {
		if n+2 == len(value) || unicode.IsUpper(rune(value[n+2])) {
			value = value[:n] + "ID" + value[n+2:]
		}
		n = strings.Index(value, "Id")
	}
	n = strings.Index(value, "Vm")
	for n >= 0 {
		if n+2 == len(value) || unicode.IsUpper(rune(value[n+2])) {
			value = value[:n] + "VM" + value[n+2:]
		}
		n = strings.Index(value, "Vm")
	}
	return value
}

func FindModulePath() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// we're going to walk up the path, make sure to restore
	moduleCwd = cwd
	defer func() {
		_ = os.Chdir(moduleCwd)
	}()

	file, err := os.Open("go.mod")
	for err != nil {
		err = os.Chdir("..")
		if err != nil {
			return fmt.Errorf("cannot find go.mod in cwd path")
		}
		prev := cwd
		cwd, err = os.Getwd()
		if err != nil {
			return err
		}
		if cwd == prev {
			// e.g. Chdir("..") gets us in a loop at Linux root
			return fmt.Errorf("cannot find go.mod in cwd path")
		}
		file, err = os.Open("go.mod")
	}

	// now file is the go.mod and cwd holds the path
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(line[len("module"):])
			modulePath = cwd
			return nil
		}
	}

	return fmt.Errorf("cannot find module definition in go.mod")
}

// convert to lower case
func lower(name string) string {
	return strings.ToLower(name)
}

// convert camel case to lower case snake case
func snake(name string) string {
	// insert underscores between [a-z0-9] followed by [A-Z]
	name = camelPart.ReplaceAllStringFunc(name, func(sub string) string {
		return sub[:1] + "_" + sub[1:]
	})

	// insert underscores between double [A-Z] followed by [a-z]
	name = camelPartWithID.ReplaceAllStringFunc(name, func(sub string) string {
		n := len(sub)
		return sub[:n-2] + "_" + sub[n-2:]
	})

	// lowercase the entire final result
	return lower(name)
}

// uncapitalize first letter
func uncapitalize(name string) string {
	if name == "" {
		return ""
	}
	return lower(name[:1]) + name[1:]
}

// convert to upper case
func upper(name string) string {
	return strings.ToUpper(name)
}
