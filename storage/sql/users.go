package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/users"
)

type usersBackend struct {
	db     *sql.DB
	dbType string
}

func PermFromString(s string) users.Permissions {
	var perm users.Permissions
	if s == "" {
		return perm
	}
	err := json.Unmarshal([]byte(s), &perm)
	checkError(err, "Fail to parse perm from string")
	return perm
}

func PermToString(perm users.Permissions) string {
	data, err := json.Marshal(perm)
	if checkError(err, "Fail to stringify users.Permissions") {
		return ""
	}
	return string(data)
}

func CommandsFromString(s string) []string {
	if s == "" {
		return make([]string, 0)
	}
	var commands []string
	err := json.Unmarshal([]byte(s), &commands)
	checkError(err, "Fail to parse users Commands")
	return commands
}

func CommandsToString(commands []string) string {
	data, err := json.Marshal(commands)
	if checkError(err, "Fail to stringify users commands") {
		return ""
	}
	return string(data)
}

func SortingFromString(s string) files.Sorting {
	if s == "" {
		return files.Sorting{}
	}
	var sorting files.Sorting
	err := json.Unmarshal([]byte(s), &sorting)
	checkError(err, "Fail to parse Sorting from string")
	return sorting
}

func SortingToString(sorting files.Sorting) string {
	data, err := json.Marshal(sorting)
	if checkError(err, "Fail to stringify files.Sorting") {
		return ""
	}
	return string(data)
}

func rulesFromString(s string) []rules.Rule {
	rules := make([]rules.Rule, 0)
	if s == "" {
		return rules
	}
	err := json.Unmarshal([]byte(s), &rules)
	checkError(err, "Fail to parse Rules from string")
	return rules
}

func RulesToString(rules []rules.Rule) string {
	data, err := json.Marshal(rules)
	if checkError(err, "Fail to stringify []rules.Rule") {
		return ""
	}
	return string(data)
}

var adminUser = createAdminUser()

func createAdminUser() users.User {
	userDefault := defaultSettings.Defaults
	user := users.User{}
	user.Username = "admin"
	user.Password = "admin"
	user.Scope = userDefault.Scope
	user.LockPassword = false
	user.ViewMode = userDefault.ViewMode
	user.Perm = users.Permissions{
		Admin:    true,
		Execute:  true,
		Create:   true,
		Rename:   true,
		Modify:   true,
		Delete:   true,
		Share:    true,
		Download: true,
	}
	user.Commands = userDefault.Commands
	user.Sorting = userDefault.Sorting
	return user
}

func InitUserTable(db *sql.DB, dbType string) error {
	primaryKey := "integer primary key"
	if dbType == "postgres" {
		primaryKey = "serial primary key"
	} else if dbType == "mysql" {
		primaryKey = "int unsigned primary key auto_increment"
	}
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id %s, username text, password text, scope text, locale text, lockpassword integer, viewmode text, perm text, commands text, sorting text, rules text, hidedotfiles integer, dateformat integer, singleclick integer);", quoteName(dbType, UsersTable), primaryKey)
	_, err := db.Exec(sql)
	checkError(err, "Fail to create users table")
	return err
}

func newUsersBackend(db *sql.DB, dbType string) usersBackend {
	InitUserTable(db, dbType)
	return usersBackend{db: db, dbType: dbType}
}

func (s usersBackend) GetBy(i interface{}) (*users.User, error) {
	columns := []string{"id", "username", "password", "scope", "locale", "lockpassword", "viewmode", "perm", "commands", "sorting", "rules", "hidedotfiles", "dateformat", "singleclick"}
	columnsStr := strings.Join(columns, ",")
	var conditionStr string
	switch i.(type) {
	case uint:
		conditionStr = fmt.Sprintf("id=%v", i)
	case int:
		conditionStr = fmt.Sprintf("id=%v", i)
	case string:
		conditionStr = fmt.Sprintf("username='%v'", i)
	default:
		return nil, errors.ErrInvalidDataType
	}
	userID := uint(0)
	username := ""
	password := ""
	scope := ""
	locale := ""
	lockpassword := false
	var viewmode users.ViewMode = users.ListViewMode
	perm := ""
	commands := ""
	sorting := ""
	rules := ""
	hidedotfiles := false
	dateformat := false
	singleclick := false
	user := users.User{}
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", columnsStr, quoteName(s.dbType, UsersTable), conditionStr)
	err := s.db.QueryRow(sql).Scan(&userID, &username, &password, &scope, &locale, &lockpassword, &viewmode, &perm, &commands, &sorting, &rules, &hidedotfiles, &dateformat, &singleclick)
	if checkError(err, "") {
		return nil, err
	}
	user.ID = userID
	user.Username = username
	user.Password = password
	user.Scope = scope
	user.Locale = locale
	user.LockPassword = lockpassword
	user.ViewMode = viewmode
	user.Perm = PermFromString(perm)
	user.Commands = CommandsFromString(commands)
	user.Sorting = SortingFromString(sorting)
	user.Rules = rulesFromString(rules)
	user.HideDotfiles = hidedotfiles
	user.DateFormat = dateformat
	user.SingleClick = singleclick
	return &user, nil
}

