// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared"
)

//go:embed templates/error.tmpl
var tplError string

type ErrorTemplateParams struct {
	BaseTemplateParams
	Code       int
	StatusText string
	Message    string
}

const errorTplName = "_error"

func (d *Dashboard) errorInit(e *echo.Echo, r renderer) {
	r[errorTplName] = d.makeTemplate(e, tplError)

	e.HTTPErrorHandler = d.handleError
}

// same as https://github.com/labstack/echo/blob/151ed6b3f150163352985448b5630ab69de40aa5/echo.go#L347
// but renders HTML instead of json
func (d *Dashboard) handleError(err error, c echo.Context) {
	he, ok := err.(*echo.HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err,
		}
	}

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(he.Code)
		} else {
			authContext, ok := c.Get("auth").(*authentication.AuthContext)

			if ok && authContext.Scheme() == authentication.AuthJWT && he.Code == http.StatusUnauthorized {
				err = d.redirect(c, shared.AuthRoute())
			} else {
				err = c.Render(he.Code, errorTplName, &ErrorTemplateParams{
					BaseTemplateParams: d.BaseParams(c),
					Code:               he.Code,
					StatusText:         http.StatusText(he.Code),
					Message:            fmt.Sprintf("%s", he.Message),
				})
			}
		}
		if err != nil {
			c.Echo().Logger.Error(err)
		}
	}
}
