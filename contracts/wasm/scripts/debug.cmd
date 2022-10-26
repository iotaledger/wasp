if exist ../../../packages/wasmvm/wasmlib       grepl -s "\nwasmlib = \{ git" "\nwasmlib = { path = \q../../../../../packages/wasmvm/wasmlib\q }\n#wasmlib = { git" cargo.toml
if exist ../../../../packages/wasmvm/wasmlib    grepl -s "\nwasmlib = \{ git" "\nwasmlib = { path = \q../../../../../../packages/wasmvm/wasmlib\q }\n#wasmlib = { git" cargo.toml
if exist ../../../packages/wasmvm/wasmvmhost    grepl -s "\nwasmvmhost = \{ git" "\nwasmvmhost = { path = \q../../../../../packages/wasmvm/wasmvmhost\q }\n#wasmvmhost = { git" cargo.toml
if exist ../../../../packages/wasmvm/wasmvmhost grepl -s "\nwasmvmhost = \{ git" "\nwasmvmhost = { path = \q../../../../../../packages/wasmvm/wasmvmhost\q }\n#wasmvmhost = { git" cargo.toml
