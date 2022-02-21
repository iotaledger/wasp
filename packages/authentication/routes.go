package authentication

func AuthRoute() string {
	return "/auth"
}

func AuthStatusRoute() string {
	return "/auth/status"
}

func AuthTestRoute() string {
	return "/auth/test"
}

type AuthStatusModel struct {
	Scheme  string `swagger:"desc(Authentication scheme (jwt, basic, ip))"`
	AuthURL string `swagger:"desc(JWT only)"`
}
