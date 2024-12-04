package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/titoffon/auth-service-task-BackDev/internal/db"
	"github.com/titoffon/auth-service-task-BackDev/internal/utils"
)

type AuthHandler struct {
	DB db.Database
}

func NewAuthHandler(database db.Database) *AuthHandler { //конструктор
	return &AuthHandler{DB: database}
}

// Генерация токенов
func (h *AuthHandler) GenerateTokens(c *gin.Context) {
    userID := c.DefaultQuery("user_id", "") //UUID(GUID) записывается в качествте значения user_id
    if userID == "" {
        log.Printf("отсутствует параметр user_id в запросе от IP %s", c.ClientIP())
        c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
        return
    }

    ipAddress := c.ClientIP()

    // Генерация Access токена
    accessToken, err := utils.GenerateAccessToken(userID, ipAddress)
    if err != nil {
        log.Printf("ошибка при генерации Access токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
        return
    }

    // Генерация Refresh токена
    refreshToken, err := GenerateSecureRandomToken(32)
    if err != nil {
        log.Printf("Ошибка при генерации Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash refresh token"})
        return
    }
    refreshTokenHash, err := utils.HashRefreshToken(refreshToken) //соль включена
    if err != nil {
        log.Printf("Ошибка при хэшировании Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash refresh token"})
        return
    }

    // Сохранение Refresh токена в базе данных
    err = h.DB.SaveRefreshToken(c.Request.Context(), userID, refreshTokenHash, ipAddress)
    if err != nil {
        log.Printf("Ошибка при сохранении Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save refresh token"})
        return
    }

    /*fmt.Println("refreshToken: ", refreshToken)
    refreshTokenEncode := base64.StdEncoding.EncodeToString([]byte(refreshToken))
    refreshTokenDecode, err := base64.StdEncoding.DecodeString(refreshTokenEncode)
    refreshHash, err := utils.HashRefreshToken(string(refreshTokenDecode))
    fmt.Println("refreshToken to base64: ", refreshTokenEncode)
    fmt.Println("refreshToken from base64: ", string(refreshTokenDecode), err, string(refreshTokenDecode)==refreshToken)
    fmt.Println("Hash refresh: ", refreshHash)

    fmt.Println("Compare: ", refreshTokenHash, refreshHash, refreshHash==refreshTokenHash)
    fmt.Println("Compare HASHED: ", utils.CheckRefreshToken(refreshTokenHash,  string(refreshTokenDecode)))*/

    // Возвращаем пару токенов
    c.JSON(http.StatusOK, gin.H{
        "access_token":  accessToken,
        "refresh_token": base64.StdEncoding.EncodeToString([]byte(refreshToken)),
    })
}

// Refresh токен
func (h *AuthHandler) RefreshTokens(c *gin.Context) {
    var request struct {
        AccessToken  string `json:"access_token" binding:"required"`
        RefreshToken string `json:"refresh_token" binding:"required"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        log.Printf("Ошибка при парсинге JSON-запроса: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Access token and Refresh token are required"})
        return
    }

    // Валидация Access токена
    token, err := utils.ValidateAccessToken(request.AccessToken)
    if err != nil {
        log.Printf("Недействительный Access токен: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
        return
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        log.Printf("Не удалось привести claims токена к MapClaims")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    userID := claims["user_id"].(string)
    ipAddress := claims["ip"].(string)

    // Проверка Refresh токена
    refreshTokenBytes, err := base64.StdEncoding.DecodeString(request.RefreshToken)
    if err != nil {
        log.Printf("Ошибка декодирования Refresh токена: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
        return
    }

    refreshTokenStr := string(refreshTokenBytes)
    savedHash, err := h.DB.GetRefreshTokenHash(c.Request.Context(), userID)
    if err != nil {
        log.Printf("Ошибка при получении хэша Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
        return
    }

    if utils.CheckRefreshToken(savedHash, refreshTokenStr) != nil {
        log.Printf("Недействительный Refresh токен для пользователя %s", userID)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
        return
    }


    // Проверка изменения IP адреса
    if ipAddress != c.ClientIP() {
        // Здесь нужно отправить email, в реальной ситуации
        fmt.Printf("IP address changed: %s -> %s\n", ipAddress, c.ClientIP())
    }

    // Генерация нового Access токена
    newAccessToken, err := utils.GenerateAccessToken(userID, c.ClientIP()) //генерируем новый access с новым ip
    if err != nil {
        log.Printf("Ошибка при генерации нового Access токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate new access token"})
        return
    }

    // Генерация нового Refresh токена
    newRefreshToken, err := GenerateSecureRandomToken(32)
    if err != nil {
        log.Printf("Ошибка при генерации нового Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate new refresh token"})
        return
    }
    newRefreshTokenHash, err := utils.HashRefreshToken(newRefreshToken)
    if err != nil {
        log.Printf("Ошибка при хэшировании нового Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash new refresh token"})
        return
    }

    // Обновление Refresh токена в базе данных
    err = h.DB.UpdateRefreshToken(c.Request.Context(), userID, newRefreshTokenHash)
    if err != nil {
        log.Printf("Ошибка при обновлении Refresh токена для пользователя %s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update refresh token"})
        return
    }

    // Возврат новых токенов клиенту
    c.JSON(http.StatusOK, gin.H{
        "access_token":  newAccessToken,
        "refresh_token": newRefreshToken,
    })
}

func GenerateSecureRandomToken(length int) (string, error) {
    secretToken := make([]byte, length)
    _, err := rand.Read(secretToken)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(secretToken), nil
}
