package importer

import (
	"encoding/json"
	"fmt"

	"github.com/asdine/storm/v3"
	bolt "go.etcd.io/bbolt"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

type oldUser struct {
	ID            int           `storm:"id,increment"`
	Admin         bool          `json:"admin"`
	AllowCommands bool          `json:"allowCommands"` // Execute commands
	AllowEdit     bool          `json:"allowEdit"`     // Edit/rename files
	AllowNew      bool          `json:"allowNew"`      // Create files and folders
	AllowPublish  bool          `json:"allowPublish"`  // Publish content (to use with static gen)
	LockPassword  bool          `json:"lockPassword"`
	Commands      []string      `json:"commands"`
	Locale        string        `json:"locale"`
	Password      string        `json:"password"`
	Rules         []*rules.Rule `json:"rules"`
	Scope         string        `json:"filesystem"`
	Username      string        `json:"username" storm:"index,unique"`
	ViewMode      string        `json:"viewMode"`
}

func readOldUsers(db *storm.DB) ([]*oldUser, error) {
	var oldUsers []*oldUser
	err := db.Bolt.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("User")).ForEach(func(_ []byte, v []byte) error {
			if len(v) > 0 && string(v)[0] == '{' {
				user := &oldUser{}
				err := json.Unmarshal(v, user)

				if err != nil {
					return err
				}

				oldUsers = append(oldUsers, user)
			}

			return nil
		})
	})

	return oldUsers, err
}

func convertUsersToNew(old []*oldUser) ([]*users.User, error) {
	list := []*users.User{}

	for _, oldUser := range old {
		user := &users.User{
			Username:     oldUser.Username,
			Password:     oldUser.Password,
			Scope:        oldUser.Scope,
			Locale:       oldUser.Locale,
			LockPassword: oldUser.LockPassword,
			ViewMode:     users.ViewMode(oldUser.ViewMode),
			Commands:     oldUser.Commands,
			Rules:        []rules.Rule{},
			Perm: users.Permissions{
				Admin:    oldUser.Admin,
				Execute:  oldUser.AllowCommands,
				Create:   oldUser.AllowNew,
				Rename:   oldUser.AllowEdit,
				Modify:   oldUser.AllowEdit,
				Delete:   oldUser.AllowEdit,
				Share:    true,
				Download: true,
			},
		}

		for _, rule := range oldUser.Rules {
			user.Rules = append(user.Rules, *rule)
		}

		err := user.Clean("")
		if err != nil {
			return nil, err
		}

		list = append(list, user)
	}

	return list, nil
}

func importUsers(old *storm.DB, sto *storage.Storage) error {
	oldUsers, err := readOldUsers(old)
	if err != nil {
		return err
	}

	newUsers, err := convertUsersToNew(oldUsers)
	if err != nil {
		return err
	}

	for _, user := range newUsers {
		err = sto.Users.Save(user)
		if err != nil {
			return err
		}
	}

	fmt.Printf("%d users successfully imported into the new DB.\n", len(newUsers))
	return nil
}
