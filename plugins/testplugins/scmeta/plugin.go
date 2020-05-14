// scmeta package runs integration tests by calling WebAPi to itself for SC meta data
package scmeta

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/testplugins"
	"github.com/iotaledger/wasp/plugins/webapi"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingSCMetaData"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, testplugins.Status(PluginName), configure, run)
	log    *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	go func() {
		// wait for signal from webapi
		webapi.WaitIsUp()

		log.Infof("Start running testing: %s", PluginName)

		// convert data
		scTestData := make([]*registry.SCMetaData, len(scTestDataJasonable))
		var err error
		for i := range scTestData {
			scTestData[i], err = scTestDataJasonable[i].NewSCMetaData()
			if err != nil {
				log.Errorf("TEST FAILED 1, wring testing data: %v", err)
				return
			}
		}

		myHost := config.Node.GetString(webapi.CfgBindAddress)

		failed := false
		// check if all keys exist in local registry
		for _, scdata := range scTestData {
			resp := apilib.GetPublicKeyInfo([]string{myHost}, &scdata.Address)
			if len(resp) != 1 {
				log.Errorf("TEST FAILED 3: bad response from GetPublicKeyInfo")
				return
			}
			if resp[0].Err != "" {
				log.Errorf("response from GetPublicKeyInfo for addr %s: %s", scdata.Address.String(), resp[0].Err)
				failed = true
			} else {
				log.Infof("address in registry OK: %s", scdata.Address.String())
			}
		}
		if failed {
			log.Errorf("TEST FAILED 4: not all keys are available")
			return
		}

		// retrieving SC meta data from registry calling WebAPI to itself
		for i, scdata := range scTestData {
			writeNew := false
			sc1back, exists, err := apilib.GetSCMetaData(myHost, &scdata.Address)
			if err != nil {
				log.Errorf("TEST FAILED 5: retrieving SC meta data for %s: %v", scdata.Address.String(), err)
				return
			}
			if exists {
				if !equalSC(scTestDataJasonable[i], sc1back) {
					log.Warnf("inconsistency. Data to be replaced for address %s.", scdata.Address.String())
					writeNew = true
				}
			} else {
				writeNew = true
			}
			if writeNew {
				log.Infof("writing sc meta data for address %s", scdata.Address.String())

				if err := apilib.PutSCData(myHost, *scTestDataJasonable[i]); err != nil {
					log.Errorf("failed writing sc meta data: %v", err)
				}
			}
		}

		log.Infof("TEST PASSED")
	}()
}

func equalSC(sc1, sc2 *registry.SCMetaDataJsonable) bool {
	return sc1.Address == sc2.Address &&
		sc1.Color == sc2.Color &&
		sc1.OwnerAddress == sc2.OwnerAddress &&
		sc1.Description == sc2.Description &&
		sc1.ProgramHash == sc2.ProgramHash
}
