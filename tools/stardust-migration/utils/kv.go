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
func SplitMapKey[T ~string | ~[]byte](key T, prefixToRemove ...string) (mapName, elemKey T) {
	if len(prefixToRemove) > 0 {
		key = MustRemovePrefix(key, prefixToRemove[0])
	}

	const elemSep = "."
	sepPos := strings.Index(string(key), elemSep)

	sepFound := sepPos >= 0
	sepIsNotFirstOrLastChar := sepPos != 0 && sepPos < len(key)-1
	isMapElement := sepFound && sepIsNotFirstOrLastChar

	if !isMapElement {
		// Not a map element - maybe map itself or just something else
		return T([]byte(nil)), T([]byte(nil))
	}

	if n := strings.Index(string(key[sepPos+1:]), elemSep); n != -1 {
		// ASCII code of separator could be part of bytes of map name or element key...
		panic(fmt.Sprintf("multiple map elem separators found: key = %x / %v", key, key))
	}

	return key[:sepPos], key[sepPos+1:]
}

func GetMapElemPrefixes[T ~string | ~[]byte](key T) []T {
	prefixes := make([]T, 0, 1)

	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			prefixes = append(prefixes, key[:i+1])
		}
	}

	return prefixes
}

// Split map key into map name and element key
func SplitArrayKey[T ~string | ~[]byte](key T, prefixToRemove ...string) (arrayName, elemIndex T) {
	if len(prefixToRemove) > 0 {
		key = MustRemovePrefix(key, prefixToRemove[0])
	}

	const indexSep = "#"
	sepPos := strings.Index(string(key), indexSep)

	sepFound := sepPos >= 0
	sepIsNotFirstOrLastChar := sepPos != 0 && sepPos < len(key)-1
	isArrayElement := sepFound && sepIsNotFirstOrLastChar

	if !isArrayElement {
		// Not an array element - maybe array itself or just something else
		return T([]byte(nil)), T([]byte(nil))
	}

	if n := strings.Index(string(key[sepPos+1:]), indexSep); n != -1 {
		// ASCII code of separator could be part of bytes of array name or index...
		panic(fmt.Sprintf("multiple array elem separators found: key = %x / %v", key, key))
	}

	return key[:sepPos], key[sepPos+1:]
}
