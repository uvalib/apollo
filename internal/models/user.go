package models

import (
	"errors"
	"fmt"
	"time"
)

// User : defines a serials management system user
type User struct {
	ID          int64     `json:"id"`
	ComputingID string    `db:"computing_id" json:"computingId"`
	FirstName   string    `db:"first_name" json:"firstName"`
	LastName    string    `db:"last_name" json:"lastName"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `db:"created_at" json:"-"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}

// AllUsers returns a pointer to a list of all users
//
func (db *DB) AllUsers() []User {
	users := []User{}
	db.Select(&users, "select * from users order by last_name asc")
	return users
}

// FindUserBy will find a user by a specific key: id, computing_id, email, first_name or last_name
//
func (db *DB) FindUserBy(key string, value string) (*User, error) {
	keys := []string{"id", "computing_id", "first_name", "last_name", "email"}
	validKey := false
	for _, v := range keys {
		if v == key {
			validKey = true
		}
	}
	if !validKey {
		return nil, errors.New("Invalid search key")
	}
	user := User{}
	qs := fmt.Sprintf("SELECT * FROM users WHERE %s=?", key)
	err := db.Get(&user, qs, value)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
