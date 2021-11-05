package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	//nolint:unused
	snakePart       = regexp.MustCompile(`_[a-z]`)
	camelPart       = regexp.MustCompile(`[a-z0-9][A-Z]`)
	camelPartWithID = regexp.MustCompile(`[A-Z][A-Z]+[a-z]`)
)

func calculatePadding(fields []*Field, types StringMap, snakeName bool) (nameLen, typeLen int) {
	for _, param := range fields {
		fldName := param.Name
		if snakeName {
			fldName = snake(fldName)
		}
		if nameLen < len(fldName) {
			nameLen = len(fldName)
		}
		fldType := param.Type
		if types != nil {
			fldType = types[fldType]
		}
		if typeLen < len(fldType) {
			typeLen = len(fldType)
		}
	}
	return
}

// convert lowercase snake case to camel case
//nolint:deadcode,unused
func camel(name string) string {
	// replace each underscore followed by [a-z] with [A-Z]
	return snakePart.ReplaceAllStringFunc(name, func(sub string) string {
		return strings.ToUpper(sub[1:])
	})
}

// capitalize first letter
func capitalize(name string) string {
	return upper(name[:1]) + name[1:]
}

// convert to lower case
func lower(name string) string {
	return strings.ToLower(name)
}

func FindModulePath() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	// we're going to walk up the path, make sure to restore
	ModuleCwd = cwd
	defer func() {
		_ = os.Chdir(ModuleCwd)
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
			ModuleName = strings.TrimSpace(line[len("module"):])
			ModulePath = cwd
			return nil
		}
	}

	return fmt.Errorf("cannot find module definition in go.mod")
}

// pad to specified size with spaces
func pad(name string, size int) string {
	for i := len(name); i < size; i++ {
		name += " "
	}
	return name
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
	return lower(name[:1]) + name[1:]
}

// convert to upper case
func upper(name string) string {
	return strings.ToUpper(name)
}

func sortedFields(dict FieldMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedFuncDescs(dict FuncDefMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(dict StringMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedMaps(dict StringMapMap) []string {
	keys := make([]string, 0)
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
