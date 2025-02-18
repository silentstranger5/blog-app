package web

import (
	"blog/web/auth"
	"blog/web/comments"
	"blog/web/images"
	"blog/web/posts"
	"net/http"
)

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	postsMux := posts.ServeMux()
	authMux := auth.ServeMux()
	commentsMux := comments.ServeMux()
	imagesMux := images.ServeMux()
	mux.Handle("/posts/", http.StripPrefix("/posts", postsMux))
	mux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	mux.Handle("/comments/", http.StripPrefix("/comments", commentsMux))
	mux.Handle("/images/", http.StripPrefix("/images", imagesMux))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}
