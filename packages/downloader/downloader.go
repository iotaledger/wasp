package downloader

import (
	"strings"
	"sync"

	"io/ioutil"
	"net/http"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

type Downloader struct {
	log            *logger.Logger
	ipfsGateway    string
	downloads      map[string]bool // Just a HashSet. The value of the element is not important. The existence of key in the map is what counts.
	downloadsMutex *sync.Mutex
}

func New(log *logger.Logger, ipfsGateway string) *Downloader {
	return &Downloader{
		log:            log,
		ipfsGateway:    ipfsGateway,
		downloads:      make(map[string]bool),
		downloadsMutex: &sync.Mutex{},
	}
}

// Accepted URIs:
//  * http://<url of the contents>
//          e.g. http://some.place.lt/some/contents.txt
//  * https://<url of the contents>
//          e.g. https://some.place.lt/some/contents.txt
//  * ipfs://<cid of the contents>
//          e.g. ipfs://QmeyMc1i9KLqqyqYCksDZiwntxwuiz5Z1hbLBrHvAXyjMZ
func (this *Downloader) DownloadAndStore(hash hashing.HashValue, uri string, cache coretypes.BlobCache) error {
	if this.contains(uri) {
		this.log.Warnf("File %s is already being downloaded. Skipping it.", uri)
		return nil
	} else {
		this.markStarted(uri)
		go func() {
			defer this.markCompleted(uri)

			var split []string = strings.SplitN(uri, "://", 2)
			if len(split) != 2 {
				this.log.Errorf("File uri %s is invalid.", uri)
				return
			}

			var protocol string = split[0]
			var path string = split[1]
			var download []byte
			var err error
			switch protocol {
			case "ipfs":
				download, err = this.DonwloadFromHttp(this.ipfsGateway + "/ipfs/" + path)
			case "http":
				download, err = this.DonwloadFromHttp(uri)
			case "https":
				download, err = this.DonwloadFromHttp(uri)
			default:
				this.log.Errorf("Unknown protocol %s of uri %s.", protocol, uri)
				return
			}

			if err != nil {
				this.log.Errorf("Error retrieving file %s: %s.", uri, err)
				return
			}

			var cacheHash hashing.HashValue
			cacheHash, err = cache.PutBlob(download)

			if err != nil {
				this.log.Errorf("Error putting file %s to cache: %s.", uri, err)
				return
			}

			if hash != cacheHash {
				this.log.Errorf("File %s hash mismatch!!! Expected hash: %s, hash, recieved from cache: %s.", uri, hash.String(), cacheHash.String())
				return
			}

		}()

		return nil
	}
}

func (this *Downloader) contains(uri string) bool {
	var ok bool
	this.downloadsMutex.Lock()
	_, ok = this.downloads[uri]
	this.downloadsMutex.Unlock()
	return ok
}

func (this *Downloader) markStarted(uri string) {
	this.downloadsMutex.Lock()
	this.downloads[uri] = true
	this.downloadsMutex.Unlock()
}

func (this *Downloader) markCompleted(uri string) {
	this.downloadsMutex.Lock()
	delete(this.downloads, uri)
	this.downloadsMutex.Unlock()
}

func (this *Downloader) DonwloadFromHttp(url string) ([]byte, error) {
	var response *http.Response
	var err error
	response, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
