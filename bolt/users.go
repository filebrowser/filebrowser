package bolt

import (
	"reflect"

	"github.com/asdine/storm"
	fm "github.com/filebrowser/filebrowser"
)

// UsersStore is a users store.
type UsersStore struct {
	DB *storm.DB
}

// Get gets a user with a certain id from the database.
func (u UsersStore) Get(id int, builder fm.FSBuilder) (*fm.User, error) {
	var us fm.User
	err := u.DB.One("ID", id, &us)
	if err == storm.ErrNotFound {
		return nil, fm.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	us.FileSystem = builder(us.Scope)
	return &us, nil
}

// GetByUsername gets a user with a certain username from the database.
func (u UsersStore) GetByUsername(username string, builder fm.FSBuilder) (*fm.User, error) {
	var us fm.User
	err := u.DB.One("Username", username, &us)
	if err == storm.ErrNotFound {
		return nil, fm.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	us.FileSystem = builder(us.Scope)
	return &us, nil
}

// Gets gets all the users from the database.
func (u UsersStore) Gets(builder fm.FSBuilder) ([]*fm.User, error) {
	var us []*fm.User
	err := u.DB.All(&us)
	if err == storm.ErrNotFound {
		return nil, fm.ErrNotExist
	}

	if err != nil {
		return us, err
	}

	for _, user := range us {
		user.FileSystem = builder(user.Scope)
	}

	return us, err
}

// Update updates the whole user object or only certain fields.
func (u UsersStore) Update(us *fm.User, fields ...string) error {
	if len(fields) == 0 {
		return u.Save(us)
	}

	for _, field := range fields {
		val := reflect.ValueOf(us).Elem().FieldByName(field).Interface()
		if err := u.DB.UpdateField(us, field, val); err != nil {
			return err
		}
	}

	return nil
}

// Save saves a user to the database.
func (u UsersStore) Save(us *fm.User) error {
	return u.DB.Save(us)
}

// Delete deletes a user from the database.
func (u UsersStore) Delete(id int) error {
	return u.DB.DeleteStruct(&fm.User{ID: id})
}
