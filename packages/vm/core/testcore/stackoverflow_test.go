package testcore

// func TestSandboxStackOverflow(t *testing.T) {
// 	contract := coreutil.NewContract("test stack overflow")
// 	testFunc := contract.Func("overflow")
// 	env := solo.New(t).WithNativeContract(
// 		contract.Processor(
// 			func(ctx isc.Sandbox) isc.CallArguments { return nil },
// 			testFunc.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
// 				ctx.Call(testFunc.Message(nil), nil)
// 				return nil
// 			}),
// 		),
// 	)

// 	chain := env.NewChain()

// 	err := chain.DeployContract(nil, contract.Name, contract.ProgramHash)
// 	require.NoError(t, err)

// 	_, err = chain.PostRequestSync(solo.NewCallParams(testFunc.Message(nil)).WithGasBudget(math.MaxUint64), nil)
// 	require.Error(t, err)
// 	testmisc.RequireErrorToBe(t, err, vm.ErrGasBudgetExceeded)
// }
