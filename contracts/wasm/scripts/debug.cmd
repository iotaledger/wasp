if exist ../../../packages/wasmvm/wasmvmhost grepl "\nwasmvmhost = \{ git" "\nwasmvmhost = { path = \q../../../../../packages/wasmvm/wasmvmhost\q }\n#wasmvmhost = { git" rs\main\cargo.toml
if exist ../../../../packages/wasmvm/wasmvmhost grepl "\nwasmvmhost = \{ git" "\nwasmvmhost = { path = \q../../../../../../packages/wasmvm/wasmvmhost\q }\n#wasmvmhost = { git" rs\main\cargo.toml
