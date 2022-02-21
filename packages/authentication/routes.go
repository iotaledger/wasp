package authentication

func AuthRoute() string {
	return "/auth"
}

func AuthRouteSuccess() string {
	return "/auth/success"
}

func AuthStatusRoute() string {
	return "/auth/status"
}

type AuthStatusModel struct {
	Scheme  string `swagger:"desc(Authentication scheme (jwt, basic, ip))"`
	AuthURL string `swagger:"desc(JWT only)"`
}
