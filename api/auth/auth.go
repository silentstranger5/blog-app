package auth

import (
	"blog/config"
	"blog/db/auth"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

// @Summary Register a new user
// @Tags auth
// @Accept json
// @Param user body User true "User"
// @Success 200
// @Router /api/auth/Register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read request body:", err)
		return
	}
	defer r.Body.Close()

	var user auth.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if user.Username == "" || user.Password == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	h := sha256.New()
	h.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(h.Sum(nil))

	databaseUser, err := auth.GetUser(config.DB, config.Ctx, user.Username)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if databaseUser.Id != 0 {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	err = auth.AddUser(config.DB, config.Ctx, user)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user successfully registered"))
}

// @Summary Get auth Token for the user
// @Tags auth
// @Accept json
// @Produce json
// @Param user body User true "User"
// @Success 200 {object} string
// @Router /api/auth/Token [get]
func Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to read request body:", err)
		return
	}
	defer r.Body.Close()

	var user auth.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if user.Username == "" || user.Password == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	dbUser, err := auth.GetUser(config.DB, config.Ctx, user.Username)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if dbUser.Id == 0 {
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}

	h := sha256.New()
	h.Write([]byte(user.Password))
	data := h.Sum(nil)
	passwordHash := hex.EncodeToString(data)

	if passwordHash != dbUser.Password {
		http.Error(w, "Invalid Password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": dbUser.Id,
	})
	tokenString, err := token.SignedString(config.Secret)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		log.Println("failed to sign token:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(tokenString))
}

func ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", Register)
	mux.HandleFunc("/token", Token)
	return mux
}
