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
)

// @Summary Add a comment to the post
// @Tags comments
// @Accept json
// @Param id path int true "Post ID"
// @Param Authorization header string true "Auth Token"
// @Param comment body Comment true "Comment"
// @Success 200
// @Router /api/comments/add/{id} [post]
func AddComment(w http.ResponseWriter, r *http.Request) {
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

	postId := util.ParseUrlId(r.URL.Path)
	if postId == 0 {
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

// @Summary Get comments for the post
// @Tags comments
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} []Comment
// @Router /api/comments/post/{id} [get]
func GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postId := util.ParseUrlId(r.URL.Path)
	if postId == 0 {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	commentList, err := comments.GetComments(config.DB, config.Ctx, postId)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if commentList == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(&commentList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Get the comment by ID
// @Tags comments
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} Comment
// @Router /api/comments/get/{id} [get]
func GetComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	commentId := util.ParseUrlId(r.URL.Path)
	if commentId == 0 {
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
// @Router /api/comments/update/{id} [post]
func UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	commentId := util.ParseUrlId(r.URL.Path)
	if commentId == 0 {
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
// @Router /api/comments/delete/{id} [delete]
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	commentId := util.ParseUrlId(r.URL.Path)
	if commentId == 0 {
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
	mux.HandleFunc("/add/", AddComment)
	mux.HandleFunc("/get/", GetComment)
	mux.HandleFunc("/post/", GetComments)
	mux.HandleFunc("/update/", UpdateComment)
	mux.HandleFunc("/delete/", DeleteComment)
	return mux
}
