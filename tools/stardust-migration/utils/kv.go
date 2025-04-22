package utils

import (
	"fmt"
	"strings"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_buffered "github.com/nnikolash/wasp-types-exported/packages/kv/buffered"
	old_kvdict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	"github.com/samber/lo"
)

func MustRemovePrefix[T, Prefix ~string | ~[]byte](v T, prefix Prefix) T {
	r, found := strings.CutPrefix(string(v), string(prefix))
	if !found {
		panic(fmt.Sprintf("Prefix '%v' not found: %v", prefix, v))
	}

	return T(r)
}

// Finds all map key separators and tries to parse them.
// If nothing parses - considered as not a map lement.
// If multiple pairs are parseable - panics.
func MustSplitParseMapKeyAny[T ~string | ~[]byte, P ~string | ~[]byte, MapName any, ElemKey any](key T, prefixToRemove P, f func(mapName, elemKey T) (MapName, ElemKey, error)) (T, MapName, T, ElemKey, bool) {
	if string(prefixToRemove) != "" {
		key = MustRemovePrefix(key, string(prefixToRemove))
	}

	var mapNameBytes, elemKeyBytes T
	var mapName MapName
	var elemKey ElemKey
	foundSep := false
	parsed := false
	errs := make([]string, 0)

	for i := 1; i < len(key)-1; i++ {
		if key[i] != '.' {
			continue
		}

		foundSep = true

		mb, eb := key[:i], key[i+1:]
		m, e, err := f(mb, eb)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if parsed {
			panic(fmt.Sprintf("Multiple parseable map key separators found: %x / %v", key, string(key)))
		}

		mapNameBytes, elemKeyBytes = mb, eb
		mapName, elemKey = m, e
		parsed = true
		// NOTE: We try parse ALL pairs to see if only one is parseable
	}

	if !foundSep || !parsed {
		return key, lo.Empty[MapName](), T([]byte(nil)), lo.Empty[ElemKey](), false
	}

	return mapNameBytes, mapName, elemKeyBytes, elemKey, true
}

// Split map key into map name and element key
// expectedSepPos - if 0, then separator is searched for. If > 0, then it is expected to be at this position.
// If < 0, then it is expected to be at len(key) + expectedSepPos. If no separator found where expected, then
// empty map name and element key are returned.
func MustSplitMapKey[T ~string | ~[]byte](key T, expectedSepPos int, prefixToRemove ...string) (mapName, elemKey T) {
	return lo.Must2(SplitMapKey(key, expectedSepPos, prefixToRemove...))
}

func MustSplitArrayKey[T ~string | ~[]byte](key T, expectedSepPos int, prefixToRemove ...string) (arrayName, elemIndex T) {
	return lo.Must2(SplitArrayKey(key, expectedSepPos, prefixToRemove...))
}

func SplitMapKey[T ~string | ~[]byte](key T, expectedSepPos int, prefixToRemove ...string) (mapName, elemKey T, _ error) {
	return splitContainerEntryKey(key, '.', expectedSepPos, prefixToRemove...)
}

func SplitArrayKey[T ~string | ~[]byte](key T, expectedSepPos int, prefixToRemove ...string) (arrayName, elemIndex T, _ error) {
	return splitContainerEntryKey(key, '#', expectedSepPos, prefixToRemove...)
}

func splitContainerEntryKey[T ~string | ~[]byte](key T, elemSep byte, expectedSepPos int, prefixToRemove ...string) (contName, elemKey T, _ error) {
	if len(prefixToRemove) > 0 {
		key = MustRemovePrefix(key, prefixToRemove[0])
	}

	if expectedSepPos != 0 {
		if expectedSepPos < 0 {
			if len(key)+expectedSepPos < 0 {
				// There is not enough bytes to have such a long suffix - its a container itself
				return key, T([]byte(nil)), nil
			}

			expectedSepPos = len(key) + expectedSepPos
		}
		if expectedSepPos == len(key) {
			// Separator would be right after the end of key bytes - its a container itself
			return key, T([]byte(nil)), nil
		}
		if len(key) < expectedSepPos {
			// Expected separator position is far beyond the end of key
			return T([]byte(nil)), T([]byte(nil)), fmt.Errorf("key is too small: %v < %v: key = %x / %v",
				len(key), expectedSepPos, key, string(key))
		}
		if key[expectedSepPos] != elemSep {
			// Separator not found at expected position
			return T([]byte(nil)), T([]byte(nil)), fmt.Errorf("unexpected key format: %x / %v", key, string(key))
		}

		return key[:expectedSepPos], key[expectedSepPos+1:], nil
	} else {
		sepPos := strings.IndexByte(string(key), elemSep)
		if sepPos == -1 {
			// No spearator found - its a container itself
			return key, T([]byte(nil)), nil
		}
		if sepPos == 0 || sepPos == len(key)-1 {
			// Separator is first or last character
			return T([]byte(nil)), T([]byte(nil)), fmt.Errorf("unexpected key format: %x / %v", key, string(key))
		}
		if n := strings.IndexByte(string(key[sepPos+1:]), elemSep); n != -1 {
			// Another separator found. It is ambiguous which one is the correct one.
			// ASCII code of separator could be part of bytes of container name or element key...
			return T([]byte(nil)), T([]byte(nil)), fmt.Errorf("multiple container elem separators found: key = %x / %v", key, key)
		}

		return key[:sepPos], key[sepPos+1:], nil
	}
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

// Returns KVStoreReader, which will iterate by both Sets and Dels of mutations. For Dels, value will be nil.
func DictKvFromMuts(muts *old_buffered.Mutations) old_kv.KVStoreReader {
	d := old_kvdict.New()
	for k, v := range muts.Sets {
		d[k] = v
	}
	for k := range muts.Dels {
		d[k] = nil
	}

	return d
}
