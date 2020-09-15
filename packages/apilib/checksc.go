package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"io"
	"io/ioutil"
	"os"
)

const prefix = "[checkSC] "

// CheckSC checks and reports deployment data of SC in the given list of node
// it loads bootuo data from the first node in the list and uses CommitteeNodes from that
// bootup data to check the whole committee
//goland:noinspection ALL
func CheckSC(apiHosts []string, scAddr *address.Address, textout ...io.Writer) bool {
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
	fmt.Fprintf(out, prefix+"checking deployment of smart contract at address %s\n", scAddr.String())
	var err error
	var exists, missing bool
	fmt.Fprintf(out, prefix+"loading bootup record from hosts %+v\n", apiHosts)
	var first *registry.BootupData
	var firstHost string

	bdRecords := make([]*registry.BootupData, len(apiHosts))
	for i, host := range apiHosts {
		bdRecords[i], exists, err = GetSCData(host, scAddr)
		if err != nil {
			fmt.Fprintf(out, prefix+"%2d: %s -> %v\n", i, host, err)
			ret = false
			missing = true
			continue
		}
		if !exists {
			fmt.Fprintf(out, prefix+"%2d: %s -> bootup data for %s does not exist\n", i, host, scAddr.String())
			ret = false
			missing = true
			continue
		}
		if bdRecords[i].Address != *scAddr {
			fmt.Fprintf(out, prefix+"%2d: %s -> internal error: wrong address in the bootup record. Expected %s, got %s\n",
				i, host, scAddr.String(), bdRecords[i].Address.String())
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
			fmt.Fprintf(out, prefix+"failed to load bootup data. Exit\n")
			return false
		} else {
			fmt.Fprintf(out, prefix+"some bootup records failed to load\n")
		}
	} else {
		fmt.Fprintf(out, prefix+"bootup records has been loaded from %d nodes\n", len(apiHosts))
	}
	if first != nil {
		fmt.Fprintf(out, prefix+"example bootup record was loaded from %s:\n%s\n", firstHost, first.String())
	}
	for i, bd := range bdRecords {
		host := apiHosts[i]
		if bd == nil {
			fmt.Fprintf(out, prefix+"%2d: %s -> N/A\n", i, host)
			ret = false
			continue
		}
		if bd.Address != *scAddr {
			fmt.Fprintf(out, prefix+"%2d: %s -> internal error, unexpected address %s in the bootupo data record\n",
				i, host, bd.Address.String())
			ret = false
			continue
		}
		if consistentBootupRecords(first, bdRecords[i]) {
			fmt.Fprintf(out, prefix+"%2d: %s -> bootup data OK\n", i, host)
		} else {
			fmt.Fprintf(out, prefix+"%2d: %s -> bootup data is WRONG. Expected equal to example, got %s\n",
				i, host, bdRecords[i].String())
			ret = false
		}
	}

	fmt.Fprintf(out, prefix+"checking distributed keys..\n")

	resps := GetPublicKeyInfoMulti(apiHosts, scAddr)
	var keyExample *dkgapi.GetPubKeyInfoResponse
	for i := range resps {
		if resps[i].Err == "" {
			keyExample = resps[i]
			fmt.Fprintf(out, prefix+"public key info example was taken from %s:\n%s\n", apiHosts[i], publicKeyInfoToString(keyExample))
			break
		}
	}
	for i, resp := range resps {
		host := apiHosts[i]
		if resp.Err != "" {
			fmt.Fprintf(out, prefix+"%2d: %s -> %s\n", i, host, resp.Err)
			ret = false
			continue
		}
		if !consistentPublicKeyInfo(keyExample, resps[i]) {
			fmt.Fprintf(out, prefix+"%2d: %s -> wrong key info. Expected consistent with example, got \n%v\n",
				i, host, resps[i])
			ret = false
			continue
		}
		if i != int(resps[i].Index) {
			fmt.Fprintf(out, prefix+"%2d: %s -> wrong key index. Expected %d, got %d\n", i, host, i, resps[i].Index)
			ret = false
			continue
		}
		fmt.Fprintf(out, prefix+"%2d: %s -> key is OK\n", i, host)

	}
	return ret
}

func consistentPublicKeyInfo(pki1, pki2 *dkgapi.GetPubKeyInfoResponse) bool {
	if pki1.Address != pki2.Address {
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

func publicKeyInfoToString(pki *dkgapi.GetPubKeyInfoResponse) string {
	ret := fmt.Sprintf("    Master public key: %s\n", pki.PubKeyMaster)
	ret += fmt.Sprintf("    N: %d\n", pki.N)
	ret += fmt.Sprintf("    T: %d\n", pki.T)
	ret += fmt.Sprintf("    Public keys: %+v\n", pki.PubKeys)
	return ret
}

func consistentBootupRecords(bd1, bd2 *registry.BootupData) bool {
	if bd1.Address != bd2.Address {
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
