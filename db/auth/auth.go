package auth

import (
	"context"
	"database/sql"
	"fmt"
)

type User struct {
	Id       int
	Username string
	Password string
}

func AddUser(db *sql.DB, ctx context.Context, user User) error {
	if user.Username == "" || user.Password == "" {
		return fmt.Errorf("invalid argument")
	}
	_, err := db.ExecContext(
		ctx,
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		user.Username, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(db *sql.DB, ctx context.Context, username string) (User, error) {
	if username == "" {
		return User{}, fmt.Errorf("invalid argument")
	}
	var user User
	err := db.QueryRowContext(
		ctx,
		"SELECT * FROM users WHERE username = $1",
		username,
	).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
