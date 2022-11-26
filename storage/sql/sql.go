package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
}

type DBConnectionRecord struct {
	db     *sql.DB
	dbType string
	path   string
}

var (
	dbRecords map[string]*DBConnectionRecord = map[string]*DBConnectionRecord{}
)

// GetDBType used to get the driver type of a sql.DB
// It is based on existing dbRecords
// All sql.DB should opened by OpenDB
func GetDBType(db *sql.DB) (string, error) {
	for _, record := range dbRecords {
		if record.db == db {
			return record.dbType, nil
		}
	}
	return "", errors.New("No such database open by this module")
}

func getNameQuote(dbType string) string {
	if dbType == "mysql" {
		return "`"
	}
	return "\""
}

// for mysql, it is ``
// for postgres and sqlite, it is ""
func quoteName(dbType string, name string) string {
	q := getNameQuote(dbType)
	return q + name + q
}

// placeholder for sql stmt
// for postgres, it is $1, $2, $3...
// for mysql and sqlite3, it is ?,?,?...
func placeHolder(dbType string, index int) string {
	if index <= 0 {
		panic("the placeholder index should >= 1")
	}
	if dbType == "postgres" {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

func IsDBPath(path string) bool {
	prefixes := []string{"sqlite3", "postgres", "mysql"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix+"://") {
			return true
		}
	}
	return false
}

func OpenDB(path string) (*sql.DB, error) {
	if val, ok := dbRecords[path]; ok {
		return val.db, nil
	}
	prefixes := []string{"sqlite3", "postgres", "mysql"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			db, err := connectDB(prefix, path)
			if !checkError(err, "Fail to connect database "+path) {
				dbRecords[path] = &DBConnectionRecord{db: db, dbType: prefix, path: path}
			}
			return db, err
		}
	}
	return nil, errors.New("Unsupported db scheme")
}

type DatabaseResource struct {
	scheme   string
	username string
	password string
	host     string
	port     int
	database string
}

func ParseDatabasePath(path string) (*DatabaseResource, error) {
	pattern := "^(([a-zA-Z0-9]+)://)?(([^:]+)(:(.*))?@)?([a-zA-Z0-9_.]+)(:([0-9]+))?(/([a-zA-Z0-9_-]+))?$"
	reg, err := regexp.Compile(pattern)
	if checkError(err, "Fail to compile regexp") {
		return nil, err
	}
	matches := reg.FindAllStringSubmatch(path, -1)
	if matches == nil || len(matches) == 0 {
		return nil, errors.New("Fail to parse database")
	}
	r := DatabaseResource{}
	r.scheme = matches[0][2]
	r.username = matches[0][4]
	r.password = matches[0][6]
	r.host = matches[0][7]
	if len(matches[0][9]) > 0 {
		port, err := strconv.Atoi(matches[0][9])
		if !checkError(err, "Fail to parse port") {
			r.port = port
		}
	}
	r.database = matches[0][11]
	return &r, nil
}

// mysql://user:password@host:port/db => mysql://user:password@tcp(host:port)/db
func transformMysqlPath(path string) (string, error) {
	r, err := ParseDatabasePath(path)
	if checkError(err, "Fail to parse database path") {
		return "", err
	}
	scheme := r.scheme
	if len(scheme) == 0 {
		scheme = "mysql"
	}
	credential := ""
	if len(r.username) > 0 && len(r.password) > 0 {
		credential = r.username + ":" + r.password + "@"
	} else if len(r.username) > 0 {
		credential = r.username + "@"
	}
	host := r.host
	port := r.port
	if port == 0 {
		port = 3306
	}
	if len(r.database) == 0 {
		return "", errors.New("no database found in path")
	}
	return fmt.Sprintf("%s://%stcp(%s:%d)/%s", scheme, credential, host, port, r.database), nil
}

func connectDB(dbType string, path string) (*sql.DB, error) {
	if dbType == "sqlite3" && strings.HasPrefix(path, "sqlite3://") {
		path = strings.TrimPrefix(path, "sqlite3://")
	} else if dbType == "mysql" && strings.HasPrefix(path, "mysql://") {
		p, err := transformMysqlPath(path)
		if checkError(err, "Fail to parse mysql path") {
			return nil, err
		}
		path = p
		path = strings.TrimPrefix(path, "mysql://")
	}
	db, err := sql.Open(dbType, path)
	if err == nil {
		return db, nil
	}
	return nil, err
}

func NewStorage(db *sql.DB) (*storage.Storage, error) {
	dbType, err := GetDBType(db)
	checkError(err, "Fail to get database type, maybe this sql.DB is not opened by OpenDB")

	userStore := users.NewStorage(newUsersBackend(db, dbType))
	shareStore := share.NewStorage(newShareBackend(db, dbType))
	settingsStore := settings.NewStorage(newSettingsBackend(db, dbType))
	authStore := auth.NewStorage(newAuthBackend(db, dbType), userStore)

	storage := &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
	}
	return storage, nil
}
