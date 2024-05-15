package root

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot)

var (
	// Funcs
	FuncDeployContract        = EPDeployContract{EntryPointInfo: Contract.Func("deployContract")}
	FuncGrantDeployPermission = coreutil.NewEP1(Contract, "grantDeployPermission",
		coreutil.FieldWithCodec(ParamDeployer, codec.AgentID),
	)
	FuncRevokeDeployPermission = coreutil.NewEP1(Contract, "revokeDeployPermission",
		coreutil.FieldWithCodec(ParamDeployer, codec.AgentID),
	)
	FuncRequireDeployPermissions = coreutil.NewEP1(Contract, "requireDeployPermissions",
		coreutil.FieldWithCodec(ParamDeployPermissionsEnabled, codec.Bool),
	)

	// Views
	ViewFindContract = coreutil.NewViewEP12(Contract, "findContract",
		coreutil.FieldWithCodec(ParamHname, codec.Hname),
		coreutil.FieldWithCodec(ParamContractFound, codec.Bool),
		coreutil.FieldWithCodec(ParamContractRecData, ContractRegistryCodec),
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

type EPDeployContract struct {
	coreutil.EntryPointInfo[isc.Sandbox]
}

func (e EPDeployContract) Message(name string, programHash hashing.HashValue, initParams dict.Dict) isc.Message {
	d := initParams.Clone()
	d[ParamProgramHash] = codec.HashValue.Encode(programHash)
	d[ParamName] = codec.String.Encode(name)
	return e.EntryPointInfo.Message(d)
}

type OutputContractRecords struct{}

func (c OutputContractRecords) Encode(recs map[isc.Hname]*ContractRecord) dict.Dict {
	ret := dict.Dict{}
	dst := collections.NewMap(ret, varContractRegistry)
	for hname, rec := range recs {
		dst.SetAt(codec.Hname.Encode(hname), ContractRegistryCodec.Encode(rec))
	}
	return ret
}

func (c OutputContractRecords) Decode(d dict.Dict) (map[isc.Hname]*ContractRecord, error) {
	ret := make(map[isc.Hname]*ContractRecord)
	collections.NewMapReadOnly(d, varContractRegistry).Iterate(func(k []byte, v []byte) bool {
		ret[codec.Hname.MustDecode(k)] = ContractRegistryCodec.MustDecode(v)
		return true
	})
	return ret, nil
}
