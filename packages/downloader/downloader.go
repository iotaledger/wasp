package downloader

import (
	"strings"
	"sync"

	"io/ioutil"
	"net/http"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/parameters"
)

var log *logger.Logger

var downloads map[string]bool // Just a HashSet. The value of the element is not important. The existence of key in the map is what counts.
var downloadsMutex *sync.Mutex

func Init(inLog *logger.Logger) {
	log = inLog
	downloads = make(map[string]bool)
	downloadsMutex = &sync.Mutex{}
}

// Accepted URIs:
//  * http://<url of the contents>
//          e.g. http://some.place.lt/some/contents.txt
//  * https://<url of the contents>
//          e.g. https://some.place.lt/some/contents.txt
//  * ipfs://<cid of the contents>
//          e.g. ipfs://QmeyMc1i9KLqqyqYCksDZiwntxwuiz5Z1hbLBrHvAXyjMZ
func DownloadAndStore(hash hashing.HashValue, uri string, cache coretypes.BlobCacheFull) error {
	if contains(uri) {
		log.Warnf("File %s is already being downloaded. Skipping it.", uri)
		return nil
	} else {
		markStarted(uri)
		go func() {
			defer markCompleted(uri)

			var split []string = strings.SplitN(uri, "://", 2)
			if len(split) != 2 {
				log.Errorf("File uri %s is invalid.", uri)
				return
			}

			var protocol string = split[0]
			var path string = split[1]
			var download []byte
			var err error
			switch protocol {
			case "ipfs":
				download, err = DonwloadFromHttp(parameters.IpfsGatewayAddress + "/ipfs/" + path)
			case "http":
				download, err = DonwloadFromHttp(uri)
			case "https":
				download, err = DonwloadFromHttp(uri)
			default:
				log.Errorf("Unknown protocol %s of uri %s.", protocol, uri)
				return
			}

			if err != nil {
				log.Errorf("Error retrieving file %s: %s.", uri, err)
				return
			}

			var cacheHash hashing.HashValue
			cacheHash, err = cache.PutBlob(download)

			if err != nil {
				log.Errorf("Error putting file %s to cache: %s.", uri, err)
				return
			}

			if hash != cacheHash {
				log.Errorf("File %s hash mismatch!!! Expected hash: %s, hash, recieved from cache: %s.", uri, hash.String(), cacheHash.String())
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
