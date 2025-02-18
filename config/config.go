package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB        *sql.DB         = nil
	Ctx       context.Context = context.TODO()
	SecretStr string          = "secret"
	Secret    []byte          = []byte(SecretStr)
	IP        string          = "localhost"
	Port      string          = "8080"
	Addr      string          = IP + ":" + Port
	Host      string          = "http://" + Addr
	DBFile    string          = "blog.db"
)

func NewDB(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitDB() error {
	data, err := os.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	schema := string(data)
	_, err = DB.ExecContext(Ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to execute query on database: %v", err)
	}
	return nil
}

func Setup() error {
	var err error
	Addr = IP + ":" + Port
	Host = "http://" + Addr
	Secret = []byte(SecretStr)
	DB, err = NewDB(DBFile)
	if err != nil {
		return err
	}
	return nil
}

func Reset() error {
	err := InitDB()
	if err != nil {
		return err
	}
	err = os.RemoveAll("static/images")
	if err != nil {
		return fmt.Errorf("failed to remove directory: %v", err)
	}
	err = os.Mkdir("static/images", 0750)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	return nil
}
