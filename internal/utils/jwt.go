package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey = []byte(os.Getenv("SECRET_KEY"))  // 64 байта (512 бит) для алгоритма HMAC-SHA512

func GenerateAccessToken(userID string, ipAddress string) (string, error) {
    //задаём payload
	claims := jwt.MapClaims{
        "user_id": userID,
        "ip":      ipAddress,
        "exp":     time.Now().Add(time.Minute * 15).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) //объект токен, включает в себя header и payload
    tokenString, err := token.SignedString(secretKey) //генерируется JWT токен
    if err != nil {
        return "", err
    }
    return tokenString, nil
}

// Валидация Access токена
func ValidateAccessToken(accessToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) { //сверяем access токен с header+payload+secretkey
			// Проверяем, что используется алгоритм HS512
			if token.Method.Alg() != jwt.SigningMethodHS512.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

	if err != nil {
		return nil, err
	}

	return token, nil
}