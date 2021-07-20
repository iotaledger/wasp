package accounts

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Logf("Name: %s", Contract.Name)
	t.Logf("Description: %s", Contract.Description)
	t.Logf("Program hash: %s", Contract.ProgramHash.String())
	t.Logf("Hname: %s", Contract.Hname())
}

var color = ledgerstate.Color(hashing.HashStrings("dummy string"))

func checkLedger(t *testing.T, state dict.Dict, cp string) *ledgerstate.ColoredBalances {
	total := GetTotalAssets(state)
	t.Logf("checkpoint '%s.%s':\n%s", curTest, cp, total.String())
	require.NotPanics(t, func() {
		mustCheckLedger(state, cp)
	})
	return total
}

var curTest = ""

func TestCreditDebit1(t *testing.T) {
	curTest = "TestCreditDebit1"
	state := dict.New()
	total := checkLedger(t, state, "cp0")

	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	require.NotNil(t, total)
	require.EqualValues(t, 2, total.Size())
	require.True(t, iscp.EqualColoredBalances(total, transfer))

	transfer = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 1,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp2")

	expected := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 43,
		color:                 4,
	})
	require.True(t, iscp.EqualColoredBalances(expected, total))

	require.EqualValues(t, 43, GetBalance(state, agentID1, ledgerstate.ColorIOTA))
	require.EqualValues(t, 4, GetBalance(state, agentID1, color))
	checkLedger(t, state, "cp2")

	DebitFromAccount(state, agentID1, expected)
	total = checkLedger(t, state, "cp3")
	expected = ledgerstate.NewColoredBalances(nil)
	require.True(t, iscp.EqualColoredBalances(expected, total))
}

func TestCreditDebit2(t *testing.T) {
	curTest = "TestCreditDebit2"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Size())
	require.True(t, iscp.EqualColoredBalances(expected, total))

	transfer = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		color: 2,
	})
	DebitFromAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp2")
	require.EqualValues(t, 1, total.Size())
	expected = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
	})
	require.True(t, iscp.EqualColoredBalances(expected, total))

	require.EqualValues(t, 0, GetBalance(state, agentID1, color))
	bal1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	require.True(t, iscp.EqualColoredBalances(total, ledgerstate.NewColoredBalances(bal1)))
}

func TestCreditDebit3(t *testing.T) {
	curTest = "TestCreditDebit3"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Size())
	require.True(t, iscp.EqualColoredBalances(expected, total))

	transfer = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		color: 100,
	})
	ok := DebitFromAccount(state, agentID1, transfer)
	require.False(t, ok)
	total = checkLedger(t, state, "cp2")

	require.EqualValues(t, 2, total.Size())
	expected = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	require.True(t, iscp.EqualColoredBalances(expected, total))
}

func TestCreditDebit4(t *testing.T) {
	curTest = "TestCreditDebit4"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Size())
	require.True(t, iscp.EqualColoredBalances(expected, total))

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 20,
	})
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedger(t, state, "cp2")

	keys = getAccountsIntern(state).Keys()
	require.EqualValues(t, 2, len(keys))

	expected = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	require.True(t, iscp.EqualColoredBalances(expected, total))

	bm1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	expected = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 22,
		color:                 2,
	})
	require.True(t, iscp.EqualColoredBalances(expected, ledgerstate.NewColoredBalances(bm1)))

	bm2, ok := GetAccountBalances(state, agentID2)
	require.True(t, ok)
	expected = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 20,
	})
	require.True(t, iscp.EqualColoredBalances(expected, ledgerstate.NewColoredBalances(bm2)))
}

func TestCreditDebit5(t *testing.T) {
	curTest = "TestCreditDebit5"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Size())
	require.True(t, iscp.EqualColoredBalances(expected, total))

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 50,
	})
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.False(t, ok)
	total = checkLedger(t, state, "cp2")

	keys = getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	expected = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	require.True(t, iscp.EqualColoredBalances(expected, total))

	bm1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	require.True(t, iscp.EqualColoredBalances(expected, ledgerstate.NewColoredBalances(bm1)))

	_, ok = GetAccountBalances(state, agentID2)
	require.False(t, ok)
}

func TestCreditDebit6(t *testing.T) {
	curTest = "TestCreditDebit6"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 42,
		color:                 2,
	})
	CreditToAccount(state, agentID1, transfer)
	checkLedger(t, state, "cp1")

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedger(t, state, "cp2")

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	_, ok = GetAccountBalances(state, agentID1)
	require.False(t, ok)

	bal2, ok := GetAccountBalances(state, agentID2)
	require.True(t, ok)
	require.True(t, iscp.EqualColoredBalances(total, ledgerstate.NewColoredBalances(bal2)))
}

func TestCreditDebit7(t *testing.T) {
	curTest = "TestCreditDebit7"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Size())

	agentID1 := iscp.NewRandomAgentID()
	transfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		color: 2,
	})
	CreditToAccount(state, agentID1, transfer)
	checkLedger(t, state, "cp1")

	debitTransfer := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 1,
	})
	// debit must fail
	ok := DebitFromAccount(state, agentID1, debitTransfer)
	require.False(t, ok)

	total = checkLedger(t, state, "cp1")
	require.True(t, iscp.EqualColoredBalances(transfer, total))
}
