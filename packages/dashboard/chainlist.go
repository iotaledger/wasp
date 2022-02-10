// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chainlist.tmpl
var tplChainList string

func (d *Dashboard) initChainList(e *echo.Echo, r renderer) Tab {
	route := e.GET("/chains", d.handleChainList)
	route.Name = "chainList"

	r[route.Path] = d.makeTemplate(e, tplChainList)

	return Tab{
		Path:  route.Path,
		Title: "Chains",
		Href:  route.Path,
	}
}

func (d *Dashboard) handleChainList(c echo.Context) error {
	chains, err := d.fetchChains()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, c.Path(), &ChainListTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		Chains:             chains,
	})
}

func (d *Dashboard) fetchChains() ([]*ChainOverview, error) {
	crs, err := d.wasp.GetChainRecords()
	if err != nil {
		return nil, err
	}
	r := make([]*ChainOverview, len(crs))
	for i, cr := range crs {
		rootInfo, err := d.fetchRootInfo(cr.ChainID)
		if err != nil {
			return nil, err
		}
		cmtInfo, err := d.wasp.GetChainCommitteeInfo(cr.ChainID)
		if err != nil {
			return nil, err
		}
		var cmtSize int
		if cmtInfo == nil {
			cmtSize = -1
		} else {
			cmtSize = len(cmtInfo.PeerStatus)
		}
		r[i] = &ChainOverview{
			ChainID:       cr.ChainID,
			Active:        cr.Active,
			RootInfo:      rootInfo,
			CommitteeSize: cmtSize,
			Error:         err,
		}
	}
	return r, nil
}

type ChainListTemplateParams struct {
	BaseTemplateParams
	Chains []*ChainOverview
}

type ChainOverview struct {
	ChainID       *iscp.ChainID
	Active        bool
	RootInfo      RootInfo
	CommitteeSize int
	Error         error
}
