cd %1

# remove go files
del /s go/%1/consts.go
del /s go/%1/contract.go
del /s go/%1/keys.go
del /s go/%1/lib.go
del /s go/%1/params.go
del /s go/%1/results.go
del /s go/%1/state.go
del /s go/%1/typedefs.go
del /s go/%1/events.go
del /s go/%1/eventhandlers.go
del /s go/%1/structs.go
del /s go/%1/types.go
del /s go/main.go

# remove ts files
del /s ts/%1/consts.ts
del /s ts/%1/contract.ts
del /s ts/%1/keys.ts
del /s ts/%1/lib.ts
del /s ts/%1/params.ts
del /s ts/%1/results.ts
del /s ts/%1/state.ts
del /s ts/%1/typedefs.ts
del /s ts/%1/events.ts
del /s ts/%1/eventhandlers.ts
del /s ts/%1/structs.ts
del /s ts/%1/types.ts
del /s ts/%1/index.ts
del /s ts/%1/tsconfig.json
del /s /q pkg/*.*
del /s /q ts/pkg/*.*

# remove rs files
del /s src/consts.rs
del /s src/contract.rs
del /s src/keys.rs
del /s src/lib.rs
del /s src/params.rs
del /s src/results.rs
del /s src/state.rs
del /s src/typedefs.rs
del /s src/events.rs
del /s src/eventhandlers.rs
del /s src/structs.rs
del /s src/types.rs

cd scripts
