go test -buildmode=exe -run TestBasicAccounts %1
pause
go test -buildmode=exe -run TestBasic2Accounts %1
