package utils

import (
	"fmt"
	"strings"
)

func MustRemovePrefix[T ~string | ~[]byte](v T, prefix string) T {
	r, found := strings.CutPrefix(string(v), prefix)
	if !found {
		panic(fmt.Sprintf("Prefix '%v' not found: %v", prefix, v))
	}

	return T(r)
}

// Split map key into map name and element key
func SplitMapKey[T ~string | ~[]byte](storeKey T, prefixToRemove ...string) (mapName, elemKey T) {
	if len(prefixToRemove) > 0 {
		storeKey = MustRemovePrefix(storeKey, prefixToRemove[0])
	}

	const elemSep = "."
	pos := strings.Index(string(storeKey), elemSep)

	sepFound := pos >= 0
	sepIsNotLastChar := pos < len(storeKey)-1
	isMapElement := sepFound && sepIsNotLastChar

	if isMapElement {
		return storeKey[:pos], storeKey[pos+1:]
	}

	// Not a map element - maybe map itself or just something else
	return storeKey, T("")
}

func GetMapElemPrefix[T ~string | ~[]byte](key T) (T, bool) {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return key[:i+1], true
		}
	}

	return key, false
}
