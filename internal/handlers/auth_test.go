package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/titoffon/auth-service-task-BackDev/internal/utils"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) SaveRefreshToken(ctx context.Context, userID string, refreshTokenHash string, ipAddress string) error {
	args := m.Called(ctx, userID, refreshTokenHash, ipAddress)
	return args.Error(0)
}

func (m *MockDB) GetRefreshTokenHash(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockDB) UpdateRefreshToken(ctx context.Context, userID string, newRefreshTokenHash string) error {
	args := m.Called(ctx, userID, newRefreshTokenHash)
	return args.Error(0)
}

func TestGenerateTokensHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDB := new(MockDB)
	mockDB.On("SaveRefreshToken", mock.Anything, "test_user_id", mock.Anything, "127.0.0.1").Return(nil)

	authHandler := NewAuthHandler(mockDB)
	router.POST("/auth/generate-tokens", authHandler.GenerateTokens)

	req, _ := http.NewRequest("POST", "/auth/generate-tokens?user_id=test_user_id", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])

	// Проверяем вызов SaveRefreshToken
	mockDB.AssertCalled(t, "SaveRefreshToken", mock.Anything, "test_user_id", mock.Anything, "127.0.0.1")
}

func TestGenerateTokensHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDB := new(MockDB)
	mockDB.On("SaveRefreshToken", mock.Anything, "test_user_id", mock.Anything, "127.0.0.1").Return(assert.AnError)

	authHandler := NewAuthHandler(mockDB)
	router.POST("/auth/generate-tokens", authHandler.GenerateTokens)

	req, _ := http.NewRequest("POST", "/auth/generate-tokens?user_id=test_user_id", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func generateTestAccessToken(userID, ip string) string {
    os.Setenv("SECRET_KEY", "test_secret_key") // Убедитесь, что SECRET_KEY установлен
    token, _ := utils.GenerateAccessToken(userID, ip)
    return token
}

func TestRefreshTokensHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()

    mockDB := new(MockDB)

    // Генерация валидных токенов
    accessToken := generateTestAccessToken("test_user_id", "127.0.0.1")
    refreshToken := "test_refresh_token"
    refreshTokenHash, _ := utils.HashRefreshToken(refreshToken)

    // Настройка моков
    mockDB.On("GetRefreshTokenHash", mock.Anything, "test_user_id").Return(refreshTokenHash, nil)
    mockDB.On("UpdateRefreshToken", mock.Anything, "test_user_id", mock.Anything).Return(nil)

    authHandler := NewAuthHandler(mockDB)
    router.POST("/auth/refresh-tokens", authHandler.RefreshTokens)

    requestBody, _ := json.Marshal(map[string]string{
        "access_token":  accessToken,
        "refresh_token": base64.StdEncoding.EncodeToString([]byte(refreshToken)), // base64 кодировка
    })
    req, _ := http.NewRequest("POST", "/auth/refresh-tokens", bytes.NewBuffer(requestBody))
    req.Header.Set("Content-Type", "application/json")
    req.RemoteAddr = "127.0.0.1:12345"
    resp := httptest.NewRecorder()

    router.ServeHTTP(resp, req)

    // Проверки
    assert.Equal(t, http.StatusOK, resp.Code)

    var response map[string]string
    err := json.Unmarshal(resp.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.NotEmpty(t, response["access_token"])
    assert.NotEmpty(t, response["refresh_token"])

    // Проверяем вызовы методов базы данных
    mockDB.AssertCalled(t, "GetRefreshTokenHash", mock.Anything, "test_user_id")
    mockDB.AssertCalled(t, "UpdateRefreshToken", mock.Anything, "test_user_id", mock.Anything)
}


func TestRefreshTokensHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDB := new(MockDB)
	mockDB.On("GetRefreshTokenHash", mock.Anything, "test_user_id").Return("", assert.AnError)

	authHandler := NewAuthHandler(mockDB)
	router.POST("/auth/refresh-tokens", authHandler.RefreshTokens)

	requestBody, _ := json.Marshal(map[string]string{
		"access_token":  "test_access_token",
		"refresh_token": "dGVzdF9yZWZyZXNoX3Rva2Vu", // base64 от "test_refresh_token"
	})
	req, _ := http.NewRequest("POST", "/auth/refresh-tokens", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
