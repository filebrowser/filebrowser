package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type authBackend struct {
	db     *sql.DB
	dbType string
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
		fmt.Println("ERROR: unknown auth method " + t)
		return nil, errors.ErrInvalidAuthMethod
	}
	return auther, nil
}

func (s authBackend) Save(a auth.Auther) error {
	val, err := json.Marshal(a)
	if checkError(err, "Fail to save auth.Auther") {
		return err
	}
	return SetSetting(s.db, s.dbType, "auther", string(val))
}

func newAuthBackend(db *sql.DB, dbType string) authBackend {
	return authBackend{db: db, dbType: dbType}
}
