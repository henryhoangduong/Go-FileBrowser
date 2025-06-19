package settings

import (
	"crypto/rand"
	"filebrowser/database/users"
)

const DefaultUsersHomeBasePath = "/users"

type AuthMethod string

func GenerateKey() ([]byte, error) {
	b := make([]byte, 64) //nolint:gomnd
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GetSettingsConfig(nameType string, Value string) string {
	return nameType + Value
}

func AdminPerms() users.Permissions {
	return users.Permissions{
		Modify: true,
		Share:  true,
		Admin:  true,
		Api:    true,
	}
}

func ApplyUserDefaults(u *users.User) {
	u.StickySidebar = Config.UserDefaults.StickySidebar
	u.DisableSettings = Config.UserDefaults.DisableSettings
	u.DarkMode = Config.UserDefaults.DarkMode
	u.Locale = Config.UserDefaults.Locale
	u.ViewMode = Config.UserDefaults.ViewMode
	u.SingleClick = Config.UserDefaults.SingleClick
	u.Permissions = Config.UserDefaults.Permissions
	u.Preview = Config.UserDefaults.Preview
	u.ShowHidden = Config.UserDefaults.ShowHidden
	u.DateFormat = Config.UserDefaults.DateFormat
	u.DisableOnlyOfficeExt = Config.UserDefaults.DisableOnlyOfficeExt
	u.ThemeColor = Config.UserDefaults.ThemeColor
	u.GallerySize = Config.UserDefaults.GallerySize
	u.QuickDownload = Config.UserDefaults.QuickDownload
	u.LockPassword = Config.UserDefaults.LockPassword
	if len(u.Scopes) == 0 {
		for _, source := range Config.Server.Sources {
			if source.Config.DefaultEnabled {
				u.Scopes = append(u.Scopes, users.SourceScope{
					Name:  source.Path, // backend name is path
					Scope: source.Config.DefaultUserScope,
				})
			}
		}
	}
}
