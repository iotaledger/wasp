package shared

func AuthRoute() string {
	return "/auth"
}

func AuthRouteSuccess() string {
	return "/auth/success"
}

func AuthInfoRoute() string {
	return "/auth/info"
}

type AuthInfoModel struct {
	Scheme  string `swagger:"desc(Authentication scheme (jwt, basic, ip))"`
	AuthURL string `swagger:"desc(JWT only)"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type LoginResponse struct {
	JWT   string `json:"jwt,omitempty"`
	Error error  `json:"error,omitempty"`
}
