package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

func getName(p reflect.Type) string {
	// Regex to run
	r := regexp.MustCompile("([a-zA-Z]+)\\[")

	// Return capture
	str := p.String()
	matches := r.FindStringSubmatch(str)

	if len(matches) == 0 {
		// Probably not of type field/optionalField
		return str
	}

	return matches[1]
}

var SupportedFieldOptional = reflect.TypeOf(coreutil.FieldOptional[any](""))

func isOptional(p reflect.Type) bool {
	return getName(SupportedFieldOptional) == getName(p)
}

type AccountSettings struct {
	Foo string
	Bar uint64
	Baz uint32
}

func extractFields(fields []coreutil.FieldArg) []CompiledField {
	compiled := make([]CompiledField, len(fields))

	for i, input := range fields {
		fieldType := reflect.TypeOf(input)
		inputType := input.Type()

		compiled[i] = CompiledField{
			ArgIndex:   i,
			Name:       input.Name(),
			IsOptional: isOptional(fieldType),
			Type:       inputType,
			TypeName:   getName(inputType),
		}
	}

	return compiled
}

func constructCoreContractFunction(f CoreContractFunctionStructure) CoreContractFunction {
	return CoreContractFunction{
		ContractName: f.ContractInfo().Name,
		FunctionName: f.String(),
		IsView:       f.IsView(),

		InputArgs:  extractFields(f.Inputs()),
		OutputArgs: extractFields(f.Outputs()),
	}
}

func TestGenerateVariables(t *testing.T) {
	t.Skip()

	generateContractFuncs(t)
}

func TestBCSConversion(t *testing.T) {
	t.Skip()

	gen := NewTypeGenerator()

	fmt.Println(gen.GetOutput())
}

func TestB(t *testing.T) {
	t.Skip()

	gen := NewTypeGenerator()
	gen.GenerateFunction(constructCoreContractFunction(&blocklog.ViewGetEventsForRequest))

	t.Log(gen.GetOutput())
}

func TestDumpJson(t *testing.T) {
	t.Skip()

	gen := NewTypeGenerator()

	contractFuncs := []CoreContractFunction{
		constructCoreContractFunction(&accounts.FuncDeposit),
		constructCoreContractFunction(&accounts.FuncTransferAllowanceTo),
		constructCoreContractFunction(&accounts.FuncWithdraw),
		constructCoreContractFunction(&accounts.ViewAccountObjects),
		constructCoreContractFunction(&accounts.ViewBalance),
		constructCoreContractFunction(&accounts.ViewBalanceBaseToken),
		constructCoreContractFunction(&accounts.ViewBalanceBaseTokenEVM),
		constructCoreContractFunction(&accounts.ViewBalanceCoin),
		constructCoreContractFunction(&accounts.ViewGetAccountNonce),
		constructCoreContractFunction(&accounts.ViewTotalAssets),
		constructCoreContractFunction(&blocklog.ViewGetBlockInfo),
		constructCoreContractFunction(&blocklog.ViewGetRequestIDsForBlock),
		constructCoreContractFunction(&blocklog.ViewGetRequestReceipt),
		constructCoreContractFunction(&blocklog.ViewGetRequestReceiptsForBlock),
		constructCoreContractFunction(&blocklog.ViewIsRequestProcessed),
		constructCoreContractFunction(&blocklog.ViewGetEventsForRequest),
		constructCoreContractFunction(&blocklog.ViewGetEventsForBlock),
		constructCoreContractFunction(&errors.FuncRegisterError),
		constructCoreContractFunction(&errors.ViewGetErrorMessageFormat),
		constructCoreContractFunction(&evm.FuncSendTransaction),
		constructCoreContractFunction(&evm.FuncCallContract),
		constructCoreContractFunction(&evm.FuncRegisterERC20Coin),
		constructCoreContractFunction(&evm.FuncNewL1Deposit),
		constructCoreContractFunction(&evm.ViewGetChainID),
		constructCoreContractFunction(&governance.FuncClaimChainAdmin),
		constructCoreContractFunction(&governance.FuncDelegateChainAdmin),
		constructCoreContractFunction(&governance.FuncSetPayoutAgentID),
		constructCoreContractFunction(&governance.ViewGetPayoutAgentID),
		constructCoreContractFunction(&governance.ViewGetChainAdmin),
		constructCoreContractFunction(&governance.FuncSetFeePolicy),
		constructCoreContractFunction(&governance.FuncSetGasLimits),
		constructCoreContractFunction(&governance.ViewGetFeePolicy),
		constructCoreContractFunction(&governance.ViewGetGasLimits),
		constructCoreContractFunction(&governance.FuncSetEVMGasRatio),
		constructCoreContractFunction(&governance.ViewGetEVMGasRatio),
		constructCoreContractFunction(&governance.ViewGetChainInfo),
		constructCoreContractFunction(&governance.FuncAddCandidateNode),
		constructCoreContractFunction(&governance.FuncRevokeAccessNode),
		constructCoreContractFunction(&governance.FuncChangeAccessNodes),
		constructCoreContractFunction(&governance.ViewGetChainNodes),
		constructCoreContractFunction(&governance.FuncStartMaintenance),
		constructCoreContractFunction(&governance.FuncStopMaintenance),
		constructCoreContractFunction(&governance.ViewGetMaintenanceStatus),
		constructCoreContractFunction(&governance.FuncSetMetadata),
		constructCoreContractFunction(&governance.ViewGetMetadata),
		constructCoreContractFunction(&root.ViewFindContract),
		constructCoreContractFunction(&root.ViewGetContractRecords),
	}

	for _, c := range contractFuncs {
		gen.GenerateFunction(c)
	}

	types := contractFuncs

	s := lo.Must(json.MarshalIndent(types, "", "  "))

	fmt.Print(string(s))
}

