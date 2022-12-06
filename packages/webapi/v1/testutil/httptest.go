package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func buildRequest(t *testing.T, method string, body interface{}) *http.Request {
	if body == nil {
		httptest.NewRequest(method, "/", nil)
	}

	if bodybytes, ok := body.([]byte); ok {
		req := httptest.NewRequest(method, "/", bytes.NewReader(bodybytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
		return req
	}

	if bodymap, ok := body.(map[string]string); ok {
		f := make(url.Values)
		for k, v := range bodymap {
			f.Set(k, v)
		}
		req := httptest.NewRequest(method, "/", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		return req
	}

	dataJSON, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(method, "/", bytes.NewReader(dataJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}

func CallWebAPIRequestHandler(
	t *testing.T,
	handler echo.HandlerFunc,
	method string,
	route string,
	params map[string]string,
	body interface{},
	res interface{},
	exptectedStatus int,
) {
	e := echo.New()

	req := buildRequest(t, method, body)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath(route)

	paramNames := make([]string, 0)
	paramValues := make([]string, 0)
	for k, v := range params {
		paramNames = append(paramNames, k)
		paramValues = append(paramValues, v)
	}
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)

	err := handler(c)
	if exptectedStatus >= 400 {
		require.Error(t, err)
		require.Equal(t, exptectedStatus, err.(*httperrors.HTTPError).Code)
	} else {
		require.NoError(t, err)
		require.Equal(t, exptectedStatus, rec.Code)
	}

	if res != nil {
		err = json.Unmarshal(rec.Body.Bytes(), res)
		require.NoError(t, err)
	}
}

func CallHTMLRequestHandler(t *testing.T, e *echo.Echo, handler echo.HandlerFunc, route string, params map[string]string) *goquery.Document {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath(route)

	paramNames := make([]string, 0)
	paramValues := make([]string, 0)
	for k, v := range params {
		paramNames = append(paramNames, k)
		paramValues = append(paramValues, v)
	}
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)

	err := handler(c)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)

	doc, err := goquery.NewDocumentFromReader(rec.Body)
	require.NoError(t, err)

	_, err = doc.Html()
	require.NoError(t, err)

	return doc
}
