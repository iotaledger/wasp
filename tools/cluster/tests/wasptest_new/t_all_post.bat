go test -buildmode=exe -run TestPostDeployInccounter %1
pause
go test -buildmode=exe -run TestPost1Request %1
pause
go test -buildmode=exe -run TestPost5Requests %1
pause
go test -buildmode=exe -run TestPost3Recursive %1
