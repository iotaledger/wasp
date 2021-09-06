# First Example 

The following is an example of a _Solo_ test. It deploys a new chain and invokes
a function in the `root` contract.

```go
var seed = ed25519.NewSeed([]byte("long long seed for determinism................"))

func TestTutorial1(t *testing.T) {
	env := solo.New(t, false, false, seed)
	chain := env.NewChain(nil, "ex1")

	// calls view root::GetChainInfo
	chainID, chainOwnerID, coreContracts := chain.GetInfo()
	// assert all core contracts deployed (default)
	require.EqualValues(t, len(core.AllCoreContractsByHash), len(coreContracts))

	t.Logf("chain ID: %s", chainID.String())
	t.Logf("chain owner ID: %s", chainOwnerID.String())
	for hname, rec := range coreContracts {
		t.Logf("    Core contract '%s': %s", rec.Name, iscp.NewAgentID(chainID.AsAddress(), hname))
	}
}
```

The output of the test will be something like this:

```log
=== RUN   TestTutorial1
26:50.073047100	INFO	TestTutorial1.db	dbmanager/dbmanager.go:56	creating new registry database. Persistent: false
26:50.074651000	INFO	TestTutorial1	solo/solo.go:168	Solo environment has been created with initial logical time 00:01.000000000
26:50.075719500	INFO	TestTutorial1	solo/solo.go:236	deploying new chain 'ex1'. ID: $/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM, state controller address: 1Aa4X9L2xJVQqxiLFbHj9KXA8YtQqBLHsytuCxBn1kneM
26:50.075719500	INFO	TestTutorial1	solo/solo.go:238	     chain '$/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM'. state controller address: 1Aa4X9L2xJVQqxiLFbHj9KXA8YtQqBLHsytuCxBn1kneM
26:50.075719500	INFO	TestTutorial1	solo/solo.go:239	     chain '$/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM'. originator address: 1CeHWHSLdqfJijBSM3KLqk44MTJBFGrYQ1tJGkKuWqr8q
26:50.075719500	INFO	TestTutorial1.db	dbmanager/dbmanager.go:58	creating new database for chain $/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM. Persistent: false
26:50.107454300	INFO	TestTutorial1	solo/clock.go:35	AdvanceClockBy: logical clock advanced by 2ns to 00:01.000000002
26:50.108564300	INFO	TestTutorial1.ex1	solo/run.go:127	state transition --> #1. Requests in the block: 1. Outputs: 1
26:50.108564300	INFO	TestTutorial1	solo/clock.go:44	ClockStep: logical clock advanced by 1ms to 00:01.001000002
26:50.108564300	INFO	TestTutorial1.ex1	solo/req.go:268	callView: blocklog::getLatestBlockInfo
26:50.108564300	INFO	TestTutorial1.ex1	solo/req.go:268	callView: blocklog::getRequestReceiptsForBlock
26:50.108564300	INFO	TestTutorial1.ex1	solo/run.go:148	REQ: 'tx/[0]J2FrZnvQBbkD5kfPzZFkZfQAK7PYTD9Kh5QsKSaYdBAf'
26:50.108564300	INFO	TestTutorial1.ex1	solo/solo.go:312	chain 'ex1' deployed. Chain ID: $/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM
26:50.108564300	INFO	TestTutorial1.ex1	solo/req.go:268	callView: root::getChainInfo
    tutorial_test.go:28: chain ID: $/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM
    tutorial_test.go:29: chain owner ID: A/1CeHWHSLdqfJijBSM3KLqk44MTJBFGrYQ1tJGkKuWqr8q::00000000
    tutorial_test.go:31:     Core contract 'governance': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::17cf909f
    tutorial_test.go:31:     Core contract 'root': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::cebf5908
    tutorial_test.go:31:     Core contract 'blob': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::fd91bc63
    tutorial_test.go:31:     Core contract 'blocklog': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::f538ef2b
    tutorial_test.go:31:     Core contract 'eventlog': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::661aa7d8
    tutorial_test.go:31:     Core contract 'accounts': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::3c4b5e02
    tutorial_test.go:31:     Core contract '_default': A/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM::00000000
--- PASS: TestTutorial1 (0.04s)
```

:::note
Addresses, chain IDs and other hashes will be the same on each run of the test because of the constant seed. Also, the log produced by the test contains timestamps from computer timer, while the Solo environment operates in its own logical time
:::

The [core contracts](../core_concepts/core_contracts/overview.md) listed in the log (`_default`, `accounts`, `blob`, `blocklog`, `root`) are automatically deployed on each new chain. You can see
them listed in the test log together with their _contract IDs_.

The output fragment in the log `state transition #0 --> #1` means the state of
the chain has changed from block index 0 (the origin index of the empty state)
to block index 1. The state #0 is the empty origin state, the #1 always contains
all core smart contracts deployed on the chain as well as other variables of
chain set, such as _chainID_ and _chain owner ID_.

The _chain ID_ and _chain owner ID_ are respectively ID of the deployed
chain `$/cfQL3Vzay65ZZnPgsDKwXRRGwDWK68QkQwZqzvVs8UXM` and the address (in the
form of _agent ID_) from which the chain has been deployed:
`A/1CeHWHSLdqfJijBSM3KLqk44MTJBFGrYQ1tJGkKuWqr8q::00000000` (the prefixes `$/`
and `A/` indicate that what follows are a chain ID and an agent ID, respectively).