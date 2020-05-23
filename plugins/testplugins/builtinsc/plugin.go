// scmeta plugin runs test by creating smart contract meta data records
package builtinsc

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/testplugins"
	"github.com/iotaledger/wasp/plugins/webapi"
	"time"
)

// PluginName is the name of the database plugin.
const PluginName = "TestingBuiltinSC"

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
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		go runInitSC(1, shutdownSignal)
		go runInitSC(2, shutdownSignal)
		go runInitSC(3, shutdownSignal)
	})
	if err != nil {
		log.Errorf("can't start daemon")
	}
}

// reads the registry and checks if initial meta data is correct
// if not:
// - creates new meta data for the smart contracts according to new origin parameters provided
// - creates new origin transaction and posts it to the node
func runInitSC(scIndex int, shutdownSignal <-chan struct{}) {
	// wait for signal from webapi
	webapi.WaitUntilIsUp()

	par := testplugins.GetOriginParams(scIndex)
	log.Infof("Start running testing plugin %s for '%s'", PluginName, testplugins.GetScDescription(scIndex))

	myHost := config.Node.GetString(webapi.CfgBindAddress)

	originTx, scdata := apilib.CreateOriginData(par, testplugins.GetScDescription(scIndex), nodeLocations)

	resp := apilib.GetPublicKeyInfo([]string{myHost}, &scdata.Address)
	if len(resp) != 1 {
		log.Errorf("TEST for '%s' FAILED 1: bad response from GetPublicKeyInfo", testplugins.GetScDescription(scIndex))
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
			par.Address.String(), testplugins.GetScDescription(scIndex))
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
		if err := apilib.PutSCData(myHost, *scdata.Jsonable()); err != nil {
			log.Errorf("failed writing sc meta data: %v", err)
			return
		}
	} else {
		log.Debugf("OK sc meta data for address %s", scdata.Address.String())
	}
	log.Debugf("SC METADATA TEST PASSED: '%s'", testplugins.GetScDescription(scIndex))
	exit := false
	for !exit {
		select {
		case <-shutdownSignal:
			exit = true
		case <-time.After(5 * time.Second):
			if err := nodeconn.PostTransactionToNode(originTx.Transaction); err != nil {
				log.Warnf("failed to send origin tx to node. txid = %s", originTx.ID().String())
			} else {
				log.Debugf("sent origin transaction to node: '%s', txid %s", scdata.Description, originTx.ID().String())
				exit = true
			}
		}
	}
	log.Debugf("SC INIT test routine ended '%s'", testplugins.GetScDescription(scIndex))
}
