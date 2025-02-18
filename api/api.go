package api

import (
	"blog/api/auth"
	"blog/api/comments"
	"blog/api/images"
	"blog/api/posts"
	"blog/api/tags"
	"net/http"
)

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	postsMux := posts.ServeMux()
	authMux := auth.ServeMux()
	tagsMux := tags.ServeMux()
	commentsMux := comments.ServeMux()
	imagesMux := images.ServeMux()
	mux.Handle("/posts/", http.StripPrefix("/posts", postsMux))
	mux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	mux.Handle("/tags/", http.StripPrefix("/tags", tagsMux))
	mux.Handle("/comments/", http.StripPrefix("/comments", commentsMux))
	mux.Handle("/images/", http.StripPrefix("/images", imagesMux))
	return mux
}
