del /s consts.*
del /s contract.*
del /s keys.*
del /s lib.*
del /s params.*
del /s results.*
del /s state.*
del /s typedefs.*
del /s types.*
del /s /q target\*.*

rem careful not to delete all, this could fuck up gascalibration
for /d %%f in (*.) do del %%f\go\main.go

rem careful not to delete all, this could fuck up fairroulette frontend
for /d %%f in (*.) do del %%f\ts\%%f\index.ts
for /d %%f in (*.) do del %%f\ts\%%f\tsconfig.json

for /d %%f in (*.) do del /s /q %%f\pkg\*.*
for /d %%f in (*.) do del /s /q %%f\ts\pkg\*.*

cd gascalibration
for /d %%f in (*.) do del %%f\go\main.go
for /d %%f in (*.) do del %%f\ts\%%f\index.ts
for /d %%f in (*.) do del %%f\ts\%%f\tsconfig.json
for /d %%f in (*.) do del /s /q %%f\pkg\*.*
for /d %%f in (*.) do del /s /q %%f\ts\pkg\*.*
cd ..
