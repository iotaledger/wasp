// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package scmeta

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/testplugins"
	"github.com/iotaledger/wasp/plugins/webapi"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingSCMetaData"

var (
	Plugin        = node.NewPlugin(PluginName, testplugins.Status(PluginName), configure, run)
	log           *logger.Logger
	nodeLocations = []string{
		"127.0.0.1:4000",
		"127.0.0.1:4001",
		"127.0.0.1:4002",
		"127.0.0.1:4003",
	}
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	go runTestSC(testplugins.SC1)
	go runTestSC(testplugins.SC2)
	go runTestSC(testplugins.SC3)
}

func runTestSC(par apilib.NewOriginParams) {
	// wait for signal from webapi
	webapi.WaitUntilIsUp()

	log.Infof("Start running testing plugin %s for '%s'", PluginName, par.Description)

	myHost := config.Node.GetString(webapi.CfgBindAddress)

	_, scdata := testplugins.CreateOriginData(par, nodeLocations)

	resp := apilib.GetPublicKeyInfo([]string{myHost}, &scdata.Address)
	if len(resp) != 1 {
		log.Errorf("TEST for '%s' FAILED 1: bad response from GetPublicKeyInfo", par.Description)
		return
	}
	failed := false
	if resp[0].Err != "" {
		log.Errorf("response from GetPublicKeyInfo for addr %s: %s", scdata.Address.String(), resp[0].Err)
		failed = true
	} else {
		log.Infof("OK address in registry: %s", scdata.Address.String())
	}
	if failed {
		log.Errorf("TEST FAILED 2: the key with address %s is not available for '%s'",
			par.Address.String(), par.Description)
		return
	}

	writeNew := false
	scDataBack, exists, err := apilib.GetSCMetaData(myHost, &scdata.Address)
	if err != nil {
		log.Errorf("TEST FAILED 3: retrieving SC meta data '%s': %v", scdata.Description, err)
		return
	}
	if exists {
		h1 := hashing.GetHashValue(scdata)
		if scb, err := scDataBack.ToSCMetaData(); err != nil {
			log.Warnf("data will be overwritten: '%s'", scdata.Description)
			writeNew = true
		} else {
			h2 := hashing.GetHashValue(scb)
			if h1 != h2 {
				log.Warnf("data will be overwritten: '%s'", scdata.Description)
				writeNew = true
			}
		}
	} else {
		writeNew = true
	}
	if writeNew {
		log.Infof("writing sc meta data for '%s', address %s", scdata.Description, scdata.Address.String())
		d := scdata.Jsonable()
		if err := apilib.PutSCData(myHost, *d); err != nil {
			log.Errorf("failed writing sc meta data: %v", err)
		}
	} else {
		log.Infof("OK sc meta data for address %s", scdata.Address.String())
	}
	log.Infof("TEST PASSED: '%s'", par.Description)
}
