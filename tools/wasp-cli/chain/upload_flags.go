package chain

import "github.com/spf13/pflag"

// TODO temporary with committee nodes. Find more user friendly way of uploading wasm blobs to the committee nodes
//  - upload to 1 or few and replicate between among nodes
//  - extend '*' option in RequestArgs with download options (web, IPFS)
var uploadNodes []int
var uploadQuorum int

func initUploadFlags(flags *pflag.FlagSet) {
	flags.IntSliceVarP(&uploadNodes, "upload-nodes", "", []int{0, 1, 2, 3}, "wasp nodes for blob upload")
	flags.IntVarP(&uploadQuorum, "upload-quorum", "", 3, "quorum for blob upload")
}
