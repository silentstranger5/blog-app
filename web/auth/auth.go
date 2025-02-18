package auth

import (
	"blog/config"
	"blog/db/auth"
	"blog/util"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"
)

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		token, err := util.ParseAuthCookie(r)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		if token != "" {
			http.Redirect(w, r, "/web/posts/get", http.StatusSeeOther)
		}

		files := []string{
			"templates/base.html", "templates/auth/register.html",
		}
		err = util.Template(files, template.FuncMap{}, w, nil)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		if r.FormValue("password") != r.FormValue("confirm-password") {
			http.Error(w, "Passwords don't match", http.StatusBadRequest)
			return
		}

		user := auth.User{Username: r.FormValue("username"), Password: r.FormValue("password")}
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to convert to json:", err)
			return
		}

		body, status, err := util.Request(
			"POST", config.Host+"/api/auth/register",
			"", bytes.NewBuffer(data),
		)
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", status)
			log.Println(err)
			return
		}
		if status != http.StatusOK {
			http.Error(w, string(body), status)
			return
		}

		w.Header().Set("HX-Redirect", "/web/auth/login")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		token, err := util.ParseAuthCookie(r)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if token != "" {
			http.Redirect(w, r, "/web/posts/get", http.StatusSeeOther)
		}

		files := []string{
			"templates/base.html", "templates/auth/login.html",
		}
		err = util.Template(files, template.FuncMap{}, w, nil)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		userData := auth.User{Username: r.FormValue("username"), Password: r.FormValue("password")}
		data, err := json.Marshal(userData)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to convert to json:", err)
			return
		}

		body, status, err := util.Request(
			"POST", config.Host+"/api/auth/token",
			"", bytes.NewBuffer(data),
		)
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", status)
			log.Println(err)
			return
		}
		if status != http.StatusOK {
			http.Error(w, string(body), status)
			return
		}

		cookie := &http.Cookie{
			Name:    "Token",
			Value:   string(body),
			Expires: time.Now().Add(24 * time.Hour),
			Path:    "/",
		}

		http.SetCookie(w, cookie)
		w.Header().Set("HX-Redirect", "/web/posts/get")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie := &http.Cookie{
		Name:   "Token",
		Value:  "",
		MaxAge: 0,
		Path:   "/",
	}

	http.SetCookie(w, cookie)
	w.Header().Set("HX-Redirect", "/web/posts/get")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	return mux
}
