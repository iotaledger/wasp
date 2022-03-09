package util

import (
	"os"
)

// TODO get rid of drand dependency

const defaultRelativePath = "wasm"

func LocateFile(fileName string, relativePath ...string) string {
	relPath := defaultRelativePath
	if len(relativePath) > 0 {
		relPath = relativePath[0]
	}
	// check if this file exists
	exists, err := ExistsFilePath(fileName)
	if err != nil {
		panic(err)
	}
	if exists {
		return fileName
	}

	// walk up the directory tree to find the Wasm repo folder
	path := relPath
	_, err = ExistsFilePath(path)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		path = "../" + path
		exists, err = ExistsFilePath(path)
		if err != nil {
			panic(err)
		}
		if exists {
			break
		}
	}

	// check if file is in Wasm repo
	path = path + "/" + fileName
	exists, err = ExistsFilePath(path)
	if err != nil {
		panic(err)
	}
	if !exists {
		panic("Missing wasm file: " + fileName)
	}
	return path
}

// ExistsFilePath returns whether the given file or directory exists.
func ExistsFilePath(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
