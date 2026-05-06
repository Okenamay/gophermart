package auth

import (
	"fmt"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// claims определяет структуру JWT claims
type claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildJWTString создает новый JWT токен для указанного ID пользователя
func BuildJWTString(conf *config.Cfg, userID int) (string, error) {
	tokenExp := time.Duration(conf.TokenExpiry) * time.Hour

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(conf.AuthorizationKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserIDFromToken парсит строку токена и возвращает ID пользователя
func GetUserIDFromToken(conf *config.Cfg, tokenString string) (int, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(conf.AuthorizationKey), nil
		})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	return claims.UserID, nil
}

// HashPassword генерирует bcrypt-хеш для пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash сравнивает пароль и его хеш
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
