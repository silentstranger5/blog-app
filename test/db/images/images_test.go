package images_test

import (
	"blog/config"
	"blog/db/auth"
	"blog/db/images"
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
	m.Run()
}

func TestAddImage(t *testing.T) {
	tests := []struct {
		image images.Image
		error bool
	}{
		{images.Image{AuthorId: 1, Name: "photo.jpg"}, false},
		{images.Image{AuthorId: 1, Name: "photo.jpg"}, false},
		{images.Image{AuthorId: 1, Name: "picture.jpg"}, false},
		{images.Image{AuthorId: 2, Name: "picture.png"}, false},
		{images.Image{AuthorId: 2, Name: "landscape.jpg"}, false},
		{images.Image{AuthorId: 3, Name: "ethereal.jpg"}, true},
		{images.Image{AuthorId: 1}, true},
		{images.Image{}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := images.AddImage(config.DB, config.Ctx, test.image)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetImages(t *testing.T) {
	imageList := []images.Image{
		{Id: 1, AuthorId: 1, Name: "photo.jpg"},
		{Id: 3, AuthorId: 1, Name: "picture.jpg"},
		{Id: 4, AuthorId: 2, Name: "picture.png"},
		{Id: 5, AuthorId: 2, Name: "landscape.jpg"},
	}
	dbImages, err := images.GetImages(config.DB, config.Ctx)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
	var zero time.Time
	for i := range dbImages {
		dbImages[i].Created = zero
	}
	if !reflect.DeepEqual(imageList, dbImages) {
		t.Fatalf("test failed: %v", dbImages)
	}
}

func TestGetImage(t *testing.T) {
	tests := []struct {
		image images.Image
		error bool
	}{
		{images.Image{Id: 1, AuthorId: 1, Name: "photo.jpg"}, false},
		{images.Image{Id: 3, AuthorId: 1, Name: "picture.jpg"}, false},
		{images.Image{Id: 4, AuthorId: 2, Name: "picture.png"}, false},
		{images.Image{Id: 5, AuthorId: 2, Name: "landscape.jpg"}, false},
		{images.Image{Id: 6, AuthorId: 1}, true},
		{images.Image{Id: 7}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			image, err := images.GetImage(config.DB, config.Ctx, test.image.Id)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
			if err != nil {
				return
			}
			var zero time.Time
			image.Created = zero
			if image != test.image {
				t.Fatalf("test failed: %v", image)
			}
		})
	}
}

func TestDeleteImage(t *testing.T) {
	for i := range 6 {
		err := images.DeleteImage(config.DB, config.Ctx, i+1)
		if err != nil {
			t.Fatalf("test failed: %v", err)
		}
		image, err := images.GetImage(config.DB, config.Ctx, i+1)
		if err == nil {
			t.Fatalf("test failed: %v", image)
		}
	}
}
