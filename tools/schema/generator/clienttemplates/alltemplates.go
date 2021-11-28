package clienttemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var config = map[string]string{
	"language":   "Client",
	"extension":  ".ts",
	"rootFolder": "client",
	"funcRegexp": `N/A`,
}

var Templates = []map[string]string{
	config,
	common,
	eventsTs,
}

var TypeDependent = model.StringMapMap{}

var common = map[string]string{
	// *******************************
	"tmp": `
tmp`,
}
