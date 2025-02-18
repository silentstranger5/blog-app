package auth_test

import (
	"blog/config"
	"blog/db/auth"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	err := os.Chdir("../../..")
	if err != nil {
		panic(err)
	}
	config.DBFile = ":memory:"
	err = config.Setup()
	if err != nil {
		panic(err)
	}
	err = config.InitDB()
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestAddUser(t *testing.T) {
	tests := []struct {
		user  auth.User
		error bool
	}{
		{auth.User{Username: "user", Password: "password"}, false},
		{auth.User{Username: "guest", Password: "password"}, false},
		{auth.User{Username: "user", Password: "password"}, true},
		{auth.User{Username: "user"}, true},
		{auth.User{Password: "password"}, true},
		{auth.User{}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := auth.AddUser(config.DB, config.Ctx, test.user)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		user  auth.User
		error bool
	}{
		{auth.User{Id: 1, Username: "user", Password: "password"}, false},
		{auth.User{Id: 2, Username: "guest", Password: "password"}, false},
		{auth.User{Id: 3, Username: "johndoe"}, true},
		{auth.User{Id: 1}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			user, err := auth.GetUser(config.DB, config.Ctx, test.user.Username)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
			if err != nil {
				return
			}
			if user != test.user {
				t.Fatalf("test failed: %v", user)
			}
		})
	}
}
