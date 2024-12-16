package testcore

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"github.com/samber/lo"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
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
var SupportedField = reflect.TypeOf(coreutil.Field[any](""))

func isSupportedType(p reflect.Type) bool {
	fmt.Println(getName(SupportedField))
	fmt.Println(getName(SupportedFieldOptional))
	fmt.Println(getName(p))

	if getName(SupportedField) == getName(p) {
		return true
	}

	if getName(SupportedFieldOptional) == getName(p) {
		return true
	}

	return false
}

func isOptional(p reflect.Type) bool {
	return getName(SupportedFieldOptional) == getName(p)
}

type CoreContractFunction struct {
	ContractName string
	FunctionName string
	IsView       bool
	InputArgs    []CompiledField
	OutputArgs   []CompiledField
}

type CompiledField struct {
	Name       string
	ArgIndex   int
	Type       string
	IsOptional bool
}

type CoreContractFunctionStructure interface {
	Inputs() []coreutil.FieldArg
	Outputs() []coreutil.FieldArg
	Hname() isc.Hname
	String() string
	ContractInfo() *coreutil.ContractInfo
	IsView() bool
}

func extractFields(fields []coreutil.FieldArg) []CompiledField {
	compiled := make([]CompiledField, len(fields))

	for i, input := range fields {
		fieldType := reflect.TypeOf(input)
		inputType := input.Type()

		typeStr := ""
		if inputType != nil {
			typeStr = inputType.String()
		} else {
			fmt.Printf("type is nil for %s", input.Name())
		}

		compiled[i] = CompiledField{
			ArgIndex:   i,
			Name:       input.Name(),
			IsOptional: isOptional(fieldType),
			Type:       typeStr,
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
	fset := token.NewFileSet()

	output := "var contractFuncs []CoreContractFunction = []CoreContractFunction{\n"

	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == "interface.go" {
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing %s: %v", path, err)
			}

			// Get package name
			packageName := file.Name.Name

			ast.Inspect(file, func(n ast.Node) bool {
				if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.VAR {
					for _, spec := range decl.Specs {
						if valueSpec, ok := spec.(*ast.ValueSpec); ok {
							for _, name := range valueSpec.Names {
								if strings.HasPrefix(name.Name, "Func") ||
									strings.HasPrefix(name.Name, "View") {
									output += fmt.Sprintf("\t\tconstructCoreContractFunction(&%s.%s),\n",
										packageName, name.Name)
								}
							}
						}
					}
				}
				return true
			})
		}
		return nil
	})

	output += "\t}"

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}

	fmt.Println(output)
}

func TestA(t *testing.T) {
	var contractFuncs []CoreContractFunction = []CoreContractFunction{
		constructCoreContractFunction(&accounts.FuncDeposit),
		constructCoreContractFunction(&accounts.FuncTransferAccountToChain),
		constructCoreContractFunction(&accounts.FuncTransferAllowanceTo),
		constructCoreContractFunction(&accounts.FuncWithdraw),
		constructCoreContractFunction(&accounts.ViewAccountObjects),
		constructCoreContractFunction(&accounts.ViewAccountObjectsInCollection),
		constructCoreContractFunction(&accounts.ViewBalance),
		constructCoreContractFunction(&accounts.ViewBalanceBaseToken),
		constructCoreContractFunction(&accounts.ViewBalanceBaseTokenEVM),
		constructCoreContractFunction(&accounts.ViewBalanceCoin),
		constructCoreContractFunction(&accounts.ViewGetAccountNonce),
		constructCoreContractFunction(&accounts.ViewObjectBCS),
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
		constructCoreContractFunction(&evm.FuncRegisterERC721NFTCollection),
		constructCoreContractFunction(&evm.FuncNewL1Deposit),
		constructCoreContractFunction(&evm.ViewGetChainID),
		constructCoreContractFunction(&governance.FuncRotateStateController),
		constructCoreContractFunction(&governance.FuncAddAllowedStateControllerAddress),
		constructCoreContractFunction(&governance.FuncRemoveAllowedStateControllerAddress),
		constructCoreContractFunction(&governance.ViewGetAllowedStateControllerAddresses),
		constructCoreContractFunction(&governance.FuncClaimChainOwnership),
		constructCoreContractFunction(&governance.FuncDelegateChainOwnership),
		constructCoreContractFunction(&governance.FuncSetPayoutAgentID),
		constructCoreContractFunction(&governance.FuncSetMinCommonAccountBalance),
		constructCoreContractFunction(&governance.ViewGetPayoutAgentID),
		constructCoreContractFunction(&governance.ViewGetMinCommonAccountBalance),
		constructCoreContractFunction(&governance.ViewGetChainOwner),
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
		constructCoreContractFunction(&sbtestsc.FuncEventLogGenericData),
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
		constructCoreContractFunction(&sbtestsc.FuncStackOverflow),
	}

	fmt.Println(string(lo.Must(json.Marshal(contractFuncs))))
}
