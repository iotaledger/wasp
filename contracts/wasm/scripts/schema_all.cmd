@echo off
cd ..
go install ../../tools/schema
for /d %%f in (*.) do call scripts\schema_build.cmd %%f %1
cd gascalibration
for /d %%f in (*.) do call ..\scripts\schema_build.cmd %%f %1
cd ..\scripts
