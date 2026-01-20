package graphqltypes

// TransactionResponseOptions specifies what data to include in transaction responses
type TransactionResponseOptions struct {
	// Whether to show transaction input data
	ShowInput bool
	// Whether to show transaction effects
	ShowEffects bool
	// Whether to show transaction events
	ShowEvents bool
	// Whether to show object changes
	ShowObjectChanges bool
	// Whether to show balance changes
	ShowBalanceChanges bool
	// Whether to show raw transaction input
	ShowRawInput bool
}

// ObjectDataOptions specifies what data to include in object responses
type ObjectDataOptions struct {
	// Whether to fetch the object type
	ShowType bool
	// Whether to fetch the object content
	ShowContent bool
	// Whether to fetch the object content in BCS bytes
	ShowBcs bool
	// Whether to fetch the object owner
	ShowOwner bool
	// Whether to fetch the previous transaction digest
	ShowPreviousTransaction bool
	// Whether to fetch the storage rebate
	ShowStorageRebate bool
	// Whether to fetch the display metadata
	ShowDisplay bool
}
