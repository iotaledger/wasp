package processor

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util/sema"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"io/ioutil"
	"net/url"
	"path"
	"sync"
)

type processorInstance struct {
	vmtypes.Processor
	timedLock *sema.Lock
}

// TODO implement multiple workers/instances per program hash. Currently only one

var (
	processors      = make(map[string]processorInstance)
	processorsMutex sync.RWMutex
)

// LoadProcessorAsync creates and registers processor for program hash
// asynchronously
// possibly, locates Wasm program code in the file system, in IPFS etc
func LoadProcessorAsync(programHash string, onFinish func(err error)) {
	go func() {
		proc, err := loadProcessor(programHash)
		if err != nil {
			onFinish(err)
			return
		}

		processorsMutex.Lock()
		processors[programHash] = processorInstance{
			Processor: proc,
			timedLock: sema.New(),
		}
		processorsMutex.Unlock()

		onFinish(nil)
	}()
}

// loadProcessor creates processor instance
// first tries to resolve known program hashes used for testing
// then tries to create from the binary in the registry cache
// finally tries to load binary code from the location in the metadata
func loadProcessor(progHashStr string) (vmtypes.Processor, error) {
	proc, ok := examples.LoadProcessor(progHashStr)
	if ok {
		return proc, nil
	}
	progHash, err := hashing.HashValueFromBase58(progHashStr)
	binaryCode, exist, err := registry.GetProgramCode(&progHash)

	md, exist, err := registry.GetProgramMetadata(&progHash)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if exist {
		return vmtypes.FromBinaryCode(md.VMType, binaryCode)
	}

	if exist {
		binaryCode, err := loadBinaryCode(md.Location, &progHash)
		if err != nil {
			return nil, fmt.Errorf("failed to load program's binary data from location %s, program hash = %s",
				md.Location, progHashStr)
		}
		return vmtypes.FromBinaryCode(md.VMType, binaryCode)
	}
	return nil, fmt.Errorf("no metadata for program hash %s", progHashStr)
}

// loads binary code of the VM, possibly from remote location
// caches it into the the registry
func loadBinaryCode(location string, progHash *hashing.HashValue) ([]byte, error) {
	urlStruct, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	var data []byte
	switch urlStruct.Scheme {
	case "file":
		file := path.Join(parameters.GetString(parameters.VMBinaryDir), urlStruct.Host)
		if data, err = ioutil.ReadFile(file); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown Wasm binary location scheme '%s'", urlStruct.Scheme)
	}

	h := hashing.HashData(data)
	if *h != *progHash {
		return nil, fmt.Errorf("binary data or hash is not valid")
	}
	_, err = registry.SaveProgramCode(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
