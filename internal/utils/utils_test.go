package utils

import (
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestHashRefreshToken(t *testing.T) {
	token := "test_refresh_token"
	hash, err := HashRefreshToken(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = CheckRefreshToken(hash, token)
	assert.NoError(t, err)

	err = CheckRefreshToken(hash, "invalid_token")
	assert.Error(t, err)
}

func TestGenerateAccessToken(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	ip := "127.0.0.1"

	token, err := GenerateAccessToken(userID, ip)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, parsedToken)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["user_id"])
	assert.Equal(t, ip, claims["ip"])
}
