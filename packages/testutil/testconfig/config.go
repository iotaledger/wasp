// Package testconfig provides utilities for test configuration management
package testconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const (
	testconfigFile = ".testconfig"
)

var (
	loadedConfigs     = make(map[string]sectionCfg)
	loadedConfigsLock sync.Mutex
)

type sectionCfg struct {
	C        *viper.Viper
	WasFound bool
}

func LoadConfig(sectionName string) (_ *viper.Viper, configFound bool) {
	loadedConfigsLock.Lock()
	defer loadedConfigsLock.Unlock()

	if loadedCfg, ok := loadedConfigs[sectionName]; ok {
		return loadedCfg.C, loadedCfg.WasFound
	}

	c := viper.NewWithOptions(viper.EnvKeyReplacer(&envKeyReplacer{
		RemovePrefix: strings.ToUpper(sectionName) + ".",
		AddPrefix:    "TEST_",
	}))

	c.SetConfigName(testconfigFile)
	c.SetConfigType("json")
	c.AddConfigPath(GetRootDir())

	if err := c.ReadInConfig(); err != nil {
		fmt.Printf("config file %v not found - using defaul values\n", testconfigFile)
	} else {
		if c = c.Sub(sectionName); c == nil {
			fmt.Printf("key %v not found in config %v - using defaul values\n", sectionName, testconfigFile)
			c = viper.New()
		} else {
			configFound = true
		}
	}

	c.AutomaticEnv()

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
		v = c.GetString(keyName)
	case *int:
		v = c.GetInt(keyName)
	case *int32:
		v = c.GetInt32(keyName)
	case *int64:
		v = c.GetInt64(keyName)
	case *uint:
		v = c.GetUint(keyName)
	case *uint32:
		v = c.GetUint32(keyName)
	case *uint64:
		v = c.GetUint64(keyName)
	case *float32:
		v = float32(c.GetFloat64(keyName))
	case *float64:
		v = c.GetFloat64(keyName)
	case *bool:
		v = c.GetBool(keyName)
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

type envKeyReplacer struct {
	RemovePrefix string
	AddPrefix    string
}

func (r *envKeyReplacer) Replace(s string) string {
	return r.AddPrefix + strings.TrimPrefix(s, r.RemovePrefix)
}
