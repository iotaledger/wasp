package clienttemplates

import "github.com/iotaledger/wasp/tools/schema/model"

var Templates = []map[string]string{
	common,
	eventsTs,
}

var TypeDependent = model.StringMapMap{}

var common = map[string]string{
	// *******************************
	"tmp": `
tmp`,
}
