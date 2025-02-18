package posts

import (
	"blog/config"
	"blog/db/comments"
	"blog/db/posts"
	"blog/db/tags"
	"blog/util"
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"
)

func getPosts(w http.ResponseWriter, r *http.Request) {
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

	body, status, err := util.Request(
		"GET", config.Host+"/api/posts/get",
		token, nil,
	)
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

func getPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	postId := util.ParseUrlId(r.URL.Path)
	if postId == 0 {
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

	body, status, err := util.Request(
		"GET", config.Host+"/api/posts/get/"+strconv.Itoa(postId),
		token, nil,
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

	body, status, err = util.Request(
		"GET", config.Host+"/api/comments/post/"+strconv.Itoa(postId),
		token, nil,
	)
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

	body, status, err = util.Request(
		"GET", config.Host+"/api/tags/get/"+strconv.Itoa(postId),
		token, nil,
	)
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

func addPost(w http.ResponseWriter, r *http.Request) {
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

		body, status, err := util.Request(
			"POST", config.Host+"/api/posts/add",
			token, bytes.NewBuffer(data),
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

		w.Header().Set("HX-Redirect", "/web/posts/get")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userId, err := util.ParseToken(token)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		body, status, err := util.Request(
			"GET", config.Host+"/api/posts/get/"+strconv.Itoa(postId),
			token, nil,
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

		body, status, err = util.Request(
			"GET", config.Host+"/api/tags/get/"+strconv.Itoa(postId),
			token, nil,
		)
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
			Tags   []tags.Tag
			UserId int
		}{post, tagList, userId}
		err = util.Template(files, funcmap, w, tdata)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	} else if r.Method == http.MethodPost {
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

		body, status, err := util.Request(
			"PUT", config.Host+"/api/posts/update/"+strconv.Itoa(postId),
			token, bytes.NewBuffer(data),
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

		w.Header().Set("HX-Redirect", "/web/posts/get/"+strconv.Itoa(postId))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	body, status, err := util.Request(
		"DELETE", config.Host+"/api/posts/delete/"+strconv.Itoa(postId),
		token, bytes.NewBuffer(data),
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

	w.Header().Set("HX-Redirect", "/web/posts/get")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func likePost(w http.ResponseWriter, r *http.Request) {
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

	body, status, err := util.Request(
		"POST", config.Host+"/api/posts/like/"+strconv.Itoa(postId),
		token, nil,
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

	body, status, err = util.Request(
		"GET", config.Host+"/api/posts/likes/"+strconv.Itoa(postId),
		token, nil,
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func dislikePost(w http.ResponseWriter, r *http.Request) {
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

	body, status, err := util.Request(
		"POST", config.Host+"/api/posts/dislike/"+strconv.Itoa(postId),
		token, nil,
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

	body, status, err = util.Request(
		"GET", config.Host+"/api/posts/likes/"+strconv.Itoa(postId),
		token, nil,
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func filterTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if !(len(parts) == 3 && parts[2] != "") {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	tagName := parts[2]

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

	body, status, err := util.Request(
		"GET", config.Host+"/api/tags/posts/"+tagName,
		token, nil,
	)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	var postList []posts.Post
	err = json.Unmarshal(body, &postList)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
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

func searchPosts(w http.ResponseWriter, r *http.Request) {
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

	body, status, err := util.Request(
		"GET", config.Host+"/api/posts/search/"+query,
		token, nil,
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

	var postList []posts.Post
	err = json.Unmarshal(body, &postList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to unmarshal JSON:", err)
		return
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
	mux.HandleFunc("/get", getPosts)
	mux.HandleFunc("/add", addPost)
	mux.HandleFunc("/get/", getPost)
	mux.HandleFunc("/update/", updatePost)
	mux.HandleFunc("/delete/", deletePost)
	mux.HandleFunc("/like/", likePost)
	mux.HandleFunc("/dislike/", dislikePost)
	mux.HandleFunc("/tag/", filterTag)
	mux.HandleFunc("/search", searchPosts)
	return mux
}
