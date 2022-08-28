#!/bin/bash
example_name=$1
cd $example_name

# remove go files
rm ./go/$example_name/consts.go
rm ./go/$example_name/contract.go
rm ./go/$example_name/keys.go
rm ./go/$example_name/lib.go
rm ./go/$example_name/params.go
rm ./go/$example_name/results.go
rm ./go/$example_name/state.go
rm ./go/$example_name/typedefs.go
rm ./go/$example_name/events.go
rm ./go/$example_name/eventhandlers.go
rm ./go/$example_name/structs.go
rm ./go/$example_name/types.go
rm ./go/main.go

# remove ts files
rm ./ts/$example_name/consts.ts
rm ./ts/$example_name/contract.ts
rm ./ts/$example_name/keys.ts
rm ./ts/$example_name/lib.ts
rm ./ts/$example_name/params.ts
rm ./ts/$example_name/results.ts
rm ./ts/$example_name/state.ts
rm ./ts/$example_name/typedefs.ts
rm ./ts/$example_name/events.ts
rm ./ts/$example_name/eventhandlers.ts
rm ./ts/$example_name/structs.ts
rm ./ts/$example_name/types.ts
rm ./ts/$example_name/index.ts
rm ./ts/$example_name/tsconfig.json
rm ./pkg/*.*
rm ./ts/pkg/*.*

# remove rs files
rm ./src/consts.rs
rm ./src/contract.rs
rm ./src/keys.rs
rm ./src/lib.rs
rm ./src/params.rs
rm ./src/results.rs
rm ./src/state.rs
rm ./src/typedefs.rs
rm ./src/events.rs
rm ./src/eventhandlers.rs
rm ./src/structs.rs
rm ./src/types.rs
