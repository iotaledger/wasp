package util

import (
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/hive.go/core/ioutils"
)

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

func CreateDirectoryForFilePath(filePath string, perm os.FileMode) error {
	if filePath == "" {
		// do not create a folder if no path is given
		return nil
	}

	dir := path.Dir(filePath)

	if err := ioutils.CreateDirectory(dir, perm); err != nil {
		return fmt.Errorf("unable to create directory \"%s\": %w", dir, err)
	}

	return nil
}
