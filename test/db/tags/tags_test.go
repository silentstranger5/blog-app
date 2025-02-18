package tags_test

import (
	"blog/config"
	"blog/db/auth"
	"blog/db/posts"
	"blog/db/tags"
	"fmt"
	"os"
	"reflect"
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
	users := []auth.User{
		{Id: 1, Username: "user", Password: "password"},
		{Id: 2, Username: "guest", Password: "password"},
	}
	for _, user := range users {
		err = auth.AddUser(config.DB, config.Ctx, user)
		if err != nil {
			panic(err)
		}
	}
	postList := []posts.Post{
		{AuthorId: 1, Title: "New Post", Text: "Post Text"},
		{AuthorId: 1, Title: "New Post", Text: "Post Text"},
		{AuthorId: 2, Title: "New Post", Text: "Post Text"},
	}
	for _, post := range postList {
		_, err = posts.AddPost(config.DB, config.Ctx, post)
		if err != nil {
			panic(err)
		}
	}
	m.Run()
}

func TestAddTags(t *testing.T) {
	tests := []struct {
		postid int
		tags   []tags.Tag
		error  bool
	}{
		{1, []tags.Tag{{Name: "first"}, {Name: "second"}}, false},
		{2, []tags.Tag{{Name: "first"}, {Name: "third"}}, false},
		{4, []tags.Tag{{Name: "first"}, {Name: "third"}}, true},
		{1, []tags.Tag{{Name: "first"}, {}}, true},
		{1, nil, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := tags.AddTags(config.DB, config.Ctx, test.postid, test.tags)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	tests := []struct {
		postid int
		tags   []tags.Tag
	}{
		{1, []tags.Tag{{Name: "first"}, {Name: "second"}}},
		{2, []tags.Tag{{Name: "first"}, {Name: "third"}}},
		{3, nil},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tags, err := tags.GetTags(config.DB, config.Ctx, test.postid)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if !reflect.DeepEqual(tags, test.tags) {
				t.Fatalf("test failed: %v", tags)
			}
		})
	}
}

func TestUpdateTags(t *testing.T) {
	tests := []struct {
		postid int
		tags   []tags.Tag
	}{
		{1, []tags.Tag{{Name: "first"}, {Name: "second"}}},
		{2, []tags.Tag{{Name: "first"}, {Name: "third"}}},
		{3, []tags.Tag{{Name: "second"}, {Name: "third"}}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := tags.UpdateTags(config.DB, config.Ctx, test.postid, test.tags)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			tags, err := tags.GetTags(config.DB, config.Ctx, test.postid)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if !reflect.DeepEqual(tags, test.tags) {
				t.Fatalf("test failed: %v", tags)
			}
		})
	}
}

func TestDeleteTags(t *testing.T) {
	for i := range 3 {
		err := tags.DeleteTags(config.DB, config.Ctx, i+1)
		if err != nil {
			t.Fatalf("test failed: %v", err)
		}
		tags, err := tags.GetTags(config.DB, config.Ctx, i+1)
		if err != nil {
			t.Fatalf("test failed: %v", err)
		}
		if tags != nil {
			t.Fatalf("test faied: %v", tags)
		}
	}
}
