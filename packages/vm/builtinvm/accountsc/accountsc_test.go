package accountsc

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Logf("Name: %s", Name)
	t.Logf("Version: %s", Version)
	t.Logf("Full name: %s", Interface.Name)
	t.Logf("Description: %s", Interface.Description)
	t.Logf("Program hash: %s", Interface.ProgramHash.String())
	t.Logf("Hname: %s", Interface.Hname())
}

var color = balance.Color(*hashing.HashStrings("dummy string"))

func checkLedger(t *testing.T, state dict.Dict, cp string) coretypes.ColoredBalances {
	total := GetTotalAssets(state)
	t.Logf("checkpoint '%s.%s':\n%s", curTest, cp, total.String())
	require.NotPanics(t, func() {
		MustCheckLedger(state, cp)
	})
	return total
}

var curTest = ""

func TestCreditDebit1(t *testing.T) {
	curTest = "TestCreditDebit1"
	state := dict.New()
	total := checkLedger(t, state, "cp0")

	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	require.NotNil(t, total)
	require.EqualValues(t, 2, total.Len())
	require.True(t, total.Equal(transfer))

	transfer = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 1,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp2")

	expected := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 43,
		color:             4,
	})
	require.True(t, expected.Equal(total))

	require.EqualValues(t, 43, GetBalance(state, agentID1, balance.ColorIOTA))
	require.EqualValues(t, 4, GetBalance(state, agentID1, color))
	checkLedger(t, state, "cp2")

	DebitFromAccount(state, agentID1, expected)
	total = checkLedger(t, state, "cp3")
	expected = cbalances.Nil
	require.True(t, expected.Equal(total))
}

func TestCreditDebit2(t *testing.T) {
	curTest = "TestCreditDebit2"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Len())
	require.True(t, expected.Equal(total))

	transfer = cbalances.NewFromMap(map[balance.Color]int64{
		color: 2,
	})
	DebitFromAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp2")
	require.EqualValues(t, 1, total.Len())
	expected = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
	})
	require.True(t, expected.Equal(total))

	require.EqualValues(t, 0, GetBalance(state, agentID1, color))
	bal1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	require.True(t, total.Equal(cbalances.NewFromMap(bal1)))
}

func TestCreditDebit3(t *testing.T) {
	curTest = "TestCreditDebit3"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Len())
	require.True(t, expected.Equal(total))

	transfer = cbalances.NewFromMap(map[balance.Color]int64{
		color: 100,
	})
	ok := DebitFromAccount(state, agentID1, transfer)
	require.False(t, ok)
	total = checkLedger(t, state, "cp2")

	require.EqualValues(t, 2, total.Len())
	expected = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	require.True(t, expected.Equal(total))
}

func TestCreditDebit4(t *testing.T) {
	curTest = "TestCreditDebit4"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Len())
	require.True(t, expected.Equal(total))

	keys := GetAccounts(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := coretypes.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 20,
	})
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedger(t, state, "cp2")

	keys = GetAccounts(state).Keys()
	require.EqualValues(t, 2, len(keys))

	expected = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	require.True(t, expected.Equal(total))

	bm1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	expected = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 22,
		color:             2,
	})
	require.True(t, expected.Equal(cbalances.NewFromMap(bm1)))

	bm2, ok := GetAccountBalances(state, agentID2)
	require.True(t, ok)
	expected = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 20,
	})
	require.True(t, expected.Equal(cbalances.NewFromMap(bm2)))
}

func TestCreditDebit5(t *testing.T) {
	curTest = "TestCreditDebit5"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, total.Len())
	require.True(t, expected.Equal(total))

	keys := GetAccounts(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := coretypes.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 50,
	})
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.False(t, ok)
	total = checkLedger(t, state, "cp2")

	keys = GetAccounts(state).Keys()
	require.EqualValues(t, 1, len(keys))

	expected = cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	require.True(t, expected.Equal(total))

	bm1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	require.True(t, expected.Equal(cbalances.NewFromMap(bm1)))

	_, ok = GetAccountBalances(state, agentID2)
	require.False(t, ok)
}

func TestCreditDebit6(t *testing.T) {
	curTest = "TestCreditDebit6"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 42,
		color:             2,
	})
	CreditToAccount(state, agentID1, transfer)
	checkLedger(t, state, "cp1")

	agentID2 := coretypes.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedger(t, state, "cp2")

	keys := GetAccounts(state).Keys()
	require.EqualValues(t, 1, len(keys))

	_, ok = GetAccountBalances(state, agentID1)
	require.False(t, ok)

	bal2, ok := GetAccountBalances(state, agentID2)
	require.True(t, ok)
	require.True(t, total.Equal(cbalances.NewFromMap(bal2)))
}

func TestCreditDebit7(t *testing.T) {
	curTest = "TestCreditDebit7"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, total.Len())

	agentID1 := coretypes.NewRandomAgentID()
	transfer := cbalances.NewFromMap(map[balance.Color]int64{
		color: 2,
	})
	CreditToAccount(state, agentID1, transfer)
	checkLedger(t, state, "cp1")

	debitTransfer := cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 1,
	})
	// debit must fail
	ok := DebitFromAccount(state, agentID1, debitTransfer)
	require.False(t, ok)

	total = checkLedger(t, state, "cp1")
	require.True(t, transfer.Equal(total))
}
