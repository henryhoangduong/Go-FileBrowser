package settings

import "filebrowser/common/errors"

type StorageBackend interface {
	Get() (*Settings, error)
	Save(*Settings) error
	GetServer() (*Server, error)
	SaveServer(*Server) error
}

type Storage struct {
	back StorageBackend
}

func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}
func (s *Storage) Get() (*Settings, error) {
	set, err := s.back.Get()
	if err != nil {
		return nil, err
	}
	return set, nil
}
func (s *Storage) Save(set *Settings) error {
	if len(set.Auth.Key) == 0 {
		return errors.ErrEmptyKey
	}

	if set.UserDefaults.Locale == "" {
		set.UserDefaults.Locale = "en"
	}

	if set.UserDefaults.ViewMode == "" {
		set.UserDefaults.ViewMode = "normal"
	}

	err := s.back.Save(set)
	if err != nil {
		return err
	}

	return nil
}
