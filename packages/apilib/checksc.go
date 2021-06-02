// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
	"github.com/iotaledger/wasp/packages/webapi/model"
)

const prefix = "[checkSC] "

// CheckDeployment checks and reports deployment data of a chain in the given list of nodes
// it loads the chainrecord from the first node in the list and uses CommitteeNodes from that
// chainrecord to check the whole committee
//goland:noinspection ALL
func CheckDeployment(apiHosts []string, chainID coretypes.ChainID, textout ...io.Writer) bool {
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
	fmt.Fprintf(out, prefix+"checking deployment of smart contract at address %s\n", chainID.String())
	var err error
	var missing bool
	fmt.Fprintf(out, prefix+"loading chainrecord record from hosts %+v\n", apiHosts)
	var first *chainrecord.ChainRecord
	var firstHost string

	bdRecords := make([]*chainrecord.ChainRecord, len(apiHosts))
	for i, host := range apiHosts {
		bdRecords[i], err = client.NewWaspClient(host).GetChainRecord(chainID)
		if err != nil {
			fmt.Fprintf(out, prefix+"%2d: %s -> %v\n", i, host, err)
			ret = false
			missing = true
			continue
		}
		if model.IsHTTPNotFound(err) {
			fmt.Fprintf(out, prefix+"%2d: %s -> chainrecord for %s does not exist\n", i, host, chainID.String())
			ret = false
			missing = true
			continue
		}
		if !bdRecords[i].ChainID.Equals(&chainID) {
			fmt.Fprintf(out, prefix+"%2d: %s -> internal error: wrong address in the chainrecord. Expected %s, got %s\n",
				i, host, chainID.String(), bdRecords[i].ChainID.String())
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
		if !bd.ChainID.Equals(&chainID) {
			fmt.Fprintf(out, prefix+"%2d: %s -> internal error, unexpected address %s in the chain record\n",
				i, host, bd.ChainID.String())
			ret = false
			continue
		}
		if bytes.Equal(first.Bytes(), bdRecords[i].Bytes()) {
			fmt.Fprintf(out, prefix+"%2d: %s -> chainrecord OK\n", i, host)
		} else {
			fmt.Fprintf(out, prefix+"%2d: %s -> chainrecord is wrong. Expected equal to example, got %s\n",
				i, host, bdRecords[i].String())
			ret = false
		}
	}

	fmt.Fprintf(out, prefix+"checking distributed keys..\n")

	chainAddr := chainID.AsAddress()
	dkShares, err := multiclient.New(apiHosts).DKSharesGet(chainAddr)
	if err != nil {
		fmt.Fprintf(out, prefix+"%s\n", err.Error())
		return false
	}

	var keyExample *model.DKSharesInfo
	for i := range dkShares {
		keyExample = dkShares[i]
		fmt.Fprintf(out, prefix+"public key info example was taken from %s:\n%s\n", apiHosts[i], publicKeyInfoToString(keyExample))
		break
	}
	for i, dkShare := range dkShares {
		host := apiHosts[i]
		if !consistentPublicKeyInfo(keyExample, dkShare) {
			fmt.Fprintf(out, prefix+"%2d: %s -> wrong key info. Expected consistent with example, got \n%v\n",
				i, host, dkShare)
			ret = false
			continue
		}
		if dkShare.PeerIndex == nil || i != int(*dkShare.PeerIndex) {
			fmt.Fprintf(out, prefix+"%2d: %s -> wrong key index. Expected %d, got %d\n", i, host, i, dkShares[i].PeerIndex)
			ret = false
			continue
		}
		fmt.Fprintf(out, prefix+"%2d: %s -> key is OK\n", i, host)

	}
	return ret
}

func consistentPublicKeyInfo(pki1, pki2 *model.DKSharesInfo) bool {
	if pki1.Address != pki2.Address {
		return false
	}
	if pki1.SharedPubKey != pki2.SharedPubKey {
		return false
	}
	if pki1.Threshold != pki2.Threshold {
		return false
	}
	if len(pki1.PubKeyShares) != len(pki2.PubKeyShares) {
		return false
	}
	for i := range pki1.PubKeyShares {
		if pki1.PubKeyShares[i] != pki2.PubKeyShares[i] {
			return false
		}
	}
	return true
}

func publicKeyInfoToString(pki *model.DKSharesInfo) string {
	ret := fmt.Sprintf("    Master public key: %s\n", pki.SharedPubKey)
	ret += fmt.Sprintf("    N: %d\n", len(pki.PubKeyShares))
	ret += fmt.Sprintf("    T: %d\n", pki.Threshold)
	ret += fmt.Sprintf("    Public keys: %+v\n", pki.PubKeyShares)
	return ret
}
