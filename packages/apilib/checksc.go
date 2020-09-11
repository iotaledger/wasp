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

// CheckSC checks and reports deployment data of SC in the given list of node
// it loads bootuo data from the first node in the list and uses CommitteeNodes from that
// bootup data to check the whole committee
//goland:noinspection ALL
func CheckSC(wasps []string, scAddr *address.Address, textout ...io.Writer) bool {
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
	var bdInit *registry.BootupData
	var initHost string
	var exists bool
	var err error
	// TODO wrong --> rewrite with api hosts
	for _, host := range wasps {
		bdInit, exists, err = GetSCData(host, scAddr)
		if err != nil {
			fmt.Fprintf(out, "GetSCData: %v\n", err)
			ret = false
			continue
		}
		if !exists {
			fmt.Fprintf(out, "GetSCData: bootup record for address %s does not exist in %s\n", scAddr.String(), host)
			ret = false
			continue
		}
		initHost = host
		break
	}
	if bdInit == nil {
		err := fmt.Errorf("failed to load initial bootup record for address %s from %+v", scAddr.String(), wasps)
		fmt.Fprintf(out, "%v\n", err)
		return false
	}
	fmt.Fprintf(out, "loaded example bootup record from node %s. Will be loading bootup data from commitee nodes:\n%s", initHost, bdInit.String())
	bdRecords := make([]*registry.BootupData, len(bdInit.CommitteeNodes))
	for i := range bdRecords {
		host := bdInit.CommitteeNodes[i]
		bdRecords[i], exists, err = GetSCData(host, scAddr)
		if err != nil {
			fmt.Fprintf(out, "%2d: %s -> %v\n", i, host, err)
			ret = false
			continue
		}
		if !exists {
			fmt.Fprintf(out, "%2d: %s -> bootup data for %s does not exist\n", i, host, scAddr.String())
			ret = false
			continue
		}
		if bdRecords[i].Address != *scAddr {
			fmt.Fprintf(out, "%2d: %s -> internal error: wrong address in the bootup record. Expected %s, got %s\n",
				i, err, scAddr.String(), bdRecords[i].Address.String())
			ret = false
			continue
		}
		if consistentBootupRecords(bdInit, bdRecords[i]) {
			fmt.Fprintf(out, "%2d: %s -> bootup data record is OK. Access nodes: %+v\n", i, host, bdRecords[i].AccessNodes)
		} else {
			fmt.Fprintf(out, "%2d: %s -> bootup data records is WRONG. Expected to be equal to the example, got:\n%s",
				i, host, bdRecords[i].String())
			ret = false
		}
	}
	fmt.Fprintf(out, "checking distributed keys..\n")
	pkinfo := GetPublicKeyInfo(initHost, scAddr)
	if pkinfo.Err != "" {
		fmt.Fprintf(out, "failed to load public key info for %s from %s\n", scAddr.String(), initHost)
		return false
	}
	fmt.Fprintf(out, "loaded example public key info for %s from %s\n%s",
		scAddr.String(), initHost, publicKeyInfoToString(pkinfo))

	resps := GetPublicKeyInfoMulti(bdInit.CommitteeNodes, scAddr)
	for i := range resps {
		host := bdInit.CommitteeNodes[i]
		if resps[i].Err != "" {
			fmt.Fprintf(out, "%2d: %s -> %s\n", i, host, resps[i].Err)
			ret = false
			continue
		}
		if !consistentPublicKeyInfo(pkinfo, resps[i]) {
			fmt.Fprintf(out, "%2d: %s -> wrong key info. Expected consistent with example, got \n%s\n",
				i, host, resps[i])
			ret = false
			continue
		}
		if i != int(resps[i].Index) {
			fmt.Fprintf(out, "%2d: %s -> wrong key index. Expected %d, got %d\n", i, host, i, resps[i].Index)
			ret = false
			continue
		}
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
