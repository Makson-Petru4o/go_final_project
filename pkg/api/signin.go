package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ожидается один JSON-объект"})
		return
	}
	password := os.Getenv("TODO_PASSWORD")
	want := sha256.Sum256([]byte(password))
	got := sha256.Sum256([]byte(input.Password))
	if subtle.ConstantTimeCompare(want[:], got[:]) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Неверный пароль"})
		return
	}
	now := time.Now()
	key := signingKey(password)
	claims := tokenClaims{
		PasswordHash: passwordChecksum(password, key),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось создать токен"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
