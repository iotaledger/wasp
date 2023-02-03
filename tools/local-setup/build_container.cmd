rem export variables for the build process
set BUILD_TAGS=rocksdb
for /f %%f in ('git describe --tags') do set BUILD_LD_FLAGS=-X=github.com/iotaledger/wasp/core/app.Version=%%f

rem build the wasp container
docker compose build wasp