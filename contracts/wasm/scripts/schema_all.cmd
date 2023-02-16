@echo off

set BUILD_TAGS=rocksdb
for /f %%f in ('git describe --tags') do set BUILD_LD_FLAGS=-X=github.com/iotaledger/wasp/core/app.Version=%%f
go install -ldflags %BUILD_LD_FLAGS% ../../tools/schema

cd ..
for /d %%f in (*.) do call scripts\schema_build.cmd %%f %1
cd gascalibration
for /d %%f in (*.) do call ..\scripts\schema_build.cmd %%f %1
cd ..\scripts
