package generator

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/tools/schema/model"
)

type ClientBase struct {
	GenBase
}

func (g *ClientBase) Generate() error {
	g.folder = g.rootFolder + "/" + g.s.PackageName + "client/"
	if g.s.CoreContracts {
		g.folder = g.rootFolder + "/wasmclient/" + g.s.PackageName + "/"
	}
	err := os.MkdirAll(g.folder, 0o755)
	if err != nil {
		return err
	}
	info, err := os.Stat(g.folder + "service" + g.extension)
	if err == nil && info.ModTime().After(g.s.SchemaTime) {
		fmt.Printf("skipping %s code generation\n", g.language)
		return nil
	}

	fmt.Printf("generating %s code\n", g.language)
	return g.generateCode()
}

func (g *ClientBase) generateCode() error {
	err := g.createSourceFile("events", !g.s.CoreContracts)
	if err != nil {
		return err
	}
	err = g.createSourceFile("service", true)
	if err != nil {
		return err
	}
	if !g.s.CoreContracts {
		return g.generateFuncs(g.appendEvents)
	}
	return nil
}

func (g *ClientBase) appendEvents(existing model.StringMap) {
	for _, g.currentEvent = range g.s.Events {
		name := g.s.ContractName + capitalize(g.currentEvent.Name)
		if existing[name] == "" {
			g.log("currentEvent: " + g.currentEvent.Name)
			g.setMultiKeyValues("evtName", g.currentEvent.Name)
			g.emit("funcSignature")
		}
	}
}
