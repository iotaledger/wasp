package downloader

import (
	"fmt"
	"strings"
	"sync"

	"io/ioutil"
	"net/http"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

//Downloader struct to store currently being downloaded files and othe things.
type Downloader struct {
	log         *logger.Logger
	ipfsGateway string
	// downloads is just a set of strings. The value of the element is not important. The existence of key in the map is what counts.
	downloads      map[string]bool
	downloadsMutex sync.Mutex
}

var (
	defaultDownloader *Downloader
)

//Init initialises default downloader
func Init(log *logger.Logger, ipfsGateway string) {
	defaultDownloader = New(log, ipfsGateway)
}

//New is a downloader constructor
func New(log *logger.Logger, ipfsGateway string) *Downloader {
	return &Downloader{
		log:            log,
		ipfsGateway:    ipfsGateway,
		downloads:      make(map[string]bool),
		downloadsMutex: sync.Mutex{},
	}
}

//GetDefaultDownloader returns default downloader
func GetDefaultDownloader() *Downloader {
	return defaultDownloader
}

// DownloadAndStore downloads and stores data. Accepted URIs are:
// http://<url of the contents> (e.g. http://some.place.lt/some/contents.txt)
// https://<url of the contents> (e.g. https://some.place.lt/some/contents.txt)
// ipfs://<cid of the contents> (e.g. ipfs://QmeyMc1i9KLqqyqYCksDZiwntxwuiz5Z1hbLBrHvAXyjMZ)
func (d *Downloader) DownloadAndStore(hash hashing.HashValue, uri string, cache coretypes.BlobCache) error {
	if d.contains(uri) {
		d.log.Warnf("File %s is already being downloaded. Skipping it.", uri)
		return nil
	}

	d.markStarted(uri)
	go func() {
		defer d.markCompleted(uri)

		download, err := d.download(uri)

		if err != nil {
			d.log.Errorf("Error retrieving file %s: %s.", uri, err)
			return
		}

		var cacheHash hashing.HashValue
		cacheHash, err = cache.PutBlob(download)

		if err != nil {
			d.log.Errorf("Error putting file %s to cache: %s.", uri, err)
			return
		}

		if hash != cacheHash {
			d.log.Errorf("File %s hash mismatch!!! Expected hash: %s, hash, recieved from cache: %s.", uri, hash.String(), cacheHash.String())
			return
		}

	}()

	return nil
}

func (d *Downloader) contains(uri string) bool {
	d.downloadsMutex.Lock()
	defer d.downloadsMutex.Unlock()

	_, ok := d.downloads[uri]
	return ok
}

func (d *Downloader) markStarted(uri string) {
	d.downloadsMutex.Lock()
	defer d.downloadsMutex.Unlock()

	d.downloads[uri] = true
}

func (d *Downloader) markCompleted(uri string) {
	d.downloadsMutex.Lock()
	defer d.downloadsMutex.Unlock()

	delete(d.downloads, uri)
}

func (d *Downloader) download(uri string) ([]byte, error) {
	var split []string = strings.SplitN(uri, "://", 2)
	if len(split) != 2 {
		return nil, fmt.Errorf("file uri %s is invalid", uri)
	}

	var protocol string = split[0]
	var path string = split[1]
	switch protocol {
	case "ipfs":
		return d.donwloadFromHTTP(d.ipfsGateway + "/ipfs/" + path)
	case "http":
		return d.donwloadFromHTTP(uri)
	case "https":
		return d.donwloadFromHTTP(uri)
	default:
		return nil, fmt.Errorf("unknown protocol %s of uri %s", protocol, uri)
	}
}

func (*Downloader) donwloadFromHTTP(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
