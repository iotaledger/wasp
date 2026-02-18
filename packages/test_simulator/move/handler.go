package move

import (
	"fmt"

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

// CompositeHandler dispatches MoveCall commands to the appropriate handler.
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

	switch {
	case pkgID == *iotago.IotaPackageIDMoveStdlib:
		modules = stdlibModules
	case pkgID == *iotago.IotaPackageIDIotaFramework:
		modules = frameworkModules
	case pkgID == *iotago.IotaPackageIDIotaSystem:
		return nil, fmt.Errorf("unsupported iota-system call: %s::%s", call.Module, call.Function)
	default:
		modules = iscModules
	}

	funcs, ok := modules[call.Module]
	if !ok {
		return nil, fmt.Errorf("unknown module: %s::%s", call.Module, call.Function)
	}

	fn, ok := funcs[call.Function]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s::%s", call.Module, call.Function)
	}

	return fn(ctx, call, args)
}

var _ l1.MoveCallHandler = (*CompositeHandler)(nil)
