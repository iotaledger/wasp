package webapi

import (
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

// ApiGroupModifier is required as it is impossible with echoSwagger to define a group with mixed authentication rules.
// Most of our routes are protected but very few are not.
//
// While it is possible to create two different groups such as (chainAdm, chainPub),
// this will pollute the code generation and the documentation itself,
// as it will create empty groups for controllers that define no public routes,
// duplicate code files and increase the client lib size even further
//
// Furthermore, it's forbidden to create two groups with the same name to support two different authentication rules.
// This wrapper sets a configurable modifier for each route that is assigned to this Group modifier struct.
// See: api.go -> loadControllers
type ApiGroupModifier struct {
	group           echoswagger.ApiGroup
	OverrideHandler func(api echoswagger.Api)
}

func (p *ApiGroupModifier) CallOverrideHandler(api echoswagger.Api) echoswagger.Api {
	if p.OverrideHandler != nil {
		p.OverrideHandler(api)
	}
	return api
}

func (p *ApiGroupModifier) Add(method, path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.Add(method, path, h, m...)

	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.GET(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.POST(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.PUT(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.DELETE(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.OPTIONS(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.HEAD(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.PATCH(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *ApiGroupModifier) SetDescription(desc string) echoswagger.ApiGroup {
	return p.group.SetDescription(desc)
}

func (p *ApiGroupModifier) SetExternalDocs(desc, url string) echoswagger.ApiGroup {
	return p.group.SetExternalDocs(desc, url)
}

func (p *ApiGroupModifier) SetSecurity(names ...string) echoswagger.ApiGroup {
	return p.group.SetSecurity(names...)
}

func (p *ApiGroupModifier) SetSecurityWithScope(s map[string][]string) echoswagger.ApiGroup {
	return p.group.SetSecurityWithScope(s)
}

func (p *ApiGroupModifier) EchoGroup() *echo.Group {
	return p.group.EchoGroup()
}
