package bolt

import (
	"reflect"

	"github.com/asdine/storm"
	fm "github.com/hacdias/filemanager"
)

type UsersStore struct {
	DB *storm.DB
}

func (u UsersStore) Get(id int) (*fm.User, error) {
	var us *fm.User
	err := u.DB.One("ID", id, us)
	if err == storm.ErrNotFound {
		return nil, fm.ErrUserNotExist
	}

	if err != nil {
		return nil, err
	}

	return &fm.User{}, nil
}

func (u UsersStore) Gets() ([]*fm.User, error) {
	var us []*fm.User
	err := u.DB.All(us)
	return us, err
}

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

func (u UsersStore) Save(us *fm.User) error {
	return u.DB.Save(us)
}

func (u UsersStore) Delete(id int) error {
	return u.DB.DeleteStruct(&fm.User{ID: id})
}
