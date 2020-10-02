go test -buildmode=exe -run TestWasmVMSend5Requests1Sec
pause
go test -buildmode=exe -run TestWasmSend1ReqIncSimple
pause
go test -buildmode=exe -run TestWasmSend1ReqIncRepeatSuccessTimelock
pause
go test -buildmode=exe -run TestWasmChainIncTimelock
pause
go test -buildmode=exe -run TestWasmSend1Bet
pause
go test -buildmode=exe -run TestWasmSend5Bets
pause
go test -buildmode=exe -run TestWasmSendBetsAndPlay
