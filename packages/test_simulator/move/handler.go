// Package move implements Move contract execution for the L1 simulator.
package move

import (
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var frameworkModules = map[string]map[string]l1.MoveCallFunc{
	"coin":     CoinHandlers,
	"transfer": TransferHandlers,
	"borrow":   BorrowHandlers,
}

var stdlibModules = map[string]map[string]l1.MoveCallFunc{
	"option": OptionHandlers,
}

var iscModules = map[string]map[string]l1.MoveCallFunc{
	iscmove.AnchorModuleName:    AnchorHandlers,
	iscmove.RequestModuleName:   RequestHandlers,
	iscmove.AssetsBagModuleName: AssetsBagHandlers,
}

type CompositeHandler struct{}

func NewCompositeHandler(_ iotago.PackageID) *CompositeHandler {
	return &CompositeHandler{}
}

func (h *CompositeHandler) ExecuteMoveCall(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if call.Package == nil {
		return nil, fmt.Errorf("nil package in MoveCall")
	}

	pkgID := *call.Package
	var modules map[string]map[string]l1.MoveCallFunc

	switch pkgID {
	case *iotago.IotaPackageIDMoveStdlib:
		modules = stdlibModules
	case *iotago.IotaPackageIDIotaFramework:
		modules = frameworkModules
	case *iotago.IotaPackageIDIotaSystem:
		return nil, fmt.Errorf("unsupported iota-system call: %s::%s", call.Module, call.Function)
	default:
		modules = iscModules
	}

	funcs, ok := modules[call.Module]
	if !ok {
		// For unknown modules in user-published packages, try the generic coin mint handler.
		// Any published coin module follows the pattern: mint(treasury_cap, amount, recipient, ctx).
		if call.Function == "mint" {
			return genericCoinMint(ctx, call, args)
		}
		return nil, fmt.Errorf("unknown module: %s::%s", call.Module, call.Function)
	}

	fn, ok := funcs[call.Function]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s::%s", call.Module, call.Function)
	}

	return fn(ctx, call, args)
}

// genericCoinMint handles mint(treasury_cap, amount, recipient) for user-published coin modules.
// The coin type is derived from the package ID and module name: <pkgID>::<module>::<MODULE>.
func genericCoinMint(ctx *l1.CallContext, call *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("%s::mint requires 3 arguments (treasury_cap, amount, recipient)", call.Module)
	}

	amount, err := extractUint64(args[1])
	if err != nil {
		return nil, fmt.Errorf("%s::mint: amount: %w", call.Module, err)
	}

	recipient, err := extractAddress(args[2])
	if err != nil {
		return nil, fmt.Errorf("%s::mint: recipient: %w", call.Module, err)
	}

	coinType := fmt.Sprintf("%s::%s::%s", call.Package.String(), call.Module, strings.ToUpper(call.Module))

	coinID := ctx.FreshID()
	coinObj := createCoinObject(ctx, coinID, coinType, amount)
	coinObj.Owner = l1.SimOwner{AddressOwner: &recipient}
	ctx.Store.Put(coinObj)

	return nil, nil
}

var _ l1.MoveCallHandler = (*CompositeHandler)(nil)
