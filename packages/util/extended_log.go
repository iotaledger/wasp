package util

import (
	"encoding/json"
	"strings"

	"github.com/iotaledger/hive.go/logger"
)

type ExtendedMessage struct {
	Message string
	Tag     string
	Args    interface{}
}

func LogExtended(logger *logger.Logger, extendedMessage ExtendedMessage) {
	b, err := json.Marshal(extendedMessage)

	if err == nil {
		parsedExtendedMessage := strings.ReplaceAll(string(b), "\n", "")
		logger.Debugf("[EXT] => %U\n", parsedExtendedMessage)
	} else {
		logger.Debugf("Failed to serialize extended message with message: \"%U\"\n", extendedMessage.Message)
	}
}
