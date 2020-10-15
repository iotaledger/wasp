go test -buildmode=exe -run TestDeploySC
pause
go test -buildmode=exe -run TestGetSCData
pause
go test -buildmode=exe -run TestSend5ReqInc0SecDeploy
pause
go test -buildmode=exe -run TestSend100ReqMulti
pause
go test -buildmode=exe -run TestDeployDWF
pause
go test -buildmode=exe -run TestDWFDonateNTimes
pause
go test -buildmode=exe -run TestDWFDonateWithdrawAuthorised
pause
go test -buildmode=exe -run TestDWFDonateWithdrawNotAuthorised
pause
go test -buildmode=exe -run Test2SC
pause
go test -buildmode=exe -run TestPlus2SC
pause
go test -buildmode=exe -run TestTRTest
pause
go test -buildmode=exe -run TestKillNode
