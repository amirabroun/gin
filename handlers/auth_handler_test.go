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

const contact = "abroun234@gmail.com"
const passord = "123456"

func TestLogin(t *testing.T) {
	db := database.InitDB()
	
	repo := repository.NewUserRepository(db)
	handler := NewAuthHandler(repo)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := loginRequest{
		Contact:  contact,
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

	userId, err := utils.ParseToken(response.Token)

	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	user, err := repo.FindByID(userId)

	if err != nil {
		t.Fatalf("failed to get user")
	}

	if user.Contact != contact {
		t.Fatalf("failed to get user by id %d: %v", userId, err)
	}

	t.Logf("token: %s", response.Token)
	t.Logf("contact: %s", user.Contact)
}
