package wasmclient

import "github.com/iotaledger/wasp/packages/kv/dict"

type ClientView struct {
	svc *Service
	err error
	res dict.Dict
}

func (v *ClientView) Call(viewName string, args *Arguments) {
	v.res, v.err = v.svc.CallView(viewName, args)
}

func (v *ClientView) Error() error {
	return v.err
}

func (v *ClientView) Results() Results {
	return Results{res: v.res}
}
