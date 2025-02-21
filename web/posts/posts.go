package posts

import (
	"blog/config"
	"blog/db/comments"
	"blog/db/posts"
	"blog/db/tags"
	"blog/util"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"
)

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var userId int
	token, err := util.ParseAuthCookie(r)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if token != "" {
		userId, err = util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	url := fmt.Sprintf("%s/api/posts/", config.Host)
	body, status, err := util.Request("GET", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		http.Error(w, string(body), status)
		return
	}

	var postList []posts.Post
	if status != http.StatusNotFound {
		err = json.Unmarshal(body, &postList)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	for i, post := range postList {
		lines := strings.Split(post.Text, "\n")
		var text string
		if len(lines) > 5 {
			text = strings.Join(lines[:5], "\n")
		} else {
			text = strings.Join(lines, "\n")
		}
		text = strings.ReplaceAll(text, "\r\n", "\n")
		markdown := string(blackfriday.Run(
			[]byte(text),
			blackfriday.WithExtensions(
				blackfriday.CommonExtensions|
					blackfriday.HardLineBreak,
			),
		))
		markdown = strings.ReplaceAll(markdown, "<img ", "<img class=\"d-none\" ")
		postList[i].Text = markdown
	}
	path := "/web/posts" + r.URL.String()

	files := []string{
		"templates/base.html", "templates/posts/posts.html",
	}
	funcmap := template.FuncMap{
		"dateformat": func(t time.Time) string { return t.Format("January 2, 2006") },
		"html":       func(s string) template.HTML { return template.HTML(s) },
	}
	tdata := struct {
		Posts  []posts.Post
		UserId int
		Path   string
	}{postList, userId, path}
	err = util.Template(files, funcmap, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func getId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}
	postId, err := strconv.Atoi(pathVal)
	if err != nil {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	var userId int
	token, err := util.ParseAuthCookie(r)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if token != "" {
		userId, err = util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	path := fmt.Sprintf("%s/api/posts/%d", config.Host, postId)
	body, status, err := util.Request("GET", path, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	var post posts.Post
	err = json.Unmarshal(body, &post)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	post.Text = strings.ReplaceAll(post.Text, "\r\n", "\n")
	markdown := blackfriday.Run(
		[]byte(post.Text),
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions|
				blackfriday.HardLineBreak,
		),
	)
	post.Text = string(markdown)

	path = fmt.Sprintf("%s/api/posts/%d/comments", config.Host, postId)
	body, status, err = util.Request("GET", path, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
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

	path = fmt.Sprintf("%s/api/posts/%d/tags", config.Host, postId)
	body, status, err = util.Request("GET", path, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		http.Error(w, string(body), status)
		return
	}

	var tagList []tags.Tag
	if status != http.StatusNotFound {
		err = json.Unmarshal(body, &tagList)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to unmarshal JSON:", err)
			return
		}
	}

	files := []string{
		"templates/base.html", "templates/posts/post.html",
	}
	funcmap := template.FuncMap{
		"split":      func(text string) []string { return strings.Split(text, "\n") },
		"dateformat": func(t time.Time, format string) string { return t.Format(format) },
		"escape":     func(s string) string { return url.QueryEscape(s) },
		"html":       func(s string) template.HTML { return template.HTML(s) },
	}
	tdata := struct {
		Post     posts.Post
		Comments []comments.Comment
		Tags     []tags.Tag
		UserId   int
	}{post, commentList, tagList, userId}
	err = util.Template(files, funcmap, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		token, err := util.ParseAuthCookie(r)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if err == http.ErrNoCookie || token == "" {
			http.Redirect(w, r, "/web/auth/login", http.StatusSeeOther)
			return
		}
		userId, err := util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		files := []string{
			"templates/base.html", "templates/posts/add.html",
		}
		tdata := struct{ UserId int }{userId}
		err = util.Template(files, template.FuncMap{}, w, tdata)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	} else if r.Method == http.MethodPost {
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

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to parse form:", err)
			return
		}

		var tagList []tags.Tag
		if r.FormValue("tags") != "" {
			for _, tagName := range strings.Split(r.FormValue("tags"), ", ") {
				tagList = append(tagList, tags.Tag{Name: tagName})
			}
		}

		post := posts.Post{Title: r.FormValue("title"), Text: r.FormValue("text"),
			Tags: tagList}
		data, err := json.Marshal(post)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to convert to json:", err)
			return
		}

		url := fmt.Sprintf("%s/api/posts/", config.Host)
		body, status, err := util.Request("POST", url, token, bytes.NewBuffer(data))
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", status)
			log.Println(err)
			return
		}
		if status != http.StatusOK {
			http.Error(w, string(body), status)
			return
		}

		w.Header().Set("HX-Redirect", "/web/posts/get")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pathVal := r.PathValue("id")
		if pathVal == "" {
			http.Error(w, "Invalid URL Format", http.StatusBadRequest)
			return
		}
		postId, err := strconv.Atoi(pathVal)
		if err != nil {
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

		url := fmt.Sprintf("%s/api/posts/%d", config.Host, postId)
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

		var post posts.Post
		err = json.Unmarshal(body, &post)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if post.AuthorId != userId {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		url = fmt.Sprintf("%s/api/posts/%d/tags", config.Host, postId)
		body, status, err = util.Request("GET", url, token, nil)
		if status == http.StatusInternalServerError {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if status != http.StatusOK && status != http.StatusNotFound {
			http.Error(w, string(body), status)
			return
		}

		var tagList []tags.Tag
		if status != http.StatusNotFound {
			err = json.Unmarshal(body, &tagList)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				log.Println("failed to unmarshal JSON:", err)
				return
			}
		}
		post.Tags = tagList

		files := []string{
			"templates/base.html", "templates/posts/update.html",
		}
		funcmap := template.FuncMap{
			"join": func(tagList []tags.Tag) string {
				tagNames := make([]string, 0)
				for _, tag := range tagList {
					tagNames = append(tagNames, tag.Name)
				}
				return strings.Join(tagNames, ", ")
			},
		}
		tdata := struct {
			Post   posts.Post
			UserId int
		}{post, userId}
		err = util.Template(files, funcmap, w, tdata)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	} else if r.Method == http.MethodPost {
		pathVal := r.PathValue("id")
		if pathVal == "" {
			http.Error(w, "Invalid URL Format", http.StatusBadRequest)
			return
		}
		postId, err := strconv.Atoi(pathVal)
		if err != nil {
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

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		var tagList []tags.Tag
		if r.FormValue("tags") != "" {
			tagNames := strings.Split(r.FormValue("tags"), ", ")
			for _, tagName := range tagNames {
				tagList = append(tagList, tags.Tag{Name: tagName})
			}
		}

		post := posts.Post{Title: r.FormValue("title"), Text: r.FormValue("text"),
			Tags: tagList}
		data, err := json.Marshal(post)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to convert to json:", err)
			return
		}

		url := fmt.Sprintf("%s/api/posts/%d", config.Host, postId)
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

		url = fmt.Sprintf("/web/posts/get/%d", postId)
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}
	postId, err := strconv.Atoi(pathVal)
	if err != nil {
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

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	post := posts.Post{Title: r.FormValue("title"), Text: r.FormValue("text")}
	data, err := json.Marshal(post)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to convert to json:", err)
		return
	}

	url := fmt.Sprintf("%s/api/posts/%d", config.Host, postId)
	body, status, err := util.Request("DELETE", url, token, bytes.NewBuffer(data))
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	w.Header().Set("HX-Redirect", "/web/posts/get")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func like(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}
	postId, err := strconv.Atoi(pathVal)
	if err != nil {
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

	url := fmt.Sprintf("%s/api/posts/%d/like", config.Host, postId)
	body, status, err := util.Request("POST", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	url = fmt.Sprintf("%s/api/posts/%d/likes", config.Host, postId)
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func dislike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}
	postId, err := strconv.Atoi(pathVal)
	if err != nil {
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

	url := fmt.Sprintf("%s/api/posts/%d/dislike", config.Host, postId)
	body, status, err := util.Request("POST", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	url = fmt.Sprintf("%s/api/posts/%d/likes", config.Host, postId)
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func tag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	tagName := r.PathValue("name")
	if tagName == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	var userId int
	token, err := util.ParseAuthCookie(r)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if token != "" {
		userId, err = util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	url := fmt.Sprintf("%s/api/posts/tags/t/%s", config.Host, tagName)
	body, status, err := util.Request("GET", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		http.Error(w, string(body), status)
		return
	}

	var postList []posts.Post
	if status != http.StatusNotFound {
		err = json.Unmarshal(body, &postList)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	for i, post := range postList {
		lines := strings.Split(post.Text, "\n")
		var text string
		if len(lines) > 5 {
			text = strings.Join(lines[:5], "\n")
		} else {
			text = strings.Join(lines, "\n")
		}
		text = strings.ReplaceAll(text, "\r\n", "\n")
		markdown := string(blackfriday.Run(
			[]byte(text),
			blackfriday.WithExtensions(
				blackfriday.CommonExtensions|
					blackfriday.HardLineBreak,
			),
		))
		markdown = strings.ReplaceAll(markdown, "<img ", "<img class=\"d-none\" ")
		postList[i].Text = markdown
	}
	path := "/web/posts" + r.URL.String()

	files := []string{
		"templates/base.html", "templates/posts/posts.html",
	}
	funcmap := template.FuncMap{
		"dateformat": func(t time.Time) string { return t.Format("January 2, 2006") },
		"html":       func(s string) template.HTML { return template.HTML(s) },
	}
	tdata := struct {
		Posts  []posts.Post
		UserId int
		Path   string
	}{postList, userId, path}
	err = util.Template(files, funcmap, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var userId int
	token, err := util.ParseAuthCookie(r)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if token != "" {
		userId, err = util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("%s/api/posts/search/q/%s", config.Host, query)
	body, status, err := util.Request("GET", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", status)
		log.Println(err)
		return
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		http.Error(w, string(body), status)
		return
	}

	var postList []posts.Post
	if status != http.StatusNotFound {
		err = json.Unmarshal(body, &postList)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println("failed to unmarshal JSON:", err)
			return
		}
	}

	for i, post := range postList {
		lines := strings.Split(post.Text, "\n")
		var text string
		if len(lines) > 5 {
			text = strings.Join(lines[:5], "\n")
		} else {
			text = strings.Join(lines, "\n")
		}
		text = strings.ReplaceAll(text, "\r\n", "\n")
		markdown := string(blackfriday.Run(
			[]byte(text),
			blackfriday.WithExtensions(
				blackfriday.CommonExtensions|
					blackfriday.HardLineBreak,
			),
		))
		markdown = strings.ReplaceAll(markdown, "<img ", "<img class=\"d-none\" ")
		postList[i].Text = markdown
	}
	path := "/web/posts" + r.URL.String()

	files := []string{
		"templates/base.html", "templates/posts/posts.html",
	}
	funcmap := template.FuncMap{
		"dateformat": func(t time.Time) string { return t.Format("January 2, 2006") },
		"html":       func(s string) template.HTML { return template.HTML(s) },
	}
	tdata := struct {
		Posts  []posts.Post
		UserId int
		Path   string
	}{postList, userId, path}
	err = util.Template(files, funcmap, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/add", add)
	mux.HandleFunc("/update/{id}", update)
	mux.HandleFunc("GET /get", get)
	mux.HandleFunc("GET /get/{id}", getId)
	mux.HandleFunc("DELETE /delete/{id}", delete)
	mux.HandleFunc("POST /like/{id}", like)
	mux.HandleFunc("POST /dislike/{id}", dislike)
	mux.HandleFunc("GET /tag/{name}", tag)
	mux.HandleFunc("GET /search", search)
	return mux
}
