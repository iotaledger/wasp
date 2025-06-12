// Package testconfig provides utilities for test configuration management
package testconfig

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"fortio.org/safecast"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	testconfigFile = ".testconfig"
)

var (
	loadedConfigs     = make(map[string]sectionCfg)
	loadedConfigsLock sync.Mutex
)

type sectionCfg struct {
	C        *koanf.Koanf
	WasFound bool
}

func LoadConfig(sectionName string) (_ *koanf.Koanf, configFound bool) {
	loadedConfigsLock.Lock()
	defer loadedConfigsLock.Unlock()

	if loadedCfg, ok := loadedConfigs[sectionName]; ok {
		return loadedCfg.C, loadedCfg.WasFound
	}

	c := koanf.New(".")
	if err := c.Load(file.Provider(path.Join(GetRootDir(), testconfigFile)), json.Parser()); err != nil {
		fmt.Printf("config file %v not found - using default values\n", testconfigFile)
	} else {
		subKeys := c.Cut(sectionName)
		if len(subKeys.Keys()) == 0 {
			fmt.Printf("key %v not found in config %v - using default values\n", sectionName, testconfigFile)
		} else {
			c = subKeys
			configFound = true
		}
	}

	prefix := "TEST_"
	removePrefix := strings.ToUpper(sectionName) + "_"

	envProvider := env.Provider(prefix, ".", func(s string) string {
		s = strings.TrimPrefix(s, removePrefix)
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	})

	if err := c.Load(envProvider, nil); err != nil {
		fmt.Printf("failed to load env vars: %v\n", err)
	}

	loadedConfigs[sectionName] = sectionCfg{
		C:        c,
		WasFound: configFound,
	}

	return c, configFound
}

func Get[ValueType any](section string, keyName string, def ValueType) ValueType {
	c, configFound := LoadConfig(section)
	if !configFound {
		return def
	}

	var v interface{}

	switch interface{}(&def).(type) {
	case *string:
		v = c.String(keyName)
	case *int:
		v = c.Int(keyName)
	case *int32:
		var err error
		if v, err = safecast.Convert[int32](c.Int64(keyName)); err != nil {
			panic(fmt.Sprintf("integer overflow when converting to int32: %v", err))
		}
	case *int64:
		v = c.Int64(keyName)
	case *uint:
		var err error
		if v, err = safecast.Convert[uint](c.Int(keyName)); err != nil {
			panic(fmt.Sprintf("integer overflow when converting to uint: %v", err))
		}
	case *uint32:
		var err error
		if v, err = safecast.Convert[uint32](c.Int(keyName)); err != nil {
			panic(fmt.Sprintf("integer overflow when converting to uint32: %v", err))
		}
	case *uint64:
		var err error
		if v, err = safecast.Convert[uint64](c.Int64(keyName)); err != nil {
			panic(fmt.Sprintf("integer overflow when converting to uint64: %v", err))
		}
	case *float32:
		v = float32(c.Float64(keyName))
	case *float64:
		v = c.Float64(keyName)
	case *bool:
		v = c.Bool(keyName)
	case *any:
		v = c.Get(keyName)
	}

	return v.(ValueType)
}

func GetRootDir() string {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find go.mod file")
		}
		dir = parent
	}
}
