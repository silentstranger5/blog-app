package comments_test

import (
	"blog/config"
	"blog/db/auth"
	"blog/db/comments"
	"blog/db/posts"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
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

func TestAddComment(t *testing.T) {
	tests := []struct {
		comment comments.Comment
		error   bool
	}{
		{comments.Comment{AuthorId: 1, PostId: 1, Text: "First Comment"}, false},
		{comments.Comment{AuthorId: 1, PostId: 1, Text: "Second Comment"}, false},
		{comments.Comment{AuthorId: 2, PostId: 2, Text: "Third Comment"}, false},
		{comments.Comment{AuthorId: 1, PostId: 2, Text: "Fourth Comment"}, false},
		{comments.Comment{AuthorId: 1, PostId: 3, Text: "New Comment"}, true},
		{comments.Comment{AuthorId: 3, PostId: 1, Text: "New Comment"}, true},
		{comments.Comment{AuthorId: 3, PostId: 3, Text: "New Comment"}, true},
		{comments.Comment{AuthorId: 1, PostId: 1}, true},
		{comments.Comment{AuthorId: 1}, true},
		{comments.Comment{PostId: 1}, true},
		{comments.Comment{}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := comments.AddComment(config.DB, config.Ctx, test.comment)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetComments(t *testing.T) {
	tests := []struct {
		postid   int
		comments []comments.Comment
	}{
		{1, []comments.Comment{
			{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "First Comment"},
			{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "Second Comment"},
		}},
		{2, []comments.Comment{
			{Id: 3, AuthorId: 2, PostId: 2, Author: "guest", Text: "Third Comment"},
			{Id: 4, AuthorId: 1, PostId: 2, Author: "user", Text: "Fourth Comment"},
		}},
		{3, nil},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			commentList, err := comments.GetComments(config.DB, config.Ctx, test.postid)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			for i := range commentList {
				commentList[i].Created = zero
			}
			if !reflect.DeepEqual(commentList, test.comments) {
				t.Fatalf("test failed: %v", commentList)
			}
		})
	}
}

func TestGetComment(t *testing.T) {
	tests := []struct {
		comment comments.Comment
		error   bool
	}{
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "First Comment"}, false},
		{comments.Comment{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "Second Comment"}, false},
		{comments.Comment{Id: 3, AuthorId: 2, PostId: 2, Author: "guest", Text: "Third Comment"}, false},
		{comments.Comment{Id: 4, AuthorId: 1, PostId: 2, Author: "user", Text: "Fourth Comment"}, false},
		{comments.Comment{Id: 5}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			comment, err := comments.GetComment(config.DB, config.Ctx, test.comment.Id)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
			if err != nil {
				return
			}
			var zero time.Time
			comment.Created = zero
			if comment != test.comment {
				t.Fatalf("test failed: %v", comment)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	tests := []struct {
		comment comments.Comment
	}{
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "New First Comment"}},
		{comments.Comment{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "New Second Comment"}},
		{comments.Comment{Id: 3, AuthorId: 2, PostId: 2, Author: "guest", Text: "New Third Comment"}},
		{comments.Comment{Id: 4, AuthorId: 1, PostId: 2, Author: "user", Text: "New Fourth Comment"}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := comments.UpdateComment(config.DB, config.Ctx, test.comment.Id, test.comment)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			comment, err := comments.GetComment(config.DB, config.Ctx, test.comment.Id)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			comment.Created = zero
			if comment != test.comment {
				t.Fatalf("test failed: %v", comment)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	for i := range 5 {
		err := comments.DeleteComment(config.DB, config.Ctx, i+1)
		if err != nil {
			t.Fatalf("test failed: %v", err)
		}
		comment, err := comments.GetComment(config.DB, config.Ctx, i+1)
		if err == nil {
			t.Fatalf("test failed: %v", comment)
		}
	}
}
