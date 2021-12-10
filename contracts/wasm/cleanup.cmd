del /s consts.*
del /s contract.*
del /s keys.*
del /s lib.*
del /s params.*
del /s results.*
del /s state.*
del /s typedefs.*
del /s types.*
del /s main.go
del /s index.ts
del /s tsconfig.json
del /s /q *.wasm
for /d %%f in (*.) do del /s /q %%f\go\pkg\*.*
for /d %%f in (*.) do del /s /q %%f\pkg\*.*
for /d %%f in (*.) do del /s /q %%f\ts\pkg\*.*
del /s /q target\*.*
