package auth

import (
	"filebrowser/common/errors"
	"filebrowser/database/users"
	"net/http"
	"os"
)

const MethodProxyAuth = "proxy"

type ProxyAuth struct {
	Header string `json:"header"`
}

func (a ProxyAuth) Auth(r *http.Request, usr *users.Storage) (*users.User, error) {
	username := r.Header.Get(a.Header)
	user, err := usr.Get(username)
	if err == errors.ErrNotExist {
		return nil, os.ErrPermission
	}

	return user, err
}
func (a ProxyAuth) LoginPage() bool {
	return false
}
