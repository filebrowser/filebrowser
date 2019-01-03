package bolt

import (
	"reflect"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/types"
)

func (st Backend) GetUserByID(id uint) (*types.User, error) {
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

func (st Backend) GetUserByUsername(username string) (*types.User, error) {
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

func (st Backend) GetUsers() ([]*types.User, error) {
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

func (st Backend) UpdateUser(user *types.User, fields ...string) error {
	if len(fields) == 0 {
		return st.SaveUser(user)
	}

	for _, field := range fields {
		val := reflect.ValueOf(user).Elem().FieldByName(field).Interface()
		if err := st.DB.UpdateField(user, field, val); err != nil {
			return err
		}
	}

	return nil
}

func (st Backend) SaveUser(user *types.User) error {
	err := st.DB.Save(user)
	if err == storm.ErrAlreadyExists {
		return types.ErrExist
	}
	return err
}

func (st Backend) DeleteUserByID(id uint) error {
	return st.DB.DeleteStruct(&types.User{ID: id})
}

func (st Backend) DeleteUserByUsername(username string) error {
	user, err := st.GetUserByUsername(username)
	if err != nil {
		return err
	}

	return st.DB.DeleteStruct(user)
}
