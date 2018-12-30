package bolt

import (
	"reflect"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/types"
)

type UsersStore struct {
	DB *storm.DB
}

func (st UsersStore) Get(id uint) (*types.User, error) {
	user := &types.User{}
	err := st.DB.One("ID", id, user)
	if err == storm.ErrNotFound {
		return nil, types.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (st UsersStore) GetByUsername(username string) (*types.User, error) {
	user := &types.User{}
	err := st.DB.One("Username", username, user)
	if err == storm.ErrNotFound {
		return nil, types.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (st UsersStore) Gets() ([]*types.User, error) {
	users := []*types.User{}
	err := st.DB.All(&users)
	if err == storm.ErrNotFound {
		return nil, types.ErrNotExist
	}

	if err != nil {
		return users, err
	}

	return users, err
}

func (st UsersStore) Update(user *types.User, fields ...string) error {
	if len(fields) == 0 {
		return st.Save(user)
	}

	for _, field := range fields {
		val := reflect.ValueOf(user).Elem().FieldByName(field).Interface()
		if err := st.DB.UpdateField(user, field, val); err != nil {
			return err
		}
	}

	return nil
}

func (st UsersStore) Save(user *types.User) error {
	err := st.DB.Save(user)
	if err == storm.ErrAlreadyExists {
		return types.ErrExist
	}
	return err
}

func (st UsersStore) Delete(id uint) error {
	return st.DB.DeleteStruct(&types.User{ID: id})
}

func (st UsersStore) DeleteByUsername(username string) error {
	user, err := st.GetByUsername(username)
	if err != nil {
		return err
	}

	return st.DB.DeleteStruct(user)
}
