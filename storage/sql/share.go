package sql

import (
	"database/sql"
	"fmt"

	"github.com/filebrowser/filebrowser/v2/share"
)

type shareBackend struct {
	db *sql.DB
}

type linkRecord interface {
	Scan(dest ...interface{}) error
}

func InitShareTable(db *sql.DB) error {
	sql := "create table if not exists share_links (path string, userid integer, expire integer, passwordhash string, token string)"
	_, err := db.Exec(sql)
	return err
}

func parseLink(row linkRecord) (*share.Link, error) {
	path := ""
	hash := ""
	userid := uint(0)
	expire := int64(0)
	passwordhash := ""
	token := ""
	err := row.Scan(&path, &hash, &userid, &expire, &passwordhash, &token)
	if checkError(err, "Fail to parse record for share.Link") {
		return nil, err
	}
	link := share.Link{}
	link.Path = path
	link.Hash = hash
	link.UserID = userid
	link.Expire = expire
	link.PasswordHash = passwordhash
	link.Token = token
	return &link, nil
}

func queryLinks(db *sql.DB, condition string) ([]*share.Link, error) {
	sql := "select hash, path, userid, expire, passwordhash, token from share_links"
	if len(condition) > 0 {
		sql = sql + " where " + condition
	}
	rows, err := db.Query(sql)
	if checkError(err, "Fail to Query links") {
		return nil, err
	}
	var links []*share.Link = []*share.Link{}
	for rows.Next() {
		link, err := parseLink(rows)
		if checkError(err, "Fail to parse record for share.Link") {
			continue
		}
		links = append(links, link)
	}
	return links, nil
}

func (s shareBackend) All() ([]*share.Link, error) {
	return queryLinks(s.db, "")
}

func (s shareBackend) FindByUserID(id uint) ([]*share.Link, error) {
	condition := fmt.Sprintf("userid=%d", id)
	return queryLinks(s.db, condition)
}

func (s shareBackend) GetByHash(hash string) (*share.Link, error) {
	sql := fmt.Sprintf("select hash, path, userid, expire, passwordhash, token from share_links where hash='%s'", hash)
	return parseLink(s.db.QueryRow(sql))
}

func (s shareBackend) GetPermanent(path string, id uint) (*share.Link, error) {
	sql := fmt.Sprintf("select hash, path, userid, expire, passwordhash, token from share_links where path='%s' and userid=%d", path, id)
	return parseLink(s.db.QueryRow(sql))
}

func (s shareBackend) Gets(path string, id uint) ([]*share.Link, error) {
	condition := fmt.Sprintf("userid=%d and path='%s'", id, path)
	return queryLinks(s.db, condition)
}
func (s shareBackend) Save(l *share.Link) error {
	sql := fmt.Sprintf("insert into share_links (hash, path, userid, expire, passwordhash, token) values('%s', '%s', %d, %d, '%s', '%s')", l.Hash, l.Path, l.UserID, l.Expire, l.PasswordHash, l.Token)
	_, err := s.db.Exec(sql)
	return err
}
func (s shareBackend) Delete(hash string) error {
	sql := fmt.Sprintf("DELETE FROM share_links WHERE hash='%s'", hash)
	_, err := s.db.Exec(sql)
	return err
}
