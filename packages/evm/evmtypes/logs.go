// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
)

func LogMatches(log *types.Log, addresses []common.Address, topics [][]common.Hash) bool {
	return logMatchesAddresses(log, addresses) && logMatchesAllEvents(log, topics)
}

func logMatchesAddresses(log *types.Log, addresses []common.Address) bool {
	if len(addresses) == 0 {
		return true
	}
	return lo.Contains(addresses, log.Address)
}

func logMatchesAllEvents(log *types.Log, events [][]common.Hash) bool {
	if len(events) == 0 {
		return true
	}
	if len(events) > len(log.Topics) {
		return false
	}
	for i, topics := range events {
		if len(topics) > 0 && !lo.Contains(topics, log.Topics[i]) {
			return false
		}
	}
	return true
}
