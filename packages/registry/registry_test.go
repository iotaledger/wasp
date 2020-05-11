package registry

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	ownAddr := PortAddr{
		Port: 1000,
		Addr: "127.0.0.1",
	}
	list, err := GetSCDataList(&ownAddr)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(list), 0)
}
