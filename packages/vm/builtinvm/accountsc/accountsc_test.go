package accountsc

import "testing"

func TestBasic(t *testing.T) {
	t.Logf("Name: %s", Name)
	t.Logf("Version: %s", Version)
	t.Logf("Full name: %s", FullName)
	t.Logf("Description: %s", Description)
	t.Logf("Program hash: %s", ProgramHash.String())
	t.Logf("Hname: %s", Hname)
	t.Logf("Total assets account: %s", TotalAssetsAccountID.String())
}
