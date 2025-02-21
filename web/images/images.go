package images

import (
	"blog/config"
	"blog/db/images"
	"blog/util"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to parse form:", err)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to get file:", err)
		return
	}
	defer file.Close()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("image", header.Filename)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to create form file:", err)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to copy form file:", err)
		return
	}
	writer.Close()

	url := fmt.Sprintf("%s/api/images/", config.Host)
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to create request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to upload file:", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to upload file:", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		http.Error(w, string(body), res.StatusCode)
		return
	}

	url = fmt.Sprintf("%s/api/images/", config.Host)
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

	var imageList []images.Image
	err = json.Unmarshal(body, &imageList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to unmarshal JSON:", err)
		return
	}

	files := []string{"templates/images/images.html"}
	tdata := struct {
		Images []images.Image
		UserId int
	}{imageList, userId}
	err = util.Template(files, template.FuncMap{}, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func gallery(w http.ResponseWriter, r *http.Request) {
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

	url := fmt.Sprintf("%s/api/images/", config.Host)
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

	var imageList []images.Image
	err = json.Unmarshal(body, &imageList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to unmarshal JSON:", err)
		return
	}

	files := []string{
		"templates/base.html", "templates/images/gallery.html",
	}
	tdata := struct {
		Images []images.Image
		UserId int
	}{imageList, userId}
	err = util.Template(files, template.FuncMap{}, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	imageId := util.ParseUrlId(r.URL.Path)
	if imageId == 0 {
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
	if err == http.ErrNoCookie || token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userId, err = util.ParseToken(token)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	url := fmt.Sprintf("%s/api/images/%d", config.Host, imageId)
	body, status, err := util.Request("DELETE", url, token, nil)
	if status == http.StatusInternalServerError {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if status != http.StatusOK {
		http.Error(w, string(body), status)
		return
	}

	url = fmt.Sprintf("%s/api/images/", config.Host)
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

	var imageList []images.Image
	err = json.Unmarshal(body, &imageList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to unmarshal JSON:", err)
		return
	}

	files := []string{"templates/images/images.html"}
	tdata := struct {
		Images []images.Image
		UserId int
	}{imageList, userId}
	err = util.Template(files, template.FuncMap{}, w, tdata)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/gallery", gallery)
	mux.HandleFunc("/delete/", delete)
	return mux
}
