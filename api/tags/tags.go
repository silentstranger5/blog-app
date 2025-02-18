package tags

import (
	"blog/config"
	"blog/db/posts"
	"blog/db/tags"
	"blog/util"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// @Summary Get all tags for the post
// @Tags tags
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} []Tag
// @Router /api/tags/get/{id} [get]
func GetTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	postId := util.ParseUrlId(r.URL.Path)
	if postId == 0 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
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

// @Summary Get posts associated with the tag
// @Tags tags
// @Produce json
// @Param name path string true "Tag Name"
// @Success 200 {object} posts.Posts
// @Router /api/tags/post/{name} [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if !(len(parts) == 3 && parts[2] != "") {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	tagName, err := url.QueryUnescape(parts[2])
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to escape string:", err)
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
	mux.HandleFunc("/get/", GetTags)
	mux.HandleFunc("/posts/", GetPosts)
	return mux
}
