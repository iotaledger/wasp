package testplugins

import "testing"

func TestData(t *testing.T) {
	for i := 0; i < 3; i++ {
		par := GetOriginParams(i)
		t.Logf("dscr: %s\n addr: %s\nowner addr: %s\nprogram hash: %s",
			GetScDescription(i), par.Address.String(), par.OwnerAddress.String(), par.ProgramHash.String())
	}
}
