// the package contain all dependencies with the hornet color model
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
