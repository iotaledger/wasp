package wasmclient

type ClientView struct {
	svc *Service
}

func (f *ClientView) Call(viewName string, args *Arguments) Results {
	return f.svc.CallView(viewName, args)
}