func (s usersBackend) Gets() ([]*users.User, error) {
	sql := fmt.Sprintf("SELECT id, username, password, scope, lockpassword, viewmode, perm,commands,sorting,rules FROM %s", quoteName(s.dbType, UsersTable))
	rows, err := s.db.Query(sql)
	if checkError(err, "Fail to Query []*users.User") {
		return nil, err
	}
	var users2 []*users.User = make([]*users.User, 0)
	for rows.Next() {
		id := 0
		username := ""
		password := ""
		scope := ""
		lockpassword := false
		var viewmode users.ViewMode = "list"
		perm := ""
		commands := ""
		sorting := ""
		rules := ""
		err := rows.Scan(&id, &username, &password, &scope, &lockpassword, &viewmode, &perm, &commands, &sorting, &rules)
		if checkError(err, "Fail to parse record for user.User") {
			continue
		}
		user := users.User{}
		user.ID = uint(id)
		user.Username = username
		user.Password = password
		user.Scope = scope
		user.LockPassword = lockpassword
		user.ViewMode = viewmode
		user.Perm = PermFromString(perm)
		user.Commands = CommandsFromString(commands)
		user.Sorting = SortingFromString(sorting)
		user.Rules = rulesFromString(rules)

		users2 = append(users2, &user)
	}
	return users2, nil
}

func (s usersBackend) updateUser(id uint, user *users.User) error {
	lockpassword := 0
	if user.LockPassword {
		lockpassword = 1
	}
	sql := fmt.Sprintf(
		"UPDATE %s SET username='%s',password='%s',scope='%s',lockpassword=%d,viewmode='%s',perm='%s',commands='%s',sorting='%s',rules='%s' WHERE id=%d",
		quoteName(s.dbType, UsersTable),
		user.Username,
		user.Password,
		user.Scope,
		lockpassword,
		user.ViewMode,
		PermToString(user.Perm),
		CommandsToString(user.Commands),
		SortingToString(user.Sorting),
		RulesToString(user.Rules),
		user.ID,
	)
	_, err := s.db.Exec(sql)
	checkError(err, "Fail to update user")
	return err
}

func (s usersBackend) insertUser(user *users.User) error {
	columnSpec := [][]string{
		{"username", "'%s'"},
		{"password", "'%s'"},
		{"scope", "'%s'"},
		{"locale", "'%s'"},
		{"lockpassword", "%s"},
		{"viewmode", "'%s'"},
		{"perm", "'%s'"},
		{"commands", "'%s'"},
		{"sorting", "'%s'"},
		{"rules", "'%s'"},
		{"hidedotfiles", "%s"},
		{"dateformat", "%s"},
		{"singleclick", "%s"},
	}
	columns := []string{}
	specs := []string{}
	for _, c := range columnSpec {
		columns = append(columns, c[0])
		specs = append(specs, c[1])
	}
	columnStr := strings.Join(columns, ",")
	specStr := strings.Join(specs, ",")
	sqlFormat := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", quoteName(s.dbType, UsersTable), columnStr, specStr)
	if s.dbType == "postgres" {
		sqlFormat = sqlFormat + " RETURNING id;"
	}
	sql := fmt.Sprintf(
		sqlFormat,
		user.Username,
		user.Password,
		user.Scope,
		user.Locale,
		boolToString(user.LockPassword),
		user.ViewMode,
		PermToString(user.Perm),
		CommandsToString(user.Commands),
		SortingToString(user.Sorting),
		RulesToString(user.Rules),
		boolToString(user.HideDotfiles),
		boolToString(user.DateFormat),
		boolToString(user.SingleClick),
	)
	if s.dbType == "postgres" {
		id := uint(0)
		err := s.db.QueryRow(sql).Scan(&id)
		if !checkError(err, "Fail to insert user") {
			user.ID = id
		}
		return err
	}
	res, err := s.db.Exec(sql)
	if !checkError(err, "Fail to insert user") {
		id, err := res.LastInsertId()
		checkError(err, "Fail to get last inserted id")
		user.ID = uint(id)
	}
	return err
}

func (s usersBackend) Save(user *users.User) error {
	userOriginal, err := s.GetBy(user.Username)
	checkError(err, "")
	if userOriginal != nil {
		return s.updateUser(user.ID, user)
	}
	return s.insertUser(user)
}

func (s usersBackend) DeleteByID(id uint) error {
	sql := fmt.Sprintf("delete from %s where id=%d", quoteName(s.dbType, UsersTable), id)
	_, err := s.db.Exec(sql)
	checkError(err, "Fail to delete User by id")
	return err
}

func (s usersBackend) DeleteByUsername(username string) error {
	sql := fmt.Sprintf("delete from %s where username='%s'", quoteName(s.dbType, UsersTable), username)
	_, err := s.db.Exec(sql)
	checkError(err, "Fail to delete user by username")
	return err
}

func (s usersBackend) Update(u *users.User, fields ...string) error {
	if len(fields) == 0 {
		return s.Save(u)
	}
	var setItems = []string{}
	for _, field := range fields {
		userField := reflect.ValueOf(u).Elem().FieldByName(field)
		if !userField.IsValid() {
			continue
		}
		field = strings.ToLower(field)
		val := userField.Interface()
		typeStr := reflect.TypeOf(val).Kind().String()
		if typeStr == "string" {
			setItems = append(setItems, fmt.Sprintf("%s='%s'", field, val))
		} else if typeStr == "bool" {
			setItems = append(setItems, fmt.Sprintf("%s=%s", field, boolToString(val.(bool))))
		} else {
			// TODO
			setItems = append(setItems, fmt.Sprintf("%s=%s", field, val))
		}
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE id=%d", UsersTable, strings.Join(setItems, ","), u.ID)
	_, err := s.db.Exec(sql)
	checkError(err, "Fail to update user")
	return err
}
