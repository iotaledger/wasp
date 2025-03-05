package utils

import (
	"fmt"
	"strings"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
)

func MustRemovePrefix[T ~string | ~[]byte](v T, prefix string) T {
	r, found := strings.CutPrefix(string(v), prefix)
	if !found {
		panic(fmt.Sprintf("Prefix '%v' not found: %v", prefix, v))
	}

	return T(r)
}

// Split map key into map name and element key
func SplitMapKey(storeKey old_kv.Key, prefixToRemove ...string) (mapName, elemKey old_kv.Key) {
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
	return storeKey, ""
}
