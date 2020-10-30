package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"testing"
	"time"
)

const (
	erc20_wasmPath    = "wasm/erc20"
	erc20_description = "ERC-20, a PoC smart contract"

	erc20_req_init_sc  = coretypes.EntryPointCode(1)
	erc20_req_transfer = coretypes.EntryPointCode(2)
	erc20_req_approve  = coretypes.EntryPointCode(3)

	erc20_var_supply         = "supply"
	erc20_var_target_address = "addr"
	erc20_var_amount         = "amount"
)

func TestDeploymentERC20(t *testing.T) {
	if *useWasp {
		t.Fatal("erc20 test is only for wasm SC code")
		return
	}
	wasps := setup(t, "TestDeploymentErc20")

	err := loadWasmIntoWasps(wasps, erc20_wasmPath, erc20_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, "", erc20_description)
	checkSuccess(err, t, "smart contract has been created and activated")
	_ = scChain

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  erc20_description,
	}) {
		t.Fail()
	}
}

func TestInitERC20Once(t *testing.T) {
	if *useWasp {
		t.Fatal("erc20 test is only for wasm SC code")
		return
	}
	wasps := setup(t, "TestInitERC20Once")

	err := loadWasmIntoWasps(wasps, erc20_wasmPath, erc20_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, "", erc20_description)
	checkSuccess(err, t, "smart contract has been created and activated")

	client := chainclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scChain,
		scOwner.SigScheme(),
		1*time.Minute,
	)

	supply := int64(1000000)
	_, err = client.PostRequest(0, erc20_req_init_sc, nil, nil, map[string]interface{}{
		erc20_var_supply: supply,
	})
	checkSuccess(err, t, "posted initSC request")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  erc20_description,
		erc20_var_supply:            supply,
	}) {
		t.Fail()
	}
}

func TestInitERC20Twice(t *testing.T) {
	if *useWasp {
		t.Fatal("erc20 test is only for wasm SC code")
		return
	}
	wasps := setup(t, "TestInitERC20Twice")

	err := loadWasmIntoWasps(wasps, erc20_wasmPath, erc20_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, "", erc20_description)
	checkSuccess(err, t, "smart contract has been created and activated")

	client := chainclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scChain,
		scOwner.SigScheme(),
		1*time.Minute,
	)

	supply := int64(1000000)
	_, err = client.PostRequest(0, erc20_req_init_sc, nil, nil, map[string]interface{}{
		erc20_var_supply: supply,
	})
	checkSuccess(err, t, "posted initSC request 1")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  erc20_description,
		erc20_var_supply:            supply,
	}) {
		t.Fail()
	}

	// second time should not pass
	_, err = client.PostRequest(0, erc20_req_init_sc, nil, nil, map[string]interface{}{
		erc20_var_supply: supply - 1000,
	})

	checkSuccess(err, t, "posted initSC request 2")

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  erc20_description,
		erc20_var_supply:            supply,
	}) {
		t.Fail()
	}
}

func TestTransferOk(t *testing.T) {
	if *useWasp {
		t.Fatal("erc20 test is only for wasm SC code")
		return
	}
	wasps := setup(t, "TestTransferOk")

	err := loadWasmIntoWasps(wasps, erc20_wasmPath, erc20_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, "", erc20_description)
	checkSuccess(err, t, "smart contract has been created and activated")

	client := chainclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scChain,
		scOwner.SigScheme(),
		1*time.Minute,
	)

	supply := int64(1000000)
	_, err = client.PostRequest(0, erc20_req_init_sc, nil, nil, map[string]interface{}{
		erc20_var_supply: supply,
	})
	checkSuccess(err, t, "posted initSC request")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  erc20_description,
		erc20_var_supply:            supply,
	}) {
		t.Fail()
	}

	_, err = client.PostRequest(0, erc20_req_transfer, nil, nil, map[string]interface{}{
		erc20_var_target_address: "fake_address",
		erc20_var_amount:         1337,
	})
	checkSuccess(err, t, "posted transfer request")

	// TODO check properly the dictionary state
	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  erc20_description,
		erc20_var_supply:            supply,
	}) {
		t.Fail()
	}
}
