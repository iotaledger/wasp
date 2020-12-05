package alone

import (
	"testing"
)

func TestBasic(t *testing.T) {
	al := New(t, false)
	al.CheckBase()
	al.Infof("\n%s\n", al.String())
}
