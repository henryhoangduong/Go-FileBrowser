package auth

import (
	"filebrowser/database/users"
	"net/http"
	"sync"
)

var (
	revokedApiKeyList map[string]bool
	revokeMu          sync.Mutex
)

type Auther interface {
	// Auth is called to authenticate a request.
	Auth(r *http.Request, userStore *users.Storage) (*users.User, error)
	// LoginPage indicates if this auther needs a login page.
	LoginPage() bool
}

func IsRevokedApiKey(key string) bool {
	_, exists := revokedApiKeyList[key]
	return exists
}

func RevokeAPIKey(key string) {
	revokeMu.Lock()
	delete(revokedApiKeyList, key)
	revokeMu.Unlock()
}
