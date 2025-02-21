package posts

import (
	"blog/config"
	"blog/db/comments"
	"blog/db/likes"
	"blog/db/posts"
	"blog/db/tags"
	"blog/util"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// @Summary Add a new post
// @Tags posts
// @Accept json
// @Param post body Post true "Post"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Header"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/ [post]
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read request body:", err)
		return
	}
	defer r.Body.Close()

	var post posts.Post
	err = json.Unmarshal(body, &post)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if post.Title == "" || post.Text == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	post.AuthorId = userId

	postId, err := posts.AddPost(config.DB, config.Ctx, post)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if post.Tags != nil {
		for _, tag := range post.Tags {
			if tag.Name == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		}
		tagMap := make(map[tags.Tag]bool)
		tagList := make([]tags.Tag, 0)
		for _, tag := range post.Tags {
			tagMap[tag] = true
		}
		for _, tag := range post.Tags {
			_, ok := tagMap[tag]
			if ok {
				tagList = append(tagList, tag)
				delete(tagMap, tag)
			}
		}
		post.Tags = tagList
		err = tags.AddTags(config.DB, config.Ctx, postId, post.Tags)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post successfully created"))
}

// @Summary Get post list
// @Tags posts
// @Produce json
// @Success 200 {object} []Post
// @Failure 400 "Bad Request"
// @Failure 404 "Posts Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/ [get]
func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	posts, err := posts.GetPosts(config.DB, config.Ctx)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if posts == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(&posts)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Get a post by ID
// @Tags posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} []Post
// @Failure 400 "Bad Request"
// @Failure 404 "Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id} [get]
func getId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	post, err := posts.GetPost(config.DB, config.Ctx, postId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(&post)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal json:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Update a post
// @Tags posts
// @Accept json
// @Param id path int true "Post ID"
// @Param post body Post true "Post"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Token"
// @Failure 403 "No Access To Post"
// @Failure 404 "Post Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id} [put]
func update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	post, err := posts.GetPost(config.DB, config.Ctx, postId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if post.AuthorId != userId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Println(err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read request body:", err)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &post)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if post.Title == "" || post.Text == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = posts.UpdatePost(config.DB, config.Ctx, postId, post)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if post.Tags == nil {
		err = tags.DeleteTags(config.DB, config.Ctx, postId)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
	} else {
		for _, tag := range post.Tags {
			if tag.Name == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		}
		tagMap := make(map[tags.Tag]bool)
		tagList := make([]tags.Tag, 0)
		for _, tag := range post.Tags {
			tagMap[tag] = true
		}
		for _, tag := range post.Tags {
			_, ok := tagMap[tag]
			if ok {
				tagList = append(tagList, tag)
				delete(tagMap, tag)
			}
		}
		post.Tags = tagList
		err = tags.UpdateTags(config.DB, config.Ctx, postId, post.Tags)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post successfully updated"))
}

// @Summary Delete a post
// @Tags posts
// @Param id path int true "Post ID"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Token"
// @Failure 403 "No Access To Post"
// @Failure 404 "Post Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id} [delete]
func deleteId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	post, err := posts.GetPost(config.DB, config.Ctx, postId)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err == sql.ErrNoRows {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if post.AuthorId != userId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Println(err)
		return
	}

	err = posts.DeletePost(config.DB, config.Ctx, postId)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post successfully deleted"))
}

// @Summary Like a post
// @Tags posts
// @Param id path int true "Post ID"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Header"
// @Failure 404 "Post Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id}/like [post]
func like(w http.ResponseWriter, r *http.Request) {
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

	err = likes.AddLike(config.DB, config.Ctx, userId, postId, "like")
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post successfully liked"))
}

// @Summary Dislike a post
// @Tags posts
// @Param id path int true "Post ID"
// @Param Authorization header string true "Auth Token"
// @Success 200
// @Failure 400 "Bad Request"
// @Failure 401 "Invalid Auth Header"
// @Failure 404 "Post Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id}/dislike [post]
func dislike(w http.ResponseWriter, r *http.Request) {
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

	err = likes.AddLike(config.DB, config.Ctx, userId, postId, "dislike")
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post successfully disliked"))
}

// @Summary Get likes for the post
// @Tags posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} int
// @Failure 400 "Bad Request"
// @Failure 404 "Post Not Found"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id}/likes [get]
func postLikes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	likes, err := likes.GetLikes(config.DB, config.Ctx, postId)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	data, err := json.Marshal(likes)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Search posts by title
// @Tags posts
// @Param query path string true "Query"
// @Success 200 {object} []Post
// @Failure 400 "Bad Request"
// @Failure 404 "Post Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/search/q/{query} [get]
func search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var err error
	query := r.PathValue("query")
	if query == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	query, err = url.QueryUnescape(query)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	postlist, err := posts.FilterQuery(config.DB, config.Ctx, query)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if postlist == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(postlist)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Get comments for the post
// @Tags comments
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} []comments.Comment
// @Failure 400 "Bad Request"
// @Failure 404 "Comments Not Found"
// @Failure 405 "Method Not Allowed"
// @Failure 500 "Internal Error"
// @Router /api/posts/{id}/comments [get]
func postComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

// @Summary Get all tags for the post
// @Tags tags
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} []tags.Tag
// @Failure 400 "Bad Request"
// @Failure 404 "Post Not Found"
// @Failure 500 "Internal Error"
// @Failure 405 "Method Not Allowed"
// @Router /api/posts/{id}/tags [get]
func postTags(w http.ResponseWriter, r *http.Request) {
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

	tagList, err := tags.GetTags(config.DB, config.Ctx, postId)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if tagList == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(tagList)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @Summary Get posts associated with the tagPosts
// @Tags tags
// @Produce json
// @Param name path string true "Tag Name"
// @Success 200 {object} posts.Posts
// @Failure 400 "Bad Request"
// @Failure 404 "Tag Not Found"
// @Failure 500 "Internal Error"
// @Router /api/posts/tagPosts/t/{name} [get]
func tagPosts(w http.ResponseWriter, r *http.Request) {
	var err error
	tagName := r.PathValue("name")
	if tagName == "" {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}
	tagName, err = url.QueryUnescape(tagName)
	if err != nil {
		http.Error(w, "Invalid URL Format", http.StatusBadRequest)
		return
	}

	postData, err := posts.FilterTag(config.DB, config.Ctx, tags.Tag{Name: tagName})
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if postData == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(postData)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to marshal JSON:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", add)
	mux.HandleFunc("GET /", get)
	mux.HandleFunc("GET /{id}", getId)
	mux.HandleFunc("PUT /{id}", update)
	mux.HandleFunc("DELETE /{id}", deleteId)
	mux.HandleFunc("POST /{id}/like", like)
	mux.HandleFunc("POST /{id}/dislike", dislike)
	mux.HandleFunc("GET /{id}/likes", postLikes)
	mux.HandleFunc("GET /{id}/comments", postComments)
	mux.HandleFunc("GET /{id}/tags", postTags)
	mux.HandleFunc("GET /tags/t/{name}", tagPosts)
	mux.HandleFunc("GET /search/q/{query}", search)
	return mux
}
