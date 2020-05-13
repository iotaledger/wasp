package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/apilib"
	"io/ioutil"
	"os"
)

type ioParams struct {
	Hosts     []string `json:"hosts"`
	N         uint16   `json:"n"`
	T         uint16   `json:"t"`
	NumKeys   uint16   `json:"num_keys"`
	Addresses []string `json:"addresses"` //base58
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage newdks <input file path>\n")
		os.Exit(1)
	}
	fname := os.Args[1]
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	if len(params.Hosts) != int(params.N) || params.N < params.T || params.N < 4 {
		panic("wrong assembly size parameters or number rof hosts")
	}

	params.Addresses = make([]string, 0, params.NumKeys)
	numSuccess := 0
	for i := 0; i < int(params.NumKeys); i++ {
		addr, err := apilib.GenerateNewDistributedKeySet(params.Hosts, params.N, params.T)
		if err == nil {
			params.Addresses = append(params.Addresses, addr.String())
			numSuccess++
			fmt.Printf("generated new key. Address: %s\n", addr.String())
		} else {
			fmt.Printf("error: %v\n", err)
		}
	}
	data, err = json.MarshalIndent(&params, "", " ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = ioutil.WriteFile(fname+".resp.json", data, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	if numSuccess == 0 {
		return
	}
	//----- crosscheck
	fmt.Printf("crosschecking. Reading public keys back\n")
	for _, addr := range params.Addresses {
		fmt.Printf("crosschecking. Address %s\n", addr)
		a, err := address.FromBase58(addr)
		if err != nil {
			fmt.Printf("%s --> %v\n", addr, err)
			continue
		}
		resps := apilib.GetPublicKeyInfo(params.Hosts, &a)
		for i, r := range resps {
			if r == nil || r.Address != addr || r.N != params.N || r.T != params.T || int(r.Index) != i {
				fmt.Printf("%s --> returned none or wrong values\n", params.Hosts[i])
			} else {
				fmt.Printf("%s --> master pub key: %s\n", params.Hosts[i], r.PubKeyMaster)
			}
		}
	}
}
