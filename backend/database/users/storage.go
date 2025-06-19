package users

type StorageBackend interface {
	GetBy(interface{}) (*User, error)
	Gets() ([]*User, error)
	Save(u *User, changePass bool, disableScopeChange bool) error
	Update(u *User, adminActor bool, fields ...string) error
	DeleteByID(uint) error
	DeleteByUsername(string) error
}
type Store interface {
	Get(id interface{}) (user *User, err error)
	Gets() ([]*User, error)
	Update(user *User, adminActor bool, fields ...string) error
	Save(user *User, changePass bool, disableScopeChange bool) error
	Delete(id interface{}) error
	LastUpdate(id uint) int64
	AddApiKey(username uint, name string, key AuthToken) error
	DeleteApiKey(username uint, name string) error
}

func NewStorage(back StorageBackend) *Storage {
	return &Storage{
		back:    back,
		updated: map[uint]int64{},
	}
}
func (s *Storage) Get(id interface{}) (user *User, err error) {
	user, err = s.back.GetBy(id)
	if err != nil {
		return
	}
	return user, err
}
func (s *Storage) Gets() ([]*User, error) {
	users, err := s.back.Gets()
	if err != nil {
		return nil, err
	}
	return users, err
}
