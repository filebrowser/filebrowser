package sql

import (
	"database/sql"
	"encoding/json"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type authBackend struct {
	db *sql.DB
}

func InitAuthTable(db *sql.DB) error {
	sql := "create table if not exists appliction(key string primary key, value string)"
	_, err := db.Exec(sql)
	return err
}

func (s authBackend) Get(t settings.AuthMethod) (auth.Auther, error) {
	var auther auth.Auther

	switch t {
	case auth.MethodJSONAuth:
		auther = &auth.JSONAuth{}
	case auth.MethodProxyAuth:
		auther = &auth.ProxyAuth{}
	case auth.MethodHookAuth:
		auther = &auth.HookAuth{}
	case auth.MethodNoAuth:
		auther = &auth.NoAuth{}
	default:
		return nil, errors.ErrInvalidAuthMethod
	}

	val := GetSetting(s.db, "auther")
	if val == "" {
		return auther, nil
	}
	return auther, json.Unmarshal([]byte(val), auther)
}

func (s authBackend) Save(a auth.Auther) error {
	val, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return SetSetting(s.db, "auther", string(val))
}
