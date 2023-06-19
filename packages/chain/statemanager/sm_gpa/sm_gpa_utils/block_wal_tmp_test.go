package sm_gpa_utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	// "github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestReadBadData(t *testing.T) {
	t.Skip("Demo for Julius")
	for i := 0; i < 20; i++ {
		readBadData()
	}
}

func readBadData() {
	data := make([]byte, 50)
	//nolint:staticcheck // we don't care about weak random numbers here
	rand.Read(data)

	const iterations = 100_000

	now1 := time.Now()
	for n := 0; n < iterations; n++ {
		rr := rwutil.NewBytesReader(data)
		size := rr.ReadSize32()
		for i := 0; i < size; i++ {
			rr.ReadBytes()
			if rr.Err != nil {
				break
			}
		}
	}
	duration1 := time.Since(now1)

	now2 := time.Now()
	for n := 0; n < iterations; n++ {
		buf := bytes.NewBuffer(data)
		size32, _ := rwutil.ReadUint32(buf)
		size := int(size32)
		for i := 0; i < size; i++ {
			bytes, err := rwutil.ReadUint16(buf)
			if err != nil {
				break
			}
			sData := make([]byte, bytes)
			_, err = buf.Read(sData)
			if err != nil {
				break
			}
		}
	}
	duration2 := time.Since(now2)
	fmt.Printf("new %vms, old %vms\n", duration1.Milliseconds(), duration2.Milliseconds())
}
