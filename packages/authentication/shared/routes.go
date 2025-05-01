// Package shared contains common utilities and components shared across authentication modules.
package shared

func AuthRoute() string {
	return "/auth"
}

func AuthInfoRoute() string {
	return "/auth/info"
}

type AuthInfoModel struct {
	Scheme  string `json:"scheme" swagger:"desc(Authentication scheme (jwt, basic, ip)),required"`
	AuthURL string `json:"authURL" swagger:"desc(JWT only),required"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" swagger:"required"`
	Password string `json:"password" form:"password" swagger:"required"`
}

type LoginResponse struct {
	JWT   string `json:"jwt,omitempty" swagger:"required"`
	Error error  `json:"error,omitempty" swagger:""`
}
