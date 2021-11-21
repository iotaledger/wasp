package accounts

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp/colored"

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

var dummyColor colored.Color

func init() {
	var err error
	dummyColor, err = colored.ColorFromBase58EncodedString(hashing.HashStrings("dummy string").Base58())
	if err != nil {
		panic(err)
	}
}

func checkLedger(t *testing.T, state dict.Dict, cp string) colored.Balances {
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

	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	require.NotNil(t, total)
	require.EqualValues(t, 2, len(total))
	require.True(t, total.Equals(transfer))

	transfer = colored.NewBalancesForIotas(1).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	total = checkLedger(t, state, "cp2")

	expected := colored.NewBalancesForIotas(43).Add(dummyColor, 4)
	require.True(t, expected.Equals(total))

	require.EqualValues(t, 43, GetBalanceOld(state, agentID1, colored.IOTA))
	require.EqualValues(t, 4, GetBalanceOld(state, agentID1, dummyColor))
	checkLedger(t, state, "cp2")

	DebitFromAccount(state, agentID1, expected)
	total = checkLedger(t, state, "cp3")
	expected = colored.NewBalances()
	require.True(t, expected.Equals(total))
}

func TestCreditDebit2(t *testing.T) {
	curTest = "TestCreditDebit2"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, len(total))
	require.True(t, expected.Equals(total))

	transfer = colored.NewBalancesForColor(dummyColor, 2)
	DebitFromAccount(state, agentID1, transfer)
	total = checkLedger(t, state, "cp2")
	require.EqualValues(t, 1, len(total))
	expected = colored.NewBalancesForIotas(42)
	require.True(t, expected.Equals(total))

	require.EqualValues(t, 0, GetBalanceOld(state, agentID1, dummyColor))
	bal1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	require.True(t, total.Equals(bal1))
}

func TestCreditDebit3(t *testing.T) {
	curTest = "TestCreditDebit3"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, len(total))
	require.True(t, expected.Equals(total))

	transfer = colored.NewBalancesForColor(dummyColor, 100)
	ok := DebitFromAccount(state, agentID1, transfer)
	require.False(t, ok)
	total = checkLedger(t, state, "cp2")

	require.EqualValues(t, 2, len(total))
	expected = colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	require.True(t, expected.Equals(total))
}

func TestCreditDebit4(t *testing.T) {
	curTest = "TestCreditDebit4"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, len(total))
	require.True(t, expected.Equals(total))

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = colored.NewBalancesForIotas(20)
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedger(t, state, "cp2")

	keys = getAccountsIntern(state).Keys()
	require.EqualValues(t, 2, len(keys))

	expected = colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	require.True(t, expected.Equals(total))

	bm1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	expected = colored.NewBalancesForIotas(22).Add(dummyColor, 2)
	require.True(t, expected.Equals(bm1))

	bm2, ok := GetAccountBalances(state, agentID2)
	require.True(t, ok)
	expected = colored.NewBalancesForIotas(20)
	require.True(t, expected.Equals(bm2))
}

func TestCreditDebit5(t *testing.T) {
	curTest = "TestCreditDebit5"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	total = checkLedger(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 2, len(total))
	require.True(t, expected.Equals(total))

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = colored.NewBalancesForIotas(50)
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.False(t, ok)
	total = checkLedger(t, state, "cp2")

	keys = getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	expected = colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	require.True(t, expected.Equals(total))

	bm1, ok := GetAccountBalances(state, agentID1)
	require.True(t, ok)
	require.True(t, expected.Equals(bm1))

	_, ok = GetAccountBalances(state, agentID2)
	require.False(t, ok)
}

func TestCreditDebit6(t *testing.T) {
	curTest = "TestCreditDebit6"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForIotas(42).Add(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
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
	require.True(t, total.Equals(bal2))
}

func TestCreditDebit7(t *testing.T) {
	curTest = "TestCreditDebit7"
	state := dict.New()
	total := checkLedger(t, state, "cp0")
	require.EqualValues(t, 0, len(total))

	agentID1 := iscp.NewRandomAgentID()
	transfer := colored.NewBalancesForColor(dummyColor, 2)
	CreditToAccountOld(state, agentID1, transfer)
	checkLedger(t, state, "cp1")

	debitTransfer := colored.NewBalancesForIotas(1)
	// debit must fail
	ok := DebitFromAccount(state, agentID1, debitTransfer)
	require.False(t, ok)

	total = checkLedger(t, state, "cp1")
	require.True(t, transfer.Equals(total))
}
