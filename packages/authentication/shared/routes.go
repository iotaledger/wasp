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
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type LoginResponse struct {
	JWT string `json:"jwt,omitempty"`
}
