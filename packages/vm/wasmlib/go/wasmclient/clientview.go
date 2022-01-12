// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

type ClientView struct {
	svc *Service
	err error
	res ResMap
}

func (v *ClientView) Call(viewName string, args *Arguments) {
	if args == nil {
		args = &Arguments{}
	}
	v.res, v.err = v.svc.CallView(viewName, args.args)
}

func (v *ClientView) Error() error {
	return v.err
}

func (v *ClientView) Results() Results {
	return Results{res: v.res}
}
