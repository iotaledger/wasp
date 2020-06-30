// package redirects particular calls to Goshimmer
package redirect

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"net/http"
)

func HandleRedirectGetAddressOutputs(c echo.Context) error {
	addr, err := address.FromBase58(c.Param("address"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &misc.SimpleResponse{Error: err.Error()})
	}
	nodeLocation := config.Node.GetString(nodeconn.CfgNodeAPIBind)
	url := fmt.Sprintf("http://%s/utxodb/outputs/%s", nodeLocation, addr.String())
	fmt.Printf("+++++++++++++++++ HandleRedirectGetAddressOutputs: %s\n", url)
	return c.Redirect(http.StatusOK, url)
}

func HandleRedirectPostTransaction(c echo.Context) error {
	nodeLocation := config.Node.GetString(nodeconn.CfgNodeAPIBind)
	url := fmt.Sprintf("http://%s/utxodb/tx", nodeLocation)
	fmt.Printf("+++++++++++++++++ HandleRedirectPostTransaction: %s\n", url)
	return c.Redirect(http.StatusOK, url)
}
