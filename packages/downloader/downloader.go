package downloader

import (
	"strings"
	"sync"

	"io/ioutil"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

var downloads map[string]bool // Just a HashSet. The value of the element is not important. The existence of key in the map is what counts.
var downloadsMutex = &sync.Mutex{}

func DownloadAndStore(hash hashing.HashValue, uri string, cache coretypes.BlobCacheFull) error {
	if contains(uri) {
		//TODO log
		return nil
	} else {
		markStarted(uri)
		go func() {
			defer markCompleted(uri)

			var split []string = strings.SplitN(uri, "://", 2)
			if len(split) != 2 {
				//TODO log
				return
			}

			var protocol string = split[0]
			var path string = split[1]
			var download []byte
			var err error
			switch protocol {
			case "ipfs":
				download, err = DonwloadFromHttp("https://ipfs.io/ipfs/" + path)
			case "http":
				download, err = DonwloadFromHttp(uri)
			default:
				//TODO log
				return
			}

			if err != nil {
				//TODO log
				return
			}

			var cacheHash hashing.HashValue
			cacheHash, err = cache.PutBlob(download)

			if err != nil {
				//TODO log
				return
			}

			if hash != cacheHash {
				//TODO log
				return
			}

		}()

		return nil
	}
}

func contains(uri string) bool {
	var ok bool
	downloadsMutex.Lock()
	_, ok = downloads[uri]
	downloadsMutex.Unlock()
	return ok
}

func markStarted(uri string) {
	downloadsMutex.Lock()
	downloads[uri] = true
	downloadsMutex.Unlock()
}

func markCompleted(uri string) {
	downloadsMutex.Lock()
	delete(downloads, uri)
	downloadsMutex.Unlock()
}

func DonwloadFromHttp(url string) ([]byte, error) {
	var response *http.Response
	var err error
	response, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
