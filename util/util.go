package util

import (
	"blog/config"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
)

func ParseUrlId(path string) int {
	parts := strings.Split(path, "/")
	if !(len(parts) == 3) {
		return 0
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}
	return id
}

func ParseToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil,
				fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.Secret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return 0, fmt.Errorf("failed to obtain claims: %v", err)
	}

	val, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("failed to obtain user_id: %v", err)
	}
	return int(val), nil
}

func ParseAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing auth header")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid auth header format")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func ParseAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("Token")
	if err != nil && err != http.ErrNoCookie {
		return "", fmt.Errorf("failed to read cookie: %v", err)
	}
	if err == http.ErrNoCookie {
		return "", err
	}
	return cookie.Value, nil
}

func Template(files []string, funcmap template.FuncMap, wr io.Writer, data any) error {
	tmpl, err := template.New(path.Base(files[0])).Funcs(funcmap).ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}
	err = tmpl.Execute(wr, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	return nil
}

func Request(method, url, token string, rbody io.Reader) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, rbody)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("failed to make request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("failed to read response body: %v", err)
	}

	return body, res.StatusCode, nil
}
