package comments

import (
	"blog/config"
	"blog/db/comments"
	"blog/util"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"
)

func add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	postId := util.ParseUrlId(r.URL.Path)
	if postId == 0 {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	token, err := util.ParseAuthCookie(r)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == http.ErrNoCookie || token == "" {
		w.Header().Set("HX-Redirect", "/web/auth/login")
		w.WriteHeader(http.StatusSeeOther)
		w.Write([]byte("unauthorized"))
		return
	}
	userId, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to parse form:", err)
		return
	}

	comment := comments.Comment{Text: r.FormValue("comment")}
	data, err := json.Marshal(comment)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	url := fmt.Sprintf("%s/api/comments/%d", config.Host, postId)
	body, status, err := util.Request("POST", url, token, bytes.NewReader(data))
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	url = fmt.Sprintf("%s/api/posts/%d/comments", config.Host, postId)
	body, status, err = util.Request("GET", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	var commentList []comments.Comment
	err = json.Unmarshal(body, &commentList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to unmarshal JSON:", err)
		return
	}

	files := []string{"templates/comments/comments.html"}
	funcmap := template.FuncMap{
		"dateformat": func(t time.Time) string { return t.Format("2006-01-02") },
	}
	tdata := struct {
		Comments []comments.Comment
		UserId   int
	}{commentList, userId}
	err = util.Template(files, funcmap, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		commentId := util.ParseUrlId(r.URL.Path)
		if commentId == 0 {
			http.Error(w, "Invalid URL Format", http.StatusBadRequest)
			return
		}

		token, err := util.ParseAuthCookie(r)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if err == http.ErrNoCookie || token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userId, err := util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		url := fmt.Sprintf("%s/api/comments/%d", config.Host, commentId)
		body, status, err := util.Request("GET", url, token, nil)
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", status)
			log.Println(err)
			return
		}
		if status != http.StatusOK {
			http.Error(w, string(body), status)
			return
		}

		var comment comments.Comment
		err = json.Unmarshal(body, &comment)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to unmarshal JSON:", err)
			return
		}
		if comment.AuthorId != userId {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		files := []string{"templates/comments/update.html"}
		tdata := struct{ Comment comments.Comment }{comment}
		err = util.Template(files, template.FuncMap{}, w, tdata)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	} else if r.Method == http.MethodPost {
		commentId := util.ParseUrlId(r.URL.Path)
		if commentId == 0 {
			http.Error(w, "Invalid URL Format", http.StatusBadRequest)
			return
		}

		token, err := util.ParseAuthCookie(r)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if err == http.ErrNoCookie || token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userId, err := util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to parse form:", err)
			return
		}

		comment := comments.Comment{Text: r.FormValue("comment")}
		data, err := json.Marshal(comment)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to marshal JSON:", err)
			return
		}

		url := fmt.Sprintf("%s/api/comments/%d", config.Host, commentId)
		body, status, err := util.Request("PUT", url, token, bytes.NewBuffer(data))
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", status)
			log.Println(err)
			return
		}
		if status != http.StatusOK {
			http.Error(w, string(body), status)
			return
		}

		url = fmt.Sprintf("%s/api/comments/%d", config.Host, commentId)
		body, status, err = util.Request("GET", url, token, nil)
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", status)
			log.Println(err)
			return
		}
		if status != http.StatusOK {
			http.Error(w, string(body), status)
			return
		}

		err = json.Unmarshal(body, &comment)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to unmarshal JSON:", err)
			return
		}

		funcmap := template.FuncMap{
			"dateformat": func(t time.Time) string { return t.Format("2006-01-02") },
		}
		files := []string{"templates/comments/comment.html"}
		tdata := struct {
			Comment comments.Comment
			UserId  int
		}{comment, userId}
		err = util.Template(files, funcmap, w, tdata)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	commentId := util.ParseUrlId(r.URL.Path)
	if commentId == 0 {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	token, err := util.ParseAuthCookie(r)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == http.ErrNoCookie || token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	url := fmt.Sprintf("%s/api/comments/%d", config.Host, commentId)
	body, status, err := util.Request("GET", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	var comment comments.Comment
	err = json.Unmarshal(body, &comment)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to unmarshal JSON:", err)
		return
	}

	url = fmt.Sprintf("%s/api/comments/%d", config.Host, commentId)
	body, status, err = util.Request("DELETE", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	url = fmt.Sprintf("%s/api/posts/%d/comments", config.Host, comment.PostId)
	body, status, err = util.Request("GET", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		http.Error(w, string(body), status)
		return
	}

	var commentList []comments.Comment
	if status != http.StatusNotFound {
		err = json.Unmarshal(body, &commentList)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to unmarshal JSON:", err)
			return
		}
	}

	funcmap := template.FuncMap{
		"dateformat": func(t time.Time) string { return t.Format("2006-01-02") },
	}
	files := []string{"templates/comments/comments.html"}
	tdata := struct {
		Comments []comments.Comment
		UserId   int
	}{commentList, userId}
	err = util.Template(files, funcmap, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
	}
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/add/", add)
	mux.HandleFunc("/update/", update)
	mux.HandleFunc("/delete/", delete)
	return mux
}
