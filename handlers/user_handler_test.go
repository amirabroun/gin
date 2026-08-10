package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gin/app"
	"gin/handlers"
)

func TestGetAuthUserPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := setupRouter(t)

	w := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/posts", nil)

	request.Header.Add("Authorization", "Bearer "+getToken(t, router))

	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	t.Logf("status: %d", w.Code)
	t.Logf("body: %s", w.Body.String())
}

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()

	app := app.NewApp()

	return app.Router
}

func getToken(t *testing.T, router *gin.Engine) string {
	t.Helper()

	body := handlers.LoginRequest{Mobile: "09121234567", Password: "123456"}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	w := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, request)

	var response struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	if response.Token == "" {
		t.Fatalf("login did not return a token, status: %d, body: %s", w.Code, w.Body.String())
	}

	return response.Token
}
