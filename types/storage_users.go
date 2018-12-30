package types

// UsersStore is used to manage users relativey to a data storage.
type UsersStore interface {
	Get(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	Gets() ([]*User, error)
	Save(u *User) error
	Update(u *User, fields ...string) error
	Delete(id uint) error
	DeleteByUsername(username string) error
}

// UsersVerify wraps a UsersStore and makes the verifications needed.
type UsersVerify struct {
	Store UsersStore
}

// Get wraps a UsersStore.Get to verify if everything is right.
func (v UsersVerify) Get(id uint) (*User, error) {
	user, err := v.Store.Get(id)
	if err != nil {
		return nil, err
	}

	user.clean()
	return user, nil
}

// GetByUsername wraps a UsersStore.GetByUsername to verify if everything is right.
func (v UsersVerify) GetByUsername(username string) (*User, error) {
	user, err := v.Store.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	user.clean()
	return user, nil
}

// Gets wraps a UsersStore.Gets to verify if everything is right.
func (v UsersVerify) Gets() ([]*User, error) {
	users, err := v.Store.Gets()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.clean()
	}

	return users, err
}

// Update wraps a UsersStore.Update to verify if everything is right.
func (v UsersVerify) Update(user *User, fields ...string) error {
	err := user.clean(fields...)
	if err != nil {
		return err
	}

	return v.Store.Update(user, fields...)
}

// Save wraps a UsersStore.Save to verify if everything is right.
func (v UsersVerify) Save(user *User) error {
	if err := user.clean(); err != nil {
		return err
	}

	return v.Store.Save(user)
}

// Delete wraps a UsersStore.Delete to verify if everything is right.
func (v UsersVerify) Delete(id uint) error {
	return v.Store.Delete(id)
}

// DeleteByUsername wraps a UsersStore.DeleteByUsername to verify if everything is right.
func (v UsersVerify) DeleteByUsername(username string) error {
	return v.Store.DeleteByUsername(username)
}
