# IOTA GraphQL Client

This directory contains the GraphQL client implementation for IOTA, using [genqlient](https://github.com/Khan/genqlient) to generate type-safe GraphQL queries.

## Code Generation

The GraphQL code is generated from the schema using genqlient:

```bash
cd wasp/clients/iotagraphql
genqlient
```

This will regenerate `generated.go` based on:
- `schema.graphql` - The GraphQL schema
- `queries/*.graphql` - The GraphQL queries
- `genqlient.yaml` - The genqlient configuration

## Custom Code

### objectfilter_custom.go

This file contains a custom `MarshalJSON` method for the `ObjectFilter` type. This is necessary because:

1. genqlient generates non-pointer fields for optional GraphQL input fields
2. Non-pointer fields always have a value (even if it's the zero value)
3. The zero-value for `iotago.Address` (all zeros: `0x00...00`) was being sent to the GraphQL API
4. This zero-value address was incorrectly filtering out all results when using type filters

The custom marshaler ensures that zero-value fields are omitted from the JSON payload, allowing filters to work correctly.

**Important**: This file should NOT be deleted or modified when regenerating code with genqlient. It's a permanent workaround for a genqlient limitation.
