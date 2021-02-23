package downloader

import (
	"io/ioutil"
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/plugins/registry"
)

func DonwloadBlobFromHttp(url string) (hashing.HashValue, error) {
	var response *http.Response
	var err error
	response, err = http.Get(url)
	if err != nil {
		return hashing.HashValue{}, err
	}
	defer response.Body.Close()

	var result []byte
	result, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return hashing.HashValue{}, err
	}

	return registry.DefaultRegistry().PutBlob(result)
}
