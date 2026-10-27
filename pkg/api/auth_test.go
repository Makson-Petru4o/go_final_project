package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthentication(t *testing.T) {
	t.Setenv("TODO_PASSWORD", "unit-password")
	t.Setenv("TODO_JWT_SECRET", "unit-only-secret")
	key := signingKey("unit-password")
	valid := tokenClaims{PasswordHash: passwordChecksum("unit-password", key), RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	sign := func(method jwt.SigningMethod, claims tokenClaims, signingKey []byte) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, claims).SignedString(signingKey)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	expired := valid
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	missingExpiry := valid
	missingExpiry.ExpiresAt = nil
	wrongChecksum := valid
	wrongChecksum.PasswordHash = "wrong"
	cases := []struct {
		name, token string
		want        int
	}{
		{"missing", "", 401},
		{"malformed", "invalid", 401},
		{"valid", sign(jwt.SigningMethodHS256, valid, key), 204},
		{"wrong signature", sign(jwt.SigningMethodHS256, valid, []byte("other")), 401},
		{"expired", sign(jwt.SigningMethodHS256, expired, key), 401},
		{"missing exp", sign(jwt.SigningMethodHS256, missingExpiry, key), 401},
		{"wrong algorithm", sign(jwt.SigningMethodHS384, valid, key), 401},
		{"wrong checksum", sign(jwt.SigningMethodHS256, wrongChecksum, key), 401},
	}
	handler := auth(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/tasks", nil)
			if tc.token != "" {
				r.AddCookie(&http.Cookie{Name: "token", Value: tc.token})
			}
			w := httptest.NewRecorder()
			handler(w, r)
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
	token := sign(jwt.SigningMethodHS256, valid, key)
	t.Setenv("TODO_PASSWORD", "changed-password")
	r := httptest.NewRequest("GET", "/api/tasks", nil)
	r.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	handler(w, r)
	if w.Code != 401 {
		t.Fatal("password change must invalidate token even with separate secret")
	}
	t.Setenv("TODO_PASSWORD", "")
	w = httptest.NewRecorder()
	handler(w, httptest.NewRequest("GET", "/api/tasks", nil))
	if w.Code != 204 {
		t.Fatal("empty password must disable auth")
	}
}

func TestSignin(t *testing.T) {
	t.Setenv("TODO_PASSWORD", "unit-password")
	t.Setenv("TODO_JWT_SECRET", "")
	for _, tc := range []struct {
		method, body string
		status       int
	}{
		{"GET", "", 405}, {"POST", "{", 400}, {"POST", `{"password":"wrong"}`, 401}, {"POST", `{"password":"unit-password"} {}`, 400}, {"POST", `{"password":"unit-password"}`, 200},
	} {
		w := httptest.NewRecorder()
		signinHandler(w, httptest.NewRequest(tc.method, "/api/signin", strings.NewReader(tc.body)))
		if w.Code != tc.status {
			t.Fatalf("got %d, want %d", w.Code, tc.status)
		}
		var result map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if tc.status != 200 {
			if result["token"] != "" {
				t.Fatal("token returned on error")
			}
			continue
		}
		claims := &tokenClaims{}
		token, err := jwt.ParseWithClaims(result["token"], claims, func(*jwt.Token) (any, error) { return signingKey("unit-password"), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
		if err != nil || !token.Valid {
			t.Fatalf("invalid issued token: %v", err)
		}
		if claims.ExpiresAt.Sub(claims.IssuedAt.Time) != 8*time.Hour {
			t.Fatal("unexpected lifetime")
		}
		t.Setenv("TODO_PASSWORD", "changed")
		r := httptest.NewRequest("GET", "/api/tasks", nil)
		r.AddCookie(&http.Cookie{Name: "token", Value: result["token"]})
		check := httptest.NewRecorder()
		auth(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })(check, r)
		if check.Code != 401 {
			t.Fatal("fallback key must change with password")
		}
	}
}
