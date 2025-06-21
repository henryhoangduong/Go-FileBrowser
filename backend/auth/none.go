package auth

import (
	"filebrowser/database/users"
	"net/http"
)

const MethodNoAuth = "noauth"

type NoAuth struct{}

func (a NoAuth) Auth(r *http.Request, usr *users.Storage) (*users.User, error) {
	return usr.Get(uint(1))
}
func (a NoAuth) LoginPage() bool {
	return false
}
