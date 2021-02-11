package chain

import "github.com/spf13/pflag"

var uploadQuorum int

func initUploadFlags(flags *pflag.FlagSet) {
	flags.IntVarP(&uploadQuorum, "upload-quorum", "", 3, "quorum for blob upload")
}
