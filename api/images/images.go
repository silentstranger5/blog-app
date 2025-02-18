package images

import (
	"blog/config"
	"blog/db/images"
	"blog/util"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

// @Summary Upload an image
// @Tags images
// @Accept multipart/form-data
// @Param image formData file true "Image File"
// @Param Authorization header string true "Auth Header"
// @Success 200
// @Router /api/images/upload [post]
func uploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := util.ParseAuthHeader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to parse multipart form:", err)
		return
	}

	if len(r.MultipartForm.File) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to get file:", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read file:", err)
		return
	}

	err = os.WriteFile("static/images/"+header.Filename, data, 0644)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to write file:", err)
	}

	err = images.AddImage(config.DB, config.Ctx, images.Image{Name: header.Filename, AuthorId: userId})
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("image uploaded successfully"))
}

// @Summary Get images
// @Tags images
// @Produce json
// @Success 200 {object} []Image
// @Router /api/images/get [get]
func getImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	imageList, err := images.GetImages(config.DB, config.Ctx)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if imageList == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(imageList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Get Image by ID
// @Tags images
// @Produce json
// @Param id path int true "Image ID"
// @Success 200 {object} Image
// @Router /api/images/get/{id} [get]
func getImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	imageId := util.ParseUrlId(r.URL.Path)
	if imageId == 0 {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	image, err := images.GetImage(config.DB, config.Ctx, imageId)
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	data, err := json.Marshal(image)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Delete Image
// @Tags images
// @Param id path int true "Image ID"
// @Param Authorization header string true "Auth Header"
// @Success 200
// @Router /api/images/delete/{id} [delete]
func deleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := util.ParseAuthHeader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	imageId := util.ParseUrlId(r.URL.Path)
	if imageId == 0 {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	image, err := images.GetImage(config.DB, config.Ctx, imageId)
	if err == sql.ErrNoRows {
		http.Error(w, "Image Not Found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if image.AuthorId != userId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = images.DeleteImage(config.DB, config.Ctx, imageId)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = os.Remove("static/images/" + image.Name)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to remove file:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("image successfully deleted"))
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadImage)
	mux.HandleFunc("/get", getImages)
	mux.HandleFunc("/get/", getImage)
	mux.HandleFunc("/delete/", deleteImage)
	return mux
}
