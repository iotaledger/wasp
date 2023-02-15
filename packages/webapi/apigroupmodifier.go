package webapi

import (
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

// APIGroupModifier is required as it is impossible with echoSwagger to define a group with mixed authentication rules.
// Most of our routes are protected but very few are not.
//
// While it is possible to create two different groups such as (chainAdm, chainPub),
// this will pollute the code generation and the documentation itself,
// as it will create empty groups for controllers that define no public routes,
// duplicate code files and increase the client lib size even further
//
// Furthermore, it's forbidden to create two groups with the same name to support two different authentication rules.
// This wrapper adds modifiers to each route that it is assigned to.
// See: api.go -> loadControllers
type APIGroupModifier struct {
	group           echoswagger.ApiGroup
	OverrideHandler func(api echoswagger.Api)
}

func (p *APIGroupModifier) CallOverrideHandler(api echoswagger.Api) echoswagger.Api {
	if p.OverrideHandler != nil {
		p.OverrideHandler(api)
	}
	return api
}

func (p *APIGroupModifier) Add(method, path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.Add(method, path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.GET(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.POST(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.PUT(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.DELETE(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.OPTIONS(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.HEAD(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echoswagger.Api {
	wrap := p.group.PATCH(path, h, m...)
	return p.CallOverrideHandler(wrap)
}

func (p *APIGroupModifier) SetDescription(desc string) echoswagger.ApiGroup {
	return p.group.SetDescription(desc)
}

func (p *APIGroupModifier) SetExternalDocs(desc, url string) echoswagger.ApiGroup {
	return p.group.SetExternalDocs(desc, url)
}

func (p *APIGroupModifier) SetSecurity(names ...string) echoswagger.ApiGroup {
	return p.group.SetSecurity(names...)
}

func (p *APIGroupModifier) SetSecurityWithScope(s map[string][]string) echoswagger.ApiGroup {
	return p.group.SetSecurityWithScope(s)
}

func (p *APIGroupModifier) EchoGroup() *echo.Group {
	return p.group.EchoGroup()
}
