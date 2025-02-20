package migrations

import (
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_errors "github.com/nnikolash/wasp-types-exported/packages/vm/core/errors"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

func MigrateErrorsContract(oldChainState old_kv.KVStoreReader, newChainState kv.KVStore) {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_errors.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, errors.Contract.Hname())

	migrateErrorTemplates(oldContractState, newContractState)
}

func migrateErrorTemplates(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	// TODO: Iterating by prefix is unsafe. Is there another way in this case?

	oldState.Iterate(old_kv.Key(old_errors.PrefixErrorTemplateMap), func(oldKey old_kv.Key, oldVal []byte) bool {
		// When we iterate by prefix, we find both map itself and its elements.
		oldContractIDBytes, oldErrorIDBytes := SplitMapKey(oldKey, old_errors.PrefixErrorTemplateMap)
		if oldErrorIDBytes == "" {
			// Not a map element
			return true
		}

		oldContractID := lo.Must(old_isc.HnameFromBytes([]byte(oldContractIDBytes)))
		oldErrorTemplate := lo.Must(old_isc.VMErrorTemplateFromBytes(oldVal))

		newContractID := OldHnameToNewHname(oldContractID)

		// TODO: Will this also correctly migrate error code?
		w := errors.NewStateWriter(newState)
		w.ErrorCollection(newContractID).Register(oldErrorTemplate.MessageFormat())

		return true
	})
}
