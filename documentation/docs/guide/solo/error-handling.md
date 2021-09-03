# Error handling

The following test posts a request to the `example1` smart contract without
the expected parameter `paramString`. The
statement `ctx.require(par.exists(), "string parameter not found");` makes
the smart contract panic if the condition is not satisfied.

```go
func TestTutorial4(t *testing.T) {
	env := solo.New(t, false, false, seed)

	chain := env.NewChain(nil, "ex4")
	// deploy the contract on chain
	err := chain.DeployWasmContract(nil, "example1", "example_tutorial_bg.wasm")
	require.NoError(t, err)

	// call contract incorrectly (omit 'paramString')
	req := solo.NewCallParams("example1", "storeString").WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.Error(t, err)
}
```

The fragments in the output of the test:

```log
37:34.189474700	PANIC	TestTutorial4.ex4	vmcontext/log.go:12	string parameter not found

37:34.192828900	INFO	TestTutorial4.ex4	solo/run.go:148	REQ: 'tx/[0]9r5zoeusdwTcWkDTEMYjeqNj8reiUsLiHF81vExPrvNW: 'panic in VM: string parameter not found''
``` 

It shows that the panic indeed occurred. The test passes because the resulting
error was expected.

The log record

```log
37:34.192828900	INFO	TestTutorial4.ex4	solo/run.go:148	REQ: 'tx/[0]9r5zoeusdwTcWkDTEMYjeqNj8reiUsLiHF81vExPrvNW: 'panic in VM: string parameter not found''
```

is a printed receipt of the request. It is stored on the chain for each request processed.

Note that this test ends with the state `#4`, despite the fact that the last
request to the smart contract failed. This is important: **whatever happens
during the execution of a smart contract's full entry point, processing of the 
request always results in the state transition**.

The VM context catches exceptions (panics) in the program. Its consequences are
recorded in the state of the chain during the fallback processing, no matter if
the panic was triggered by the logic of the smart contract or whether it was 
triggered by the sandbox run-time code.

In the case of `example1` the error event was recorded in the immutable record
log of the chain, aka `receipt`, but the data state of the smart contract wasn't modified. In
other cases the fallback actions may be more complex.
