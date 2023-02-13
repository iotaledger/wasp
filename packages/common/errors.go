package common

import (
	"github.com/pkg/errors"
)

// ErrOperationAborted is returned when the operation was aborted e.g. by a shutdown signal.
var ErrOperationAborted = errors.New("operation was aborted")
