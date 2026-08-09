package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gin/database"
	"gin/repository"
	"gin/utils"
)

const mobile = "09121234567"
const passord = "123456"

func setupTestHandler(t *testing.T) (*AuthHandler, repository.UserRepository) {
	t.Helper()

	db := database.InitDB()
	repo := repository.NewUserRepository(db)
	handler := NewAuthHandler(repo)

	return handler, repo
}

func performLogin(t *testing.T, handler *AuthHandler) string {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := loginRequest{
		Mobile:   mobile,
		Password: passord,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var response struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if response.Token == "" {
		t.Fatal("expected a token in response, got empty string")
	}

	return response.Token
}

func TestLoginSuccess(t *testing.T) {
	handler, _ := setupTestHandler(t)
	token := performLogin(t, handler)
	t.Logf("token: %s", token)
}

func TestLoginTokenValid(t *testing.T) {
	handler, repo := setupTestHandler(t)
	token := performLogin(t, handler)

	userId, err := utils.ParseToken(token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	user, err := repo.FindByID(userId)
	if err != nil {
		t.Fatalf("failed to get user by id %d: %v", userId, err)
	}

	if user.Mobile != mobile {
		t.Fatalf("expected mobile %q, got %q", mobile, user.Mobile)
	}

	t.Logf("mobile: %s", user.Mobile)
}

func TestLoginInvalidCredentials(t *testing.T) {
	handler, _ := setupTestHandler(t)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := loginRequest{
		Mobile:   mobile,
		Password: "wrong-password",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body: %s", w.Code, w.Body.String())
	}
}
