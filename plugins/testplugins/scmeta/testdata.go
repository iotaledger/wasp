package scmeta

import "github.com/iotaledger/wasp/packages/registry"

var testWebHosts = []string{
	"127.0.0.1:8080",
	"127.0.0.1:8081",
	"127.0.0.1:8082",
	"127.0.0.1:8083",
}

var scTestDataJasonable = []*registry.SCMetaDataJsonable{
	{Address: "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9",
		Color:         "93SFFkzdYopchUfee2HWGK769dwgxJmg8TFfXY74ygji",
		OwnerAddress:  "fgCteXu7feArmkR1ArmYsXAWSU3Y7iNA1pmrzxSWwFNV",
		Description:   "test sc for wasp 1",
		ProgramHash:   "9UPqKqEHMX5xBF1DkLqYqosQWLkTorkPx3aqdBxuC84H",
		NodeLocations: testWebHosts,
	},
}
