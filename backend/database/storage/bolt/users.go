package bolt

import (
	"filebrowser/adapters/fs/files"
	"filebrowser/auth"
	"filebrowser/common/errors"
	"filebrowser/common/settings"
	"filebrowser/common/utils"
	"filebrowser/database/users"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/gtsteffaniak/go-logger/logger"
)

type usersBackend struct {
	db *storm.DB
}

func (st usersBackend) GetBy(i interface{}) (user *users.User, err error) {
	user = &users.User{}

	var arg string
	var val interface{}
	switch i := i.(type) {
	case uint:
		val = i
		arg = "ID"
	case int:
		val = uint(i)
		arg = "ID"
	case string:
		arg = "Username"
		val = i
	default:
		return nil, errors.ErrInvalidDataType
	}

	err = st.db.One(arg, val, user)

	if err != nil {
		if err == storm.ErrNotFound {
			return nil, errors.ErrNotExist
		}
		return nil, err
	}

	return
}
func (st usersBackend) Gets() ([]*users.User, error) {
	var allUsers []*users.User
	err := st.db.All(&allUsers)
	if err == storm.ErrNotFound {
		return nil, errors.ErrNotExist
	}
	if err != nil {
		return allUsers, err
	}
	return allUsers, err
}

func (st usersBackend) Update(user *users.User, actorIsAdmin bool, fields ...string) error {
	existingUser, err := st.GetBy(user.ID)
	if err != nil {
		return err
	}
	passwordUser := existingUser.LoginMethod == users.LoginMethodPassword
	enforcedOtp := settings.Config.Auth.Methods.PasswordAuth.EnforcedOtp
	if passwordUser && enforcedOtp && !user.OtpEnabled {
		return errors.ErrNoTotpConfigured
	}
	fields, err = parseFields(user, fields, actorIsAdmin)
	if err != nil {
		return err
	}

	if !slices.Contains(fields, "Password") {
		user.Password = existingUser.Password
	} else {
		if existingUser.LockPassword {
			return fmt.Errorf("password cannot be changed when lock password is enabled")
		}
	}

	if !actorIsAdmin {
		err := checkRestrictedFields(existingUser, fields)
		if err != nil {
			return err
		}
	}

	if len(fields) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// converting scopes to map of paths intead of names (names can change)
	if slices.Contains(fields, "Scopes") {
		adjustedScopes, err := settings.ConvertToBackendScopes(user.Scopes)
		if err != nil {
			return err
		}
		user.Scopes = adjustedScopes
		err = files.MakeUserDirs(user, true)
		if err != nil {
			return err
		}
	}
	// Use reflection to access struct fields
	userFields := reflect.ValueOf(user).Elem() // Get struct value
	for _, field := range fields {
		// Get the corresponding field using reflection
		fieldValue := userFields.FieldByName(field)
		if !fieldValue.IsValid() {
			return fmt.Errorf("invalid field: %s", field)
		}

		// Ensure the field is settable
		if !fieldValue.CanSet() {
			return fmt.Errorf("cannot set value of field: %s", field)
		}

		// Get the value to be stored
		val := fieldValue.Interface()
		if field == "OtpEnabled" {
			otpEnabled, _ := val.(bool)
			if otpEnabled {
				continue
			}
			field = "TOTPSecret" // if otp is disabled, we also want to clear the TOTPSecret
			val = ""             // clear the TOTPSecret
		}
		// Update the database
		if err := st.db.UpdateField(existingUser, field, val); err != nil {
			return fmt.Errorf("failed to update user field: %s, error: %v", field, err)
		}
	}

	// last revoke api keys if needed.
	if existingUser.Permissions.Api && !user.Permissions.Api && slices.Contains(fields, "Permissions") {
		for _, key := range existingUser.ApiKeys {
			auth.RevokeAPIKey(key.Key) // add to blacklist
		}
	}
	return nil
}
func (st usersBackend) Save(user *users.User, changePass, disableScopeChange bool) error {
	if user.LoginMethod == "" {
		user.LoginMethod = users.LoginMethodPassword
	}
	logger.Debugf("Saving user [%s] changepass: %v", user.Username, changePass)
	if user.LoginMethod == users.LoginMethodPassword && changePass {
		err := checkPassword(user.Password)
		if err != nil {
			return err
		}
		pass, err := users.HashPwd(user.Password)
		if err != nil {
			return err
		}
		user.Password = pass
	}

	// converting scopes to map of paths intead of names (names can change)
	adjustedScopes, err := settings.ConvertToBackendScopes(user.Scopes)
	if err != nil {
		return err
	}
	user.Scopes = adjustedScopes
	err = files.MakeUserDirs(user, disableScopeChange)
	if err != nil {
		logger.Error(err.Error())
	}
	err = st.db.Save(user)
	if err == storm.ErrAlreadyExists {
		return fmt.Errorf("user with provided username already exists")
	}
	return err
}
func (st usersBackend) DeleteByID(id uint) error {
	return st.db.DeleteStruct(&users.User{ID: id})
}

func (st usersBackend) DeleteByUsername(username string) error {
	user, err := st.GetBy(username)
	if err != nil {
		return err
	}

	return st.db.DeleteStruct(user)
}
func checkPassword(password string) error {
	if len(password) < settings.Config.Auth.Methods.PasswordAuth.MinLength {
		return fmt.Errorf("password must be at least %d characters long", settings.Config.Auth.Methods.PasswordAuth.MinLength)
	}
	return nil
}
func checkRestrictedFields(existingUser *users.User, fields []string) error {
	// Get a list of allowed fields from NonAdminEditable
	allowed := getNonAdminEditableFieldNames()
	for _, field := range fields {
		if !slices.Contains(allowed, field) {
			return fmt.Errorf("non-admins cannot modify field: %s", field)
		}
	}

	return nil
}
func getNonAdminEditableFieldNames() []string {
	var names []string
	t := reflect.TypeOf(users.NonAdminEditable{})
	for i := 0; i < t.NumField(); i++ {
		names = append(names, t.Field(i).Name)
	}
	return names
}

func parseFields(user *users.User, fields []string, actorIsAdmin bool) ([]string, error) {
	if len(fields) == 0 || fields[0] == "all" {
		fields = []string{}
		v := reflect.ValueOf(user)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := v.Type()

		// Dynamically populate fields to update
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// which=all can't update password
			switch strings.ToLower(field.Name) {
			case "id", "username", "loginmethod", "password", "apikeys", "totpenabled", "totpsecret", "totpnonce":
				// Skip these fields
				continue
			}
			fields = append(fields, field.Name)
		}
	}
	newfields := []string{}
	for _, field := range fields {
		capitalField := utils.CapitalizeFirst(field)
		if capitalField == "Scopes" {
			if !actorIsAdmin {
				continue
			}
		}
		if capitalField == "Password" {
			if user.LoginMethod != users.LoginMethodPassword {
				return nil, fmt.Errorf("password cannot be changed when login method is not password")
			}
			err := checkPassword(user.Password)
			if err != nil {
				return nil, fmt.Errorf("password does not meet complexity requirements")
			}
			value, err := users.HashPwd(user.Password)
			if err != nil {
				logger.Error(err.Error())
			}
			user.Password = value
		}
		newfields = append(newfields, capitalField)
	}

	return newfields, nil
}
