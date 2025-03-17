package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/tools/stardust-migration/bot"
)

func TestBot(t *testing.T) {
	t.Skip()
	b := bot.Get()

	b.PostWelcomeMessage("TESTMSG")

	for i := range 1000 {
		b.PostStatusUpdate(fmt.Sprintf("TEST STATUS UPDATED %d", i))
		time.Sleep(2 * time.Second)

		if i%10 == 0 {
			b.InvalidateStatusUpdateReference()
		}
	}
}
