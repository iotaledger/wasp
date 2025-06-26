package chain

import (
	"encoding/json"
	"io"
	"net/http"

	"fortio.org/safecast"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) getMempoolContents(e echo.Context) error {
	controllerutils.SetOperation(e, "get_mempool_contents")
	ch, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	pr, pw := io.Pipe()

	// TODO: This makes unprotected concurrent access to the MP state.
	go func() {
		defer pw.Close()

		ch.IterateMempool(func(req isc.Request) bool {
			jsonData, err := json.Marshal(models.RequestToJSONObject(req))
			if err != nil {
				return false
			}

			val, err := safecast.Convert[uint32](len(jsonData))
			if err != nil {
				return false
			}

			if _, err = pw.Write(codec.Encode[uint32](val)); err != nil {
				return false
			}

			if _, err = pw.Write(jsonData); err != nil {
				return false
			}

			return true
		})
	}()

	return e.Stream(http.StatusOK, "application/octet-stream", pr)
}
