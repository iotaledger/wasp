// a DUMMY file to check how build tags work
// the package will contain all dependencies with the Chrysalis color model
// only included for IOTA 1.5 ledger

//+build l1_15

package colored

const ColorLength = 42

var (
	IOTA = Color{}
	MINT = Color{}
)

func init() {
	for i := range MINT {
		MINT[i] = 0xFF
	}
}

// TODO TODO
