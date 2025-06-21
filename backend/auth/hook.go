package auth

import (
	"encoding/json"
	"filebrowser/common/settings"
	"filebrowser/database/users"
	"fmt"
	"net/http"
	"os"

	"github.com/gtsteffaniak/go-logger/logger"
)

type hookCred struct {
	Password string `json:"password"`
	Username string `json:"username"`
}
type HookAuth struct {
	Users    users.Store        `json:"-"`
	Settings *settings.Settings `json:"-"`
	Server   *settings.Server   `json:"-"`
	Cred     hookCred           `json:"-"`
	Fields   hookFields         `json:"-"`
	Command  string             `json:"command"`
}
type hookFields struct {
	Values map[string]string
}

func (a *HookAuth) Auth(r *http.Request, usr *users.Storage) (*users.User, error) {
	var cred hookCred
	if r.Body == nil {
		return nil, os.ErrPermission
	}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		logger.Error("decode body error")
		return nil, os.ErrPermission
	}

	a.Users = usr
	a.Settings = &settings.Config
	a.Server = &settings.Config.Server
	a.Cred = cred

	action, err := a.RunCommand()
	if err != nil {
		return nil, err
	}
	logger.Debugf("hook auth %v", action)

	switch action {
	case "auth":
		u, err := a.SaveUser()
		if err != nil {
			return nil, err
		}
		return u, nil
	case "block":
		logger.Error("block error")

		return nil, os.ErrPermission
	case "pass":
		logger.Error("pass error")

		u, err := a.Users.Get(a.Cred.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to get user from store: %v", err)
		}
		err = users.CheckPwd(cred.Password, u.Password)
		if err != nil {
			return nil, err
		}
		return u, nil
	default:
		return nil, fmt.Errorf("invalid hook action: %s", action)
	}
}
func (a *HookAuth) LoginPage() bool {
	return true
}
