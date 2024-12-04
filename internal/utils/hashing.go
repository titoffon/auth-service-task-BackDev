package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// Генерация bcrypt хеша для refresh токена
func HashRefreshToken(refreshToken string) (string, error) {
    hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedToken), nil
}

// Проверка bcrypt хеша для refresh токена
func CheckRefreshToken(hash string, refreshToken string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(refreshToken))
}
