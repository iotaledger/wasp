package root

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot)

var (
	// Funcs
	FuncDeployContract = coreutil.NewEP3(Contract, "deployContract",
		coreutil.FieldWithCodec(codec.HashValue), coreutil.FieldWithCodec(codec.String), coreutil.FieldWithCodec(codec.Dict))

	FuncGrantDeployPermission = coreutil.NewEP1(Contract, "grantDeployPermission",
		coreutil.FieldWithCodec(codec.AgentID),
	)
	FuncRevokeDeployPermission = coreutil.NewEP1(Contract, "revokeDeployPermission",
		coreutil.FieldWithCodec(codec.AgentID),
	)
	FuncRequireDeployPermissions = coreutil.NewEP1(Contract, "requireDeployPermissions",
		coreutil.FieldWithCodec(codec.Bool),
	)

	// Views
	ViewFindContract = coreutil.NewViewEP12(Contract, "findContract",
		coreutil.FieldWithCodec(codec.Hname),
		coreutil.FieldWithCodec(codec.Bool),
		coreutil.FieldWithCodec(ContractRegistryCodec),
	)
	ViewGetContractRecords = coreutil.NewViewEP01(Contract, "getContractRecords",
		OutputContractRecords{},
	)
)

// state variables
const (
	varSchemaVersion            = "v" // covered in: TestDeployNativeContract
	varContractRegistry         = "r" // covered in: TestDeployNativeContract
	varDeployPermissionsEnabled = "a" // covered in: TestDeployNativeContract
	varDeployPermissions        = "p" // covered in: TestDeployNativeContract
)

// request parameters
const (
	ParamDeployer                 = "dp"
	ParamHname                    = "hn"
	ParamName                     = "nm"
	ParamProgramHash              = "ph"
	ParamContractRecData          = "dt"
	ParamContractFound            = "cf"
	ParamDeployPermissionsEnabled = "de"
)

var ContractRegistryCodec = codec.NewCodecEx(ContractRecordFromBytes)

type OutputContractRecords struct{}

func (c OutputContractRecords) Encode(recs map[isc.Hname]*ContractRecord) []byte {
	ret := dict.Dict{}
	dst := collections.NewMap(ret, varContractRegistry)
	for hname, rec := range recs {
		dst.SetAt(codec.Hname.Encode(hname), ContractRegistryCodec.Encode(rec))
	}
	return ret.Bytes()
}

func (c OutputContractRecords) Decode(d []byte) (map[isc.Hname]*ContractRecord, error) {
	ret := make(map[isc.Hname]*ContractRecord)
	data, err := dict.FromBytes(d)
	if err != nil {
		return nil, err
	}

	collections.NewMapReadOnly(data, varContractRegistry).Iterate(func(k []byte, v []byte) bool {
		ret[codec.Hname.MustDecode(k)] = ContractRegistryCodec.MustDecode(v)
		return true
	})
	return ret, nil
}
