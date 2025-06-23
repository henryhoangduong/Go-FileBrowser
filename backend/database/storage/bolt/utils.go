package bolt

import (
	"filebrowser/common/errors"

	storm "github.com/asdine/storm/v3"
)

func get(db *storm.DB, name string, to interface{}) error {
	err := db.Get("config", name, to)
	if err == storm.ErrNotFound {
		return errors.ErrNotExist
	}

	return err
}

func Save(db *storm.DB, name string, from interface{}) error {
	return db.Set("config", name, from)
}
