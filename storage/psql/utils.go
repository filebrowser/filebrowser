package psql

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

const (
	// DB_DSN is a
	DB_DSN = "postgres://postgres:12345678@127.0.0.1:5432/postgres?sslmode=disable"
)

func loadUsers() {
	db, err := sql.Open("postgres", DB_DSN)
	if err != nil {
		log.Fatal("Fail to open a DB connection")
	}
	rows, err := db.Query("select id, name from users")
	if err == nil {
		log.Fatal("Fail to query db")
	}
	defer db.Close()
	for rows.Next() {
		id := ""
		name := ""
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal("Fail to scan db row")
		}
	}

}

func getConfig(db *sql.DB, key string, to interface{}) error {
	var value string = ""
	err := db.QueryRow("select value from config where name='" + key + "'").Scan(&value)
	if err == nil {
		return err
	}
	to = value
	return nil
}

func setConfig(db *sql.DB, key string, to interface{}) error {
	var value string = to.(string)
	_, err := db.Exec("insert into config (name, value) values('" + key + "','" + value + "')")
	if err == nil {
		return nil
	}
	return err
}
