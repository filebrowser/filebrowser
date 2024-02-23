package sql

import (
	"encoding/json"
	"fmt"

	"github.com/filebrowser/filebrowser/v2/settings"
)

var defaultServer = settings.Server{
	Port:                  "8080",
	Log:                   "stdout",
	EnableThumbnails:      false,
	ResizePreview:         false,
	EnableExec:            false,
	TypeDetectionByHeader: false,
}

func cloneServer(server settings.Server) settings.Server {
	data, err := json.Marshal(server)
	s := settings.Server{}
	if checkError(err, "Fail to clone settings.Server") {
		return s
	}
	err = json.Unmarshal(data, &s)
	checkError(err, "Fail to decode for settings.Server")
	return s
}

func (s settingsBackend) GetServer() (*settings.Server, error) {
	sql := fmt.Sprintf("select %s, value from %s", quoteName(s.dbType, "key"), quoteName(s.dbType, SettingsTable))
	rows, err := s.db.Query(sql)
	if checkError(err, "Fail to Query for GetServer") {
		return nil, err
	}
	server := cloneServer(defaultServer)
	key := ""
	value := ""

	for rows.Next() {
		err = rows.Scan(&key, &value)
		if checkError(err, "Fail to query settings.Settings") {
			continue
		}
		if key == "Root" {
			server.Root = value
		} else if key == "BaseURL" {
			server.BaseURL = value
		} else if key == "Socket" {
			server.Socket = value
		} else if key == "TLSKey" {
			server.TLSKey = value
		} else if key == "TLSCert" {
			server.TLSCert = value
		} else if key == "Port" {
			server.Port = value
		} else if key == "Address" {
			server.Address = value
		} else if key == "Log" {
			server.Log = value
		} else if key == "EnableThumbnails" {
			server.EnableThumbnails = boolFromString(value)
		} else if key == "ResizePreview" {
			server.ResizePreview = boolFromString(value)
		} else if key == "EnableExec" {
			server.EnableExec = boolFromString(value)
		} else if key == "TypeDetectionByHeader" {
			server.TypeDetectionByHeader = boolFromString(value)
		} else if key == "AuthHook" {
			server.AuthHook = value
		}
	}
	return &server, nil
}

func (s settingsBackend) SaveServer(ss *settings.Server) error {
	fields := []string{"Root", "BaseURL", "Socket", "TLSKey", "TLSCert", "Port", "Address", "Log", "EnableThumbnails", "ResizePreview", "EnableExec", "TypeDetectionByHeader", "AuthHook"}
	values := []string{
		ss.Root,
		ss.BaseURL,
		ss.Socket,
		ss.TLSKey,
		ss.TLSCert,
		ss.Port,
		ss.Address,
		ss.Log,
		boolToString(ss.EnableThumbnails),
		boolToString(ss.ResizePreview),
		boolToString(ss.EnableExec),
		boolToString(ss.TypeDetectionByHeader),
		ss.AuthHook}
	tx, err := s.db.Begin()
	if checkError(err, "Fail to begin db transaction") {
		return err
	}
	table := quoteName(s.dbType, SettingsTable)
	keyName := quoteName(s.dbType, "key")
	p1 := placeHolder(s.dbType, 1)
	p2 := placeHolder(s.dbType, 2)

	Insert := func(key string, value string) bool {
		insertSql := fmt.Sprintf("INSERT INTO %s (%s, value) VALUES(%s,%s)", table, keyName, p1, p2)
		stmt, err := s.db.Prepare(insertSql)
		defer stmt.Close()
		if checkError(err, "Fail to prepare statement") {
			tx.Rollback()
			return false
		}
		_, err = stmt.Exec(key, value)
		if checkError(err, "Fail to insert field "+key+" of settings.Server") {
			tx.Rollback()
			return false
		}
		return true
	}
	Update := func(key string, value string) bool {
		updateSql := fmt.Sprintf("UPDATE %s SET value=%s WHERE %s=%s", table, p1, keyName, p2)
		stmt, err := s.db.Prepare(updateSql)
		defer stmt.Close()
		if checkError(err, "Fail to prepare statement") {
			tx.Rollback()
			return false
		}
		_, err = stmt.Exec(value, key)
		if checkError(err, "Fail to update field "+key+" of settings.Server") {
			tx.Rollback()
			return false
		}
		return true
	}
	Exist := func(key string) bool {
		querySql := fmt.Sprintf("SELECT count(*) FROM %s WHERE %s=%s", table, keyName, p1)
		row := s.db.QueryRow(querySql, key)
		count := 0
		err := row.Scan(&count)
		if checkError(err, "Fail to Query "+key+" for GetServer") {
			return false
		}
		return count == 1
	}
	InsertOrUpdate := func(key string, value string) bool {
		if Exist(key) {
			return Update(key, value)
		} else {
			return Insert(key, value)
		}
	}

	for i, field := range fields {
		if !InsertOrUpdate(field, values[i]) {
			break
		}
	}
	err = tx.Commit()
	if checkError(err, "Fail to commit") {
		tx.Rollback()
		return err
	}
	return err
}
