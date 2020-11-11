go test -buildmode=exe -run TestDeployChain %1
pause
go test -buildmode=exe -run TestDeployContractOnly %1
pause
go test -buildmode=exe -run TestDeployContractAndSpawn %1
pause
go test -buildmode=exe -run TestDeployExternalContractOnly %1


