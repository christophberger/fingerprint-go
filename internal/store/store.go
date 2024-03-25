package store

import (
	"database/sql"
	"fmt"
	"time"

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
   	email text not null unique,
   	signup_fingerprint text,
	timestamp text
   );
   `

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("createTable: %w", err)
	}

	return nil
}

// AddUser adds a user to the database.
func (u *Users) Add(email, visitorId string) (status string, err error) {
	timestamp := time.Now().Format(time.RFC3339)
	sqlStmt := `insert into users(email, signup_fingerprint, timestamp) values (?, ?, ?)`
	_, err = u.db.Exec(sqlStmt, email, visitorId, timestamp)
	if err != nil {
		if err.Error()[0:17] == "constraint failed" {
			return "You already have signed up", nil
		}
		return "", fmt.Errorf("Users.Add: %w", err)
	}

	return "Thank you for signing up!", nil
}

// Check returns (true, nil) if another user already has signed up on the same device within the last minute.
func (u *Users) Check(visitorId string) (recentLogin bool, err error) {
	var signupTime string

	err = u.db.QueryRow(`select timestamp from users 
		where signup_fingerprint = ? 
		order by timestamp desc limit 1`, visitorId).Scan(&signupTime)
	if err == sql.ErrNoRows {
		// no previous signup for this fingerprint
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Users.Check: %w", err)
	}

	st, err := time.Parse(time.RFC3339, signupTime)
	if err != nil {
		return false, fmt.Errorf("Users.Check: %w", err)
	}
	if time.Since(st) < time.Minute {
		return true, nil
	}
	return

}

// Close closes the Users store
func (u *Users) Close() error {
	return u.db.Close()
}