func TestA(t *testing.T) {
	t.Skip()

	gen := NewTypeGenerator()

	contractFuncs := []CoreContractFunction{
		constructCoreContractFunction(&accounts.FuncDeposit),
		constructCoreContractFunction(&accounts.FuncTransferAllowanceTo),
		constructCoreContractFunction(&accounts.FuncWithdraw),
		constructCoreContractFunction(&accounts.ViewAccountObjects),
		constructCoreContractFunction(&accounts.ViewBalance),
		constructCoreContractFunction(&accounts.ViewBalanceBaseToken),
		constructCoreContractFunction(&accounts.ViewBalanceBaseTokenEVM),
		constructCoreContractFunction(&accounts.ViewBalanceCoin),
		constructCoreContractFunction(&accounts.ViewGetAccountNonce),
		constructCoreContractFunction(&accounts.ViewTotalAssets),
		constructCoreContractFunction(&blocklog.ViewGetBlockInfo),
		constructCoreContractFunction(&blocklog.ViewGetRequestIDsForBlock),
		constructCoreContractFunction(&blocklog.ViewGetRequestReceipt),
		constructCoreContractFunction(&blocklog.ViewGetRequestReceiptsForBlock),
		constructCoreContractFunction(&blocklog.ViewIsRequestProcessed),
		constructCoreContractFunction(&blocklog.ViewGetEventsForRequest),
		constructCoreContractFunction(&blocklog.ViewGetEventsForBlock),
		constructCoreContractFunction(&errors.FuncRegisterError),
		constructCoreContractFunction(&errors.ViewGetErrorMessageFormat),
		constructCoreContractFunction(&evm.FuncSendTransaction),
		constructCoreContractFunction(&evm.FuncCallContract),
		constructCoreContractFunction(&evm.FuncRegisterERC20Coin),
		constructCoreContractFunction(&evm.FuncNewL1Deposit),
		constructCoreContractFunction(&evm.ViewGetChainID),
		constructCoreContractFunction(&governance.FuncClaimChainAdmin),
		constructCoreContractFunction(&governance.FuncDelegateChainAdmin),
		constructCoreContractFunction(&governance.FuncSetPayoutAgentID),
		constructCoreContractFunction(&governance.ViewGetPayoutAgentID),
		constructCoreContractFunction(&governance.ViewGetGasCoinTargetValue),
		constructCoreContractFunction(&governance.FuncSetGasCoinTargetValue),
		constructCoreContractFunction(&governance.ViewGetChainAdmin),
		constructCoreContractFunction(&governance.FuncSetFeePolicy),
		constructCoreContractFunction(&governance.FuncSetGasLimits),
		constructCoreContractFunction(&governance.ViewGetFeePolicy),
		constructCoreContractFunction(&governance.ViewGetGasLimits),
		constructCoreContractFunction(&governance.FuncSetEVMGasRatio),
		constructCoreContractFunction(&governance.ViewGetEVMGasRatio),
		constructCoreContractFunction(&governance.ViewGetChainInfo),
		constructCoreContractFunction(&governance.FuncAddCandidateNode),
		constructCoreContractFunction(&governance.FuncRevokeAccessNode),
		constructCoreContractFunction(&governance.FuncChangeAccessNodes),
		constructCoreContractFunction(&governance.ViewGetChainNodes),
		constructCoreContractFunction(&governance.FuncStartMaintenance),
		constructCoreContractFunction(&governance.FuncStopMaintenance),
		constructCoreContractFunction(&governance.ViewGetMaintenanceStatus),
		constructCoreContractFunction(&governance.FuncSetMetadata),
		constructCoreContractFunction(&governance.ViewGetMetadata),
		constructCoreContractFunction(&root.ViewFindContract),
		constructCoreContractFunction(&root.ViewGetContractRecords),
		constructCoreContractFunction(&inccounter.FuncIncCounter),
		constructCoreContractFunction(&inccounter.FuncIncAndRepeatOnceAfter2s),
		constructCoreContractFunction(&inccounter.FuncIncAndRepeatMany),
		constructCoreContractFunction(&inccounter.ViewGetCounter),
		// For now ignore the sbtest contract (mainly used for testing the generator, not included in an actual lib)
		/*constructCoreContractFunction(&sbtestsc.FuncEventLogGenericData),
		constructCoreContractFunction(&sbtestsc.FuncEventLogEventData),
		constructCoreContractFunction(&sbtestsc.FuncEventLogDeploy),
		constructCoreContractFunction(&sbtestsc.FuncChainOwnerIDView),
		constructCoreContractFunction(&sbtestsc.FuncChainOwnerIDFull),
		constructCoreContractFunction(&sbtestsc.FuncCheckContextFromFullEP),
		constructCoreContractFunction(&sbtestsc.FuncCheckContextFromViewEP),
		constructCoreContractFunction(&sbtestsc.FuncTestCustomError),
		constructCoreContractFunction(&sbtestsc.FuncPanicFullEP),
		constructCoreContractFunction(&sbtestsc.FuncPanicViewEP),
		constructCoreContractFunction(&sbtestsc.FuncWithdrawFromChain),
		constructCoreContractFunction(&sbtestsc.FuncDoNothing),
		constructCoreContractFunction(&sbtestsc.FuncJustView),
		constructCoreContractFunction(&sbtestsc.FuncSetInt),
		constructCoreContractFunction(&sbtestsc.FuncGetInt),
		constructCoreContractFunction(&sbtestsc.FuncGetFibonacci),
		constructCoreContractFunction(&sbtestsc.FuncGetFibonacciIndirect),
		constructCoreContractFunction(&sbtestsc.FuncCalcFibonacciIndirectStoreValue),
		constructCoreContractFunction(&sbtestsc.FuncViewCalcFibonacciResult),
		constructCoreContractFunction(&sbtestsc.FuncGetCounter),
		constructCoreContractFunction(&sbtestsc.FuncIncCounter),
		constructCoreContractFunction(&sbtestsc.FuncSplitFunds),
		constructCoreContractFunction(&sbtestsc.FuncSplitFundsNativeTokens),
		constructCoreContractFunction(&sbtestsc.FuncPingAllowanceBack),
		constructCoreContractFunction(&sbtestsc.FuncSendLargeRequest),
		constructCoreContractFunction(&sbtestsc.FuncInfiniteLoop),
		constructCoreContractFunction(&sbtestsc.FuncInfiniteLoopView),
		constructCoreContractFunction(&sbtestsc.FuncSendNFTsBack),
		constructCoreContractFunction(&sbtestsc.FuncClaimAllowance),
		constructCoreContractFunction(&sbtestsc.FuncStackOverflow),*/
	}

	for _, c := range contractFuncs {
		gen.GenerateFunction(c)
	}

	out := gen.GetOutput()
	f, err := os.Create("D:/Coding/bcs_test/generated.ts")
	require.NoError(t, err)

	_, err = f.WriteString(out)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	fmt.Println(gen.GetOutput())
}

func TestTypes(t *testing.T) {
	t.Skip()

	objectRef := iotatest.RandomObjectRef()
	b := bcs.MustMarshal(objectRef)

	fmt.Println(b, hexutil.Encode(b))
	fmt.Println(objectRef)
}
