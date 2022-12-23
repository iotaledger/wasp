// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

type Equatable interface {
	Equals(Equatable) bool
}

func Contains[E Equatable](object E, objects []E) bool {
	for _, bh := range objects {
		if bh.Equals(object) {
			return true
		}
	}
	return false
}

func Remove[E Equatable](object E, objects []E) []E {
	for i := range objects {
		if objects[i].Equals(object) {
			objects[i] = objects[len(objects)-1]
			return objects[:len(objects)-1]
		}
	}
	return objects
}

func RemoveAll[E Equatable](objectsToRemove, objects []E) []E {
	result := objects
	for i := range objectsToRemove {
		result = Remove(objectsToRemove[i], result)
	}
	return result
}

func AllDifferent[E Equatable](objects []E) bool {
	for i := range objects {
		for j := 0; j < i; j++ {
			if objects[i].Equals(objects[j]) {
				return false
			}
		}
	}
	return true
}
