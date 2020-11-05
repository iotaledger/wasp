package apilib

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"io"
	"io/ioutil"
	"os"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/registry"
)

const prefix = "[checkSC] "

// CheckDeployment checks and reports deployment data of a chain in the given list of nodes
// it loads the chainrecord from the first node in the list and uses CommitteeNodes from that
// chainrecord to check the whole committee
//goland:noinspection ALL
func CheckDeployment(apiHosts []string, chainid *coretypes.ChainID, textout ...io.Writer) bool {
	ret := true
	var out io.Writer
	if len(textout) == 0 {
		out = os.Stdout
	} else {
		if textout[0] != nil {
			out = textout[0]
		} else {
			out = ioutil.Discard
		}
	}
	fmt.Fprintf(out, prefix+"checking deployment of smart contract at address %s\n", chainid.String())
	var err error
	var missing bool
	fmt.Fprintf(out, prefix+"loading chainrecord record from hosts %+v\n", apiHosts)
	var first *registry.ChainRecord
	var firstHost string

	bdRecords := make([]*registry.ChainRecord, len(apiHosts))
	for i, host := range apiHosts {
		bdRecords[i], err = client.NewWaspClient(host).GetChainRecord(chainid)
		if err != nil {
			fmt.Fprintf(out, prefix+"%2d: %s -> %v\n", i, host, err)
			ret = false
			missing = true
			continue
		}
		if client.IsNotFound(err) {
			fmt.Fprintf(out, prefix+"%2d: %s -> chainrecord for %s does not exist\n", i, host, chainid.String())
			ret = false
			missing = true
			continue
		}
		if bdRecords[i].ChainID != *chainid {
			fmt.Fprintf(out, prefix+"%2d: %s -> internal error: wrong address in the chainrecord. Expected %s, got %s\n",
				i, host, chainid.String(), bdRecords[i].ChainID.String())
			ret = false
			missing = true
			continue
		}
		if first == nil {
			first = bdRecords[i]
			firstHost = host
		}
	}
	if missing {
		if first == nil {
			fmt.Fprintf(out, prefix+"failed to load chainrecord. Exit\n")
			return false
		} else {
			fmt.Fprintf(out, prefix+"some chain records failed to load\n")
		}
	} else {
		fmt.Fprintf(out, prefix+"chain records have been loaded from %d nodes\n", len(apiHosts))
	}
	if first != nil {
		fmt.Fprintf(out, prefix+"example chain record was loaded from %s:\n%s\n", firstHost, first.String())
	}
	for i, bd := range bdRecords {
		host := apiHosts[i]
		if bd == nil {
			fmt.Fprintf(out, prefix+"%2d: %s -> N/A\n", i, host)
			ret = false
			continue
		}
		if bd.ChainID != *chainid {
			fmt.Fprintf(out, prefix+"%2d: %s -> internal error, unexpected address %s in the chain record\n",
				i, host, bd.ChainID.String())
			ret = false
			continue
		}
		if consistentChainRecords(first, bdRecords[i]) {
			fmt.Fprintf(out, prefix+"%2d: %s -> chainrecord OK\n", i, host)
		} else {
			fmt.Fprintf(out, prefix+"%2d: %s -> chainrecord is WRONG. Expected equal to example, got %s\n",
				i, host, bdRecords[i].String())
			ret = false
		}
	}

	fmt.Fprintf(out, prefix+"checking distributed keys..\n")

	scAddr := (address.Address)(*chainid)
	resps, err := multiclient.New(apiHosts).GetPublicKeyInfo(&scAddr)
	if err != nil {
		fmt.Fprintf(out, prefix+"%s\n", err.Error())
		return false
	}

	var keyExample *client.PubKeyInfo
	for i := range resps {
		keyExample = resps[i]
		fmt.Fprintf(out, prefix+"public key info example was taken from %s:\n%s\n", apiHosts[i], publicKeyInfoToString(keyExample))
		break
	}
	for i, resp := range resps {
		host := apiHosts[i]
		if !consistentPublicKeyInfo(keyExample, resp) {
			fmt.Fprintf(out, prefix+"%2d: %s -> wrong key info. Expected consistent with example, got \n%v\n",
				i, host, resp)
			ret = false
			continue
		}
		if i != int(resp.Index) {
			fmt.Fprintf(out, prefix+"%2d: %s -> wrong key index. Expected %d, got %d\n", i, host, i, resps[i].Index)
			ret = false
			continue
		}
		fmt.Fprintf(out, prefix+"%2d: %s -> key is OK\n", i, host)

	}
	return ret
}

func consistentPublicKeyInfo(pki1, pki2 *client.PubKeyInfo) bool {
	if *pki1.Address.Address() != *pki2.Address.Address() {
		return false
	}
	if pki1.PubKeyMaster != pki2.PubKeyMaster {
		return false
	}
	if pki1.N != pki2.N {
		return false
	}
	if pki1.T != pki2.T {
		return false
	}
	if len(pki1.PubKeys) != len(pki2.PubKeys) {
		return false
	}
	for i := range pki1.PubKeys {
		if pki1.PubKeys[i] != pki2.PubKeys[i] {
			return false
		}
	}
	return true
}

func publicKeyInfoToString(pki *client.PubKeyInfo) string {
	ret := fmt.Sprintf("    Master public key: %s\n", pki.PubKeyMaster)
	ret += fmt.Sprintf("    N: %d\n", pki.N)
	ret += fmt.Sprintf("    T: %d\n", pki.T)
	ret += fmt.Sprintf("    Public keys: %+v\n", pki.PubKeys)
	return ret
}

func consistentChainRecords(bd1, bd2 *registry.ChainRecord) bool {
	if bd1.ChainID != bd2.ChainID {
		return false
	}
	if bd1.OwnerAddress != bd2.OwnerAddress {
		return false
	}
	if bd1.Color != bd2.Color {
		return false
	}
	if len(bd1.CommitteeNodes) != len(bd2.CommitteeNodes) {
		return false
	}
	for i := range bd1.CommitteeNodes {
		if bd1.CommitteeNodes[i] != bd2.CommitteeNodes[i] {
			return false
		}
	}
	// access nodes can be any, do not check
	return true
}
