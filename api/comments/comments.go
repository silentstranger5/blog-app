package comments

import (
	"blog/config"
	"blog/db/comments"
	"blog/db/posts"
	"blog/util"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

// @Summary Add a comment to the post
// @Tags comments
// @Accept json
// @Param id path int true "Post ID"
// @Param Authorization header string true "Auth Token"
// @Param comment body Comment true "Comment"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Header"
// @Failure 404 "Post Not Found"
// @Failure 405 "Method Not Allowed"
// @Router /api/comments/{id} [post]
func add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	_, err = posts.GetPost(config.DB, config.Ctx, postId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read request body:", err)
		return
	}
	defer r.Body.Close()

	var comment comments.Comment
	err = json.Unmarshal(body, &comment)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if comment.Text == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	comment.AuthorId = userId
	comment.PostId = postId

	err = comments.AddComment(config.DB, config.Ctx, comment)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post successfully commented"))
}

// @Summary Get the comment by ID
// @Tags comments
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} Comment
// @Failure 400 "Bad Request"
// @Failure 404 "Comment Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/comments/{id} [get]
func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	commentId, err := strconv.Atoi(pathVal)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	comment, err := comments.GetComment(config.DB, config.Ctx, commentId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(comment)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Update the comment
// @Tags comments
// @Accept json
// @Param id path int true "Comment ID"
// @Param comment body Comment true "Comment"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Token"
// @Failure 403 "No Access To Comment"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/comments/{id} [put]
func update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	commentId, err := strconv.Atoi(pathVal)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read body:", err)
		return
	}
	defer r.Body.Close()

	var comment comments.Comment
	err = json.Unmarshal(body, &comment)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if comment.Text == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	dbComment, err := comments.GetComment(config.DB, config.Ctx, commentId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if userId != dbComment.AuthorId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = comments.UpdateComment(config.DB, config.Ctx, commentId, comment)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("comment successfully updated"))
}

// @Summary Delete the comment
// @Tags comments
// @Param id path int true "Comment ID"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Token"
// @Failure 403 "No Access To Comment"
// @Failure 404 "Comment Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/comments/{id} [delete]
func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathVal := r.PathValue("id")
	if pathVal == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	commentId, err := strconv.Atoi(pathVal)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
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

	dbComment, err := comments.GetComment(config.DB, config.Ctx, commentId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if userId != dbComment.AuthorId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = comments.DeleteComment(config.DB, config.Ctx, commentId)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("comment successfully deleted"))
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{id}", get)
	mux.HandleFunc("POST /{id}", add)
	mux.HandleFunc("PUT /{id}", update)
	mux.HandleFunc("DELETE /{id}", delete)
	return mux
}
