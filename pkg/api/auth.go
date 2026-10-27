package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type tokenClaims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func signingKey(password string) []byte {
	if secret := os.Getenv("TODO_JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	key := sha256.Sum256([]byte("go-todo:jwt:" + password))
	return key[:]
}

func passwordChecksum(password string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("go-todo:password:" + password))
	return hex.EncodeToString(mac.Sum(nil))
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		password := os.Getenv("TODO_PASSWORD")
		if password == "" {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("token")
		if err == nil && cookie.Value != "" {
			key := signingKey(password)
			claims := &tokenClaims{}
			token, parseErr := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (any, error) {
				return key, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
			if parseErr == nil && token.Valid && subtle.ConstantTimeCompare([]byte(claims.PasswordHash), []byte(passwordChecksum(password, key))) == 1 {
				next(w, r)
				return
			}
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}
}
