package common

import (
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/hive.go/core/ioutils"
)

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
