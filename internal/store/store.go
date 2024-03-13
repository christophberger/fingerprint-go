package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Users struct {
	db *sql.DB
}

// NewUsers opens the users database and creates a users table if it doesn't exist. It errors out if the database cannot be opened or the create statement fails.
func NewUsers(path string) (*Users, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("openDB: %w", err)
	}

	err = createTable(db)
	if err != nil {
		return nil, fmt.Errorf("openDB: %w", err)
	}

	return &Users{db: db}, nil
}

// CreateTable takes a database connection and creates a users table if it doesn't exist. It errors out if the create statement fails.
func createTable(db *sql.DB) error {
	sqlStmt := `create table if not exists users(
   	id integer primary key autoincrement,
   	email text unique,
   	password text,
   	visitor_id text unique
   );`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("createTable: %w", err)
	}

	return nil
}

// AddUser adds a user to the database.
func (u *Users) Add(email, pwHash, visitorId string) error {
	sqlStmt := `insert into users(email, password, visitor_id) values (?, ?, ?)`
	_, err := u.db.Exec(sqlStmt, email, pwHash, visitorId)
	if err != nil {
		return fmt.Errorf("addUser: %w", err)
	}

	return nil
}

// Check returns (true, nil) if a visitor ID exists in the database (false, nil) if the ID doesn't exist, and (false, err) if an error occurred.
func (u *Users) Check(visitorId string) (exists bool, err error) {
	var count int
	err = u.db.QueryRow(`select count(*) from users where visitor_id = ?`, visitorId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checkVisitorId: %w", err)
	}

	return count > 0, nil
}

// Close closes the Users store
func (u *Users) Close() error {
	return u.db.Close()
}
