package bolt

import (
	"filebrowser/auth"
	"filebrowser/common/settings"
	"filebrowser/database/share"
	"filebrowser/database/users"

	"github.com/asdine/storm/v3"
)

func NewStorage(db *storm.DB) (*auth.Storage, *users.Storage, *share.Storage, *settings.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore, err := auth.NewStorage(authBackend{db: db}, userStore)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return authStore, userStore, shareStore, settingsStore, nil
}
