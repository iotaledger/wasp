package ready

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReady1(t *testing.T) {
	r1 := New("ready1")
	err := r1.Wait(1 * time.Second)
	require.Error(t, err)
	t.Logf("%v", err)
}

func TestReady2(t *testing.T) {
	r1 := New("ready1")
	go func() {
		time.Sleep(1 * time.Second)
		r1.SetReady()
	}()
	err := r1.Wait()
	r1.MustWait()
	r1.MustWait()
	r1.MustWait()
	require.NoError(t, err)
}

func TestReady3(t *testing.T) {
	r1 := New("ready1")
	require.Panics(t, func() {
		r1.MustWait(1 * time.Second)
	})
}
