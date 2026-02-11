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

## ObjectFilter

The `ObjectFilter` input type uses `@genqlient(for: "ObjectFilter.*", pointer: true)` directives in `queries/objects.graphql` to generate pointer fields. This ensures that unset fields serialize as `null` (treated as "not specified" by the GraphQL API) rather than zero values like `0x000...000`.
