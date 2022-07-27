@echo off
go install ../../tools/schema
for /d %%f in (*.) do call schema_build.cmd %%f %1
cd gascalibration
for /d %%f in (*.) do call ..\schema_build.cmd %%f %1
cd ..
