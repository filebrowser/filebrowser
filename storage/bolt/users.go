package bolt

import (
	"reflect"

	"github.com/asdine/storm"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

type usersBackend struct {
	db *storm.DB
}

func (st usersBackend) GetBy(i interface{}) (user *users.User, err error) {
	user = &users.User{}

	var arg string
	switch i.(type) {
	case uint:
		arg = "ID"
	case string:
		arg = "Username"
	default:
		return nil, errors.ErrInvalidDataType
	}

	err = st.db.One(arg, i, user)

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

func (st usersBackend) Update(user *users.User, fields ...string) error {
	if len(fields) == 0 {
		return st.Save(user)
	}

	for _, field := range fields {
		val := reflect.ValueOf(user).Elem().FieldByName(field).Interface()
		if err := st.db.UpdateField(user, field, val); err != nil {
			return err
		}
	}

	return nil
}

func (st usersBackend) Save(user *users.User) error {
	err := st.db.Save(user)
	if err == storm.ErrAlreadyExists {
		return errors.ErrExist
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
