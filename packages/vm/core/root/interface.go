package root

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot)

var (
	// Funcs
	FuncDeployContract = coreutil.NewEP3(Contract, "deployContract",
		coreutil.FieldWithCodec(codec.HashValue),
		coreutil.FieldWithCodec(codec.String),
		coreutil.FieldWithCodec(codec.CallArguments),
	)

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
		coreutil.FieldWithCodec(codec.NewCodecEx(ContractRecordFromBytes)),
	)
	ViewGetContractRecords = coreutil.NewViewEP01(Contract, "getContractRecords",
		coreutil.FieldArrayWithCodec(codec.NewTupleCodec[isc.Hname, ContractRecord]()),
	)
)

// state variables
const (
	varSchemaVersion            = "v" // covered in: TestDeployNativeContract
	varContractRegistry         = "r" // covered in: TestDeployNativeContract
	varDeployPermissionsEnabled = "a" // covered in: TestDeployNativeContract
	varDeployPermissions        = "p" // covered in: TestDeployNativeContract
)
