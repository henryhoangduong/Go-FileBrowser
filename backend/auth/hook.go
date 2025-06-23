package auth

import (
	"encoding/json"
	"filebrowser/common/errors"
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
func (a *HookAuth) SaveUser() (*users.User, error) {
	u, err := a.Users.Get(a.Cred.Username)
	if err != nil && err != errors.ErrNotExist {
		return nil, err
	}

	if u == nil {
		// create user with the provided credentials
		d := &users.User{
			NonAdminEditable: users.NonAdminEditable{
				Password:    a.Cred.Password,
				Locale:      a.Settings.UserDefaults.Locale,
				ViewMode:    a.Settings.UserDefaults.ViewMode,
				SingleClick: a.Settings.UserDefaults.SingleClick,
				ShowHidden:  a.Settings.UserDefaults.ShowHidden,
			},
			Username:    a.Cred.Username,
			Permissions: a.Settings.UserDefaults.Permissions,
		}
		u = a.GetUser(d)

		err = a.Users.Save(u, false, false)
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	err = users.CheckPwd(a.Cred.Password, u.Password)
	if err != nil {
		return nil, err
	}

	if len(a.Fields.Values) > 1 {
		u = a.GetUser(u)
		// update user with provided fields
		err := a.Users.Update(u, u.Permissions.Admin)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}
func (a *HookAuth) LoginPage() bool {
	return true
}
func (a *HookAuth) GetUser(d *users.User) *users.User {
	// adds all permissions when user is admin
	isAdmin := d.Permissions.Admin
	perms := users.Permissions{
		Admin:  isAdmin,
		Modify: isAdmin || d.Permissions.Modify,
		Share:  isAdmin || d.Permissions.Share,
	}
	user := users.User{
		NonAdminEditable: users.NonAdminEditable{
			Password:    d.Password,
			Locale:      a.Fields.GetString("user.locale", d.Locale),
			ViewMode:    a.Fields.GetString("user.viewMode", d.ViewMode),
			SingleClick: a.Fields.GetBoolean("user.singleClick", d.SingleClick),
			ShowHidden:  a.Fields.GetBoolean("user.showHidden", d.ShowHidden),
		},
		ID:           d.ID,
		Username:     d.Username,
		Scopes:       d.Scopes,
		Permissions:  perms,
		LockPassword: true,
	}

	return &user
}
