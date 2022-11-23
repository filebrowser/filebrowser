package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/users"
)

type usersBackend struct {
	db *sql.DB
}

func PermFromString(s string) users.Permissions {
	if s == "" {
		return users.Permissions{}
	}
	var perm users.Permissions
	json.Unmarshal([]byte(s), &perm)
	return perm
}

func PermToString(perm users.Permissions) string {
	data, err := json.Marshal(perm)
	if err != nil {
		return ""
	}
	return string(data)
}

func CommandsFromString(s string) []string {
	if s == "" {
		return []string{}
	}
	var commands []string
	err := json.Unmarshal([]byte(s), &commands)
	if err != nil {
		return nil
	}
	return commands
}

func CommandsToString(commands []string) string {
	data, err := json.Marshal(commands)
	if err != nil {
		return ""
	}
	return string(data)
}

func SortingFromString(s string) files.Sorting {
	if s == "" {
		return files.Sorting{}
	}
	var sorting files.Sorting
	json.Unmarshal([]byte(s), &sorting)
	return sorting
}

func SortingToString(sorting files.Sorting) string {
	data, err := json.Marshal(sorting)
	if err != nil {
		return ""
	}
	return string(data)
}

func rulesFromString(s string) []rules.Rule {
	if s == "" {
		return []rules.Rule{}
	}
	var rules []rules.Rule
	json.Unmarshal([]byte(s), &rules)
	return rules
}

func RulesToString(rules []rules.Rule) string {
	data, err := json.Marshal(rules)
	if err != nil {
		return ""
	}
	return string(data)
}

func InitUserTable(db *sql.DB) error {
	sql := "create table if not exists users (id integer primary key, username string, password string, scope string, lockpassword bool, viewmode string, perm string, commands string, sorting string, rules string);"
	_, err := db.Exec(sql)
	return err
}

func (s usersBackend) Get(id interface{}) (*users.User, error) {
	userID := id.(uint)
	username := ""
	password := ""
	scope := ""
	lockpassword := false
	var viewmode users.ViewMode = "list"
	perm := ""
	commands := ""
	sorting := ""
	rules := ""
	sql := "select username, password, scope, lockpassword, viewmode, perm,commands,sorting,rules from users where id=" + strconv.Itoa(int(userID))
	s.db.QueryRow(sql).Scan(&username, &password, &scope, &lockpassword, &viewmode, &perm, &commands, &sorting, &rules)
	user := users.User{}
	user.ID = userID
	user.Username = username
	user.Password = password
	user.Scope = scope
	user.LockPassword = lockpassword
	user.ViewMode = viewmode
	user.Perm = PermFromString(perm)
	user.Commands = CommandsFromString(commands)
	user.Sorting = SortingFromString(sorting)
	user.Rules = rulesFromString(rules)
	return &user, nil
}

func (s usersBackend) Gets() ([]*users.User, error) {
	sql := "select id, username, password, scope, lockpassword, viewmode, perm,commands,sorting,rules from users"
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}
	var users2 []*users.User = []*users.User{}
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
		if err != nil {
			fmt.Println("Fail to parse record for user.User")
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

func (s usersBackend) GetBy(id interface{}) (*users.User, error) {
	return s.Get(id)
}

func (s usersBackend) updateUser(id uint, user *users.User) error {
	lockpassword := 0
	if user.LockPassword {
		lockpassword = 1
	}
	sql := fmt.Sprintf(
		"update users set username='%s',password='%s',scope='%s',lockpassword=%d,viewmode='%s',perm='%s',commands='%s',sorting='%s',rules='%s' where id=%d",
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
	return err
}

func (s usersBackend) insertUser(user *users.User) error {
	sql := fmt.Sprintf(
		"insert into users (username, password, scope, lockpassword, viewmode, perm, commands, sorting) values ('%s','%s','%s',%d,'%s','%s','%s','%s','%s')",
		user.Username,
		user.Password,
		user.Scope,
		user.LockPassword,
		user.ViewMode,
		PermToString(user.Perm),
		CommandsToString(user.Commands),
		SortingToString(user.Sorting),
		RulesToString(user.Rules),
	)
	_, err := s.db.Exec(sql)
	return err
}

func (s usersBackend) Save(user *users.User) error {
	userOriginal, err := s.GetBy(user.ID)
	if err != nil {
		return err
	}
	if userOriginal != nil {
		return s.updateUser(user.ID, user)
	}
	return s.insertUser(user)
}

func (s usersBackend) DeleteByID(id uint) error {
	sql := "delete from users where id=" + strconv.Itoa(int(id))
	_, err := s.db.Exec(sql)
	return err
}

func (s usersBackend) DeleteByUsername(username string) error {
	sql := "delete from users where username='" + username + "'"
	_, err := s.db.Exec(sql)
	return err
}

func (s usersBackend) Update(u *users.User, fields ...string) error {
	var setItems = []string{}
	for _, field := range fields {
		userField := reflect.ValueOf(u).Elem().FieldByName(field)
		if !userField.IsValid() {
			continue
		}
		val := userField.Interface()
		if reflect.TypeOf(val).Kind().String() == "string" {
			setItems = append(setItems, fmt.Sprintf("%s='%s'", field, val))
		} else {
			// TODO
			setItems = append(setItems, fmt.Sprintf("%s=%d", field, val))
		}
	}
	sql := fmt.Sprintf("update users set %s if id=%d", strings.Join(setItems, ","), u.ID)
	_, err := s.db.Exec(sql)
	return err
}
